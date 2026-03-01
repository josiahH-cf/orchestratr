package gui

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/josiahH-cf/orchestratr/internal/registry"
)

// testServer returns a *Server backed by cfgPath and its Handler.
func testServer(t *testing.T, cfgPath string, apiPort int) (*Server, http.Handler) {
	t.Helper()
	logger := log.New(io.Discard, "", 0)
	s := NewServer(cfgPath, apiPort, logger)
	return s, s.Handler()
}

// seedConfig writes a valid YAML config file.
func seedConfig(t *testing.T, dir string) string {
	t.Helper()
	cfg := filepath.Join(dir, "config.yml")
	data := `leader_key: ctrl+space
chord_timeout_ms: 2000
api_port: 9876
apps:
  - name: firefox
    chord: f
    command: firefox
    environment: native
  - name: terminal
    chord: t
    command: alacritty
    environment: native
`
	if err := os.WriteFile(cfg, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	return cfg
}

func TestGetConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := seedConfig(t, dir)
	_, h := testServer(t, cfgPath, 0)

	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/config status = %d, want 200", rec.Code)
	}

	var cfg registry.Config
	if err := json.NewDecoder(rec.Body).Decode(&cfg); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if cfg.LeaderKey != "ctrl+space" {
		t.Errorf("leader_key = %q, want %q", cfg.LeaderKey, "ctrl+space")
	}
	if len(cfg.Apps) != 2 {
		t.Errorf("apps count = %d, want 2", len(cfg.Apps))
	}
	if cfg.Apps[0].Name != "firefox" {
		t.Errorf("apps[0].name = %q, want %q", cfg.Apps[0].Name, "firefox")
	}
}

func TestGetConfig_MissingFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "nonexistent.yml")
	_, h := testServer(t, cfgPath, 0)

	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	// Should return default config (200), not an error.
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/config (missing) status = %d, want 200", rec.Code)
	}

	var cfg registry.Config
	if err := json.NewDecoder(rec.Body).Decode(&cfg); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(cfg.Apps) != 0 {
		t.Errorf("apps count = %d, want 0 (default)", len(cfg.Apps))
	}
}

func TestPutConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yml")
	// Start with empty file so the PUT creates it.
	os.WriteFile(cfgPath, []byte("leader_key: x\nchord_timeout_ms: 1000\napi_port: 9876\napps: []\n"), 0o644)
	_, h := testServer(t, cfgPath, 0) // no daemon

	body := `{
		"leader_key": "ctrl+k",
		"chord_timeout_ms": 3000,
		"api_port": 9877,
		"apps": [
			{"name": "code", "chord": "c", "command": "code", "environment": "native"}
		]
	}`

	req := httptest.NewRequest(http.MethodPut, "/api/config", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("PUT /api/config status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["status"] != "ok" {
		t.Errorf("response status = %v, want ok", resp["status"])
	}
	if resp["reloaded"] != false {
		t.Errorf("reloaded = %v, want false (no daemon)", resp["reloaded"])
	}
	if apps, ok := resp["apps"].(float64); !ok || apps != 1 {
		t.Errorf("apps = %v, want 1", resp["apps"])
	}

	// Verify the file was written.
	saved, err := registry.LoadAndValidate(cfgPath)
	if err != nil {
		t.Fatalf("reload saved config: %v", err)
	}
	if saved.LeaderKey != "ctrl+k" {
		t.Errorf("saved leader_key = %q, want %q", saved.LeaderKey, "ctrl+k")
	}
	if len(saved.Apps) != 1 {
		t.Fatalf("saved apps count = %d, want 1", len(saved.Apps))
	}
	if saved.Apps[0].Name != "code" {
		t.Errorf("saved apps[0].name = %q, want %q", saved.Apps[0].Name, "code")
	}
}

func TestPutConfig_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	cfgPath := seedConfig(t, dir)
	_, h := testServer(t, cfgPath, 0)

	req := httptest.NewRequest(http.MethodPut, "/api/config", strings.NewReader("{not json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("PUT bad JSON status = %d, want 400", rec.Code)
	}
}

func TestPutConfig_ValidationError(t *testing.T) {
	dir := t.TempDir()
	cfgPath := seedConfig(t, dir)
	_, h := testServer(t, cfgPath, 0)

	// Duplicate chords should fail validation.
	body := `{
		"leader_key": "ctrl+space",
		"chord_timeout_ms": 2000,
		"api_port": 9876,
		"apps": [
			{"name": "a", "chord": "f", "command": "a", "environment": "native"},
			{"name": "b", "chord": "f", "command": "b", "environment": "native"}
		]
	}`

	req := httptest.NewRequest(http.MethodPut, "/api/config", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("PUT duplicate chords status = %d, want 400, body = %s", rec.Code, rec.Body.String())
	}
}

func TestConfigMethodNotAllowed(t *testing.T) {
	dir := t.TempDir()
	cfgPath := seedConfig(t, dir)
	_, h := testServer(t, cfgPath, 0)

	req := httptest.NewRequest(http.MethodDelete, "/api/config", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("DELETE /api/config status = %d, want 405", rec.Code)
	}
}

func TestDaemonInfo_NoDaemon(t *testing.T) {
	dir := t.TempDir()
	cfgPath := seedConfig(t, dir)
	_, h := testServer(t, cfgPath, 0) // port 0 = no daemon

	req := httptest.NewRequest(http.MethodGet, "/api/daemon-info", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/daemon-info status = %d", rec.Code)
	}

	var info map[string]any
	json.NewDecoder(rec.Body).Decode(&info)
	if info["connected"] != false {
		t.Errorf("connected = %v, want false", info["connected"])
	}
	if info["api_port"] != float64(0) {
		t.Errorf("api_port = %v, want 0", info["api_port"])
	}
}

func TestDaemonInfo_MethodNotAllowed(t *testing.T) {
	dir := t.TempDir()
	cfgPath := seedConfig(t, dir)
	_, h := testServer(t, cfgPath, 0)

	req := httptest.NewRequest(http.MethodPost, "/api/daemon-info", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST /api/daemon-info status = %d, want 405", rec.Code)
	}
}

func TestStaticFileServing(t *testing.T) {
	dir := t.TempDir()
	cfgPath := seedConfig(t, dir)
	_, h := testServer(t, cfgPath, 0)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET / status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "orchestratr") {
		t.Error("index.html should contain 'orchestratr'")
	}
	if !strings.Contains(body, "/api/config") {
		t.Error("index.html should reference /api/config")
	}
}

func TestServerStartStop(t *testing.T) {
	dir := t.TempDir()
	cfgPath := seedConfig(t, dir)
	logger := log.New(io.Discard, "", 0)
	s := NewServer(cfgPath, 0, logger)

	if err := s.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}

	port := s.Port()
	if port == 0 {
		t.Fatal("Port() = 0 after Start()")
	}

	url := s.URL()
	if !strings.Contains(url, "127.0.0.1") {
		t.Errorf("URL = %q, want localhost", url)
	}

	// Starting again should error.
	if err := s.Start(); err == nil {
		t.Error("second Start() should return error")
	}

	// Verify the server responds.
	resp, err := http.Get(url + "/api/config")
	if err != nil {
		t.Fatalf("GET on running server: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /api/config status = %d", resp.StatusCode)
	}

	if err := s.Stop(); err != nil {
		t.Fatalf("Stop: %v", err)
	}

	// Stop is idempotent.
	if err := s.Stop(); err != nil {
		t.Fatalf("second Stop: %v", err)
	}
}

func TestPutConfig_TriggersReload(t *testing.T) {
	// Start a fake daemon that records whether /reload was called.
	reloaded := false
	daemon := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/reload" && r.Method == http.MethodPost {
			reloaded = true
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer daemon.Close()

	// Parse the daemon port from its URL.
	parts := strings.Split(daemon.URL, ":")
	daemonPort := 0
	fmt.Sscanf(parts[len(parts)-1], "%d", &daemonPort)

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yml")
	os.WriteFile(cfgPath, []byte("leader_key: x\nchord_timeout_ms: 1000\napi_port: 9876\napps: []\n"), 0o644)
	_, h := testServer(t, cfgPath, daemonPort)

	body := `{
		"leader_key": "ctrl+space",
		"chord_timeout_ms": 2000,
		"api_port": 9876,
		"apps": [
			{"name": "app1", "chord": "a", "command": "app1", "environment": "native"}
		]
	}`

	req := httptest.NewRequest(http.MethodPut, "/api/config", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("PUT status = %d, body = %s", rec.Code, rec.Body.String())
	}

	if !reloaded {
		t.Error("expected daemon /reload to be called")
	}

	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["reloaded"] != true {
		t.Errorf("reloaded = %v, want true", resp["reloaded"])
	}
}
