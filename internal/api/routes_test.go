package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/josiahH-cf/orchestratr/internal/registry"
)

func TestGetApps_EmptyRegistry(t *testing.T) {
	reg := registry.NewRegistry(registry.Config{})
	s := NewServer(0, "v0.0.1", reg, nil)
	handler := s.Handler()

	req := httptest.NewRequest("GET", "/apps", nil)
	req.RemoteAddr = "127.0.0.1:54321"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var apps []registry.AppEntry
	if err := json.NewDecoder(rec.Body).Decode(&apps); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if len(apps) != 0 {
		t.Errorf("got %d apps, want 0", len(apps))
	}
}

func TestGetApps_WithApps(t *testing.T) {
	cfg := registry.Config{
		Apps: []registry.AppEntry{
			{Name: "espansr", Chord: "e", Command: "espansr", Environment: "native", Description: "Espanso manager"},
			{Name: "firefox", Chord: "f", Command: "firefox", Environment: "native"},
		},
	}
	reg := registry.NewRegistry(cfg)
	s := NewServer(0, "v0.0.1", reg, nil)
	handler := s.Handler()

	req := httptest.NewRequest("GET", "/apps", nil)
	req.RemoteAddr = "127.0.0.1:54321"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	// Parse as raw JSON to verify field names are snake_case.
	var raw []map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&raw); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if len(raw) != 2 {
		t.Fatalf("got %d apps, want 2", len(raw))
	}

	// Verify snake_case field names (not PascalCase).
	first := raw[0]
	for _, key := range []string{"name", "chord", "command", "environment"} {
		if _, ok := first[key]; !ok {
			t.Errorf("missing expected snake_case field %q in JSON response", key)
		}
	}
	// Verify PascalCase is NOT present.
	for _, key := range []string{"Name", "Chord", "Command", "Environment"} {
		if _, ok := first[key]; ok {
			t.Errorf("unexpected PascalCase field %q in JSON response — missing json struct tag?", key)
		}
	}

	if first["name"] != "espansr" {
		t.Errorf("apps[0].name = %v, want %q", first["name"], "espansr")
	}
	if raw[1]["name"] != "firefox" {
		t.Errorf("apps[1].name = %v, want %q", raw[1]["name"], "firefox")
	}
}

func TestGetApps_NilRegistry(t *testing.T) {
	s := NewServer(0, "v0.0.1", nil, nil)
	handler := s.Handler()

	req := httptest.NewRequest("GET", "/apps", nil)
	req.RemoteAddr = "127.0.0.1:54321"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var apps []any
	if err := json.NewDecoder(rec.Body).Decode(&apps); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if len(apps) != 0 {
		t.Errorf("got %d apps, want 0", len(apps))
	}
}

func TestGetApps_WrongMethod(t *testing.T) {
	s := NewServer(0, "v0.0.1", nil, nil)
	handler := s.Handler()

	req := httptest.NewRequest("POST", "/apps", nil)
	req.RemoteAddr = "127.0.0.1:54321"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestPostReload_Success(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yml")

	initialYAML := `leader_key: ctrl+space
chord_timeout_ms: 2000
api_port: 9876
apps:
  - name: testapp
    chord: t
    command: echo hello
    environment: native
`
	if err := os.WriteFile(cfgPath, []byte(initialYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	reg := registry.NewRegistry(registry.Config{})
	reloadFn := func() (*registry.Config, error) {
		cfg, err := registry.LoadAndValidate(cfgPath)
		if err != nil {
			return nil, err
		}
		reg.Swap(*cfg)
		return cfg, nil
	}

	s := NewServer(0, "v0.0.1", reg, reloadFn)
	handler := s.Handler()

	req := httptest.NewRequest("POST", "/reload", nil)
	req.RemoteAddr = "127.0.0.1:54321"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if resp["status"] != "ok" {
		t.Errorf("status = %q, want %q", resp["status"], "ok")
	}

	apps, ok := resp["apps"].([]any)
	if !ok {
		t.Fatalf("apps field is %T, want []any", resp["apps"])
	}
	if len(apps) != 1 {
		t.Errorf("got %d apps, want 1", len(apps))
	}
}

func TestPostReload_ValidationError(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yml")

	// Two apps with duplicate chords.
	badYAML := `leader_key: ctrl+space
apps:
  - name: app1
    chord: a
    command: cmd1
    environment: native
  - name: app2
    chord: a
    command: cmd2
    environment: native
`
	if err := os.WriteFile(cfgPath, []byte(badYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	reloadFn := func() (*registry.Config, error) {
		return registry.LoadAndValidate(cfgPath)
	}

	s := NewServer(0, "v0.0.1", nil, reloadFn)
	handler := s.Handler()

	req := httptest.NewRequest("POST", "/reload", nil)
	req.RemoteAddr = "127.0.0.1:54321"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}

	var errResp ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
		t.Fatalf("decoding error response: %v", err)
	}
	if errResp.Code != "reload_failed" {
		t.Errorf("code = %q, want %q", errResp.Code, "reload_failed")
	}
}

func TestPostReload_NotConfigured(t *testing.T) {
	s := NewServer(0, "v0.0.1", nil, nil)
	handler := s.Handler()

	req := httptest.NewRequest("POST", "/reload", nil)
	req.RemoteAddr = "127.0.0.1:54321"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
}

func TestPostReload_WrongMethod(t *testing.T) {
	s := NewServer(0, "v0.0.1", nil, nil)
	handler := s.Handler()

	req := httptest.NewRequest("GET", "/reload", nil)
	req.RemoteAddr = "127.0.0.1:54321"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestPostTrigger_NoChord_NoEngine(t *testing.T) {
	s := NewServer(0, "v0.0.1", nil, nil)
	handler := s.Handler()

	req := httptest.NewRequest("POST", "/trigger", nil)
	req.RemoteAddr = "127.0.0.1:54321"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
}

func TestPostTrigger_NoChord_WithEngine(t *testing.T) {
	s := NewServer(0, "v0.0.1", nil, nil)
	triggered := false
	s.SetTriggerFunc(func() error {
		triggered = true
		return nil
	})
	handler := s.Handler()

	req := httptest.NewRequest("POST", "/trigger", nil)
	req.RemoteAddr = "127.0.0.1:54321"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !triggered {
		t.Error("trigger function was not called")
	}
}

func TestPostTrigger_WithChord_LaunchesApp(t *testing.T) {
	cfg := registry.Config{
		Apps: []registry.AppEntry{
			{Name: "calculator", Chord: "c", Command: "calc", Environment: "native"},
			{Name: "terminal", Chord: "t", Command: "xterm", Environment: "native"},
		},
	}
	reg := registry.NewRegistry(cfg)
	s := NewServer(0, "v0.0.1", reg, nil)

	var launchedApp string
	s.SetLaunchFunc(func(name string) (int, error) {
		launchedApp = name
		return 42, nil
	})
	handler := s.Handler()

	body := strings.NewReader(`{"chord":"c"}`)
	req := httptest.NewRequest("POST", "/trigger", body)
	req.RemoteAddr = "127.0.0.1:54321"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if launchedApp != "calculator" {
		t.Errorf("launched app = %q, want %q", launchedApp, "calculator")
	}

	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if resp["app"] != "calculator" {
		t.Errorf("resp.app = %v, want %q", resp["app"], "calculator")
	}
}

func TestPostTrigger_WithChord_NotFound(t *testing.T) {
	cfg := registry.Config{
		Apps: []registry.AppEntry{
			{Name: "calculator", Chord: "c", Command: "calc", Environment: "native"},
		},
	}
	reg := registry.NewRegistry(cfg)
	s := NewServer(0, "v0.0.1", reg, nil)
	s.SetLaunchFunc(func(name string) (int, error) {
		return 1, nil
	})
	handler := s.Handler()

	body := strings.NewReader(`{"chord":"x"}`)
	req := httptest.NewRequest("POST", "/trigger", body)
	req.RemoteAddr = "127.0.0.1:54321"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d; body: %s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestPostTrigger_WithChord_NoLauncher(t *testing.T) {
	cfg := registry.Config{
		Apps: []registry.AppEntry{
			{Name: "calculator", Chord: "c", Command: "calc", Environment: "native"},
		},
	}
	reg := registry.NewRegistry(cfg)
	s := NewServer(0, "v0.0.1", reg, nil)
	// No launchFn set
	handler := s.Handler()

	body := strings.NewReader(`{"chord":"c"}`)
	req := httptest.NewRequest("POST", "/trigger", body)
	req.RemoteAddr = "127.0.0.1:54321"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
}

func TestPostTrigger_WrongMethod(t *testing.T) {
	s := NewServer(0, "v0.0.1", nil, nil)
	handler := s.Handler()

	req := httptest.NewRequest("GET", "/trigger", nil)
	req.RemoteAddr = "127.0.0.1:54321"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}
