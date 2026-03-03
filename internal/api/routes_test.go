package api

import (
	"encoding/json"
	"fmt"
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

func TestPostTrigger_WithChord_NilRegistry(t *testing.T) {
	// When a chord is supplied but the registry is nil, the handler
	// should return 503 (registry not loaded), not panic.
	s := NewServer(0, "v0.0.1", nil, nil)
	s.SetLaunchFunc(func(name string) (int, error) {
		return 1, nil
	})
	handler := s.Handler()

	body := strings.NewReader(`{"chord":"c"}`)
	req := httptest.NewRequest("POST", "/trigger", body)
	req.RemoteAddr = "127.0.0.1:54321"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d; body: %s", rec.Code, http.StatusServiceUnavailable, rec.Body.String())
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

func TestPostLaunch_FailureSetsErrorState(t *testing.T) {
	cfg := registry.Config{
		Apps: []registry.AppEntry{
			{Name: "myapp", Chord: "m", Command: "myapp --gui", Environment: "native"},
		},
	}
	reg := registry.NewRegistry(cfg)
	s := NewServer(0, "v0.0.1", reg, nil)

	s.SetLaunchFunc(func(name string) (int, error) {
		return 0, fmt.Errorf("exec: myapp: not found")
	})
	handler := s.Handler()

	// POST /apps/myapp/launch — expect failure.
	req := httptest.NewRequest("POST", "/apps/myapp/launch", nil)
	req.RemoteAddr = "127.0.0.1:54321"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("launch status = %d, want %d; body: %s", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}

	// GET /apps/myapp/state — error should be populated.
	req2 := httptest.NewRequest("GET", "/apps/myapp/state", nil)
	req2.RemoteAddr = "127.0.0.1:54321"
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("state status = %d, want %d", rec2.Code, http.StatusOK)
	}

	var state AppState
	if err := json.NewDecoder(rec2.Body).Decode(&state); err != nil {
		t.Fatalf("decoding state: %v", err)
	}

	if state.Error == "" {
		t.Error("state.Error should be non-empty after launch failure")
	}
	if !strings.Contains(state.Error, "myapp") {
		t.Errorf("state.Error = %q, want it to contain app info", state.Error)
	}
	if !strings.Contains(state.Error, "not found") {
		t.Errorf("state.Error = %q, want it to contain the error message", state.Error)
	}
	if state.ErrorAt == nil {
		t.Error("state.ErrorAt should be non-nil after launch failure")
	}
}

func TestPostLaunch_SuccessAfterFailureClearsError(t *testing.T) {
	cfg := registry.Config{
		Apps: []registry.AppEntry{
			{Name: "myapp", Chord: "m", Command: "myapp --gui", Environment: "native"},
		},
	}
	reg := registry.NewRegistry(cfg)
	s := NewServer(0, "v0.0.1", reg, nil)

	callCount := 0
	s.SetLaunchFunc(func(name string) (int, error) {
		callCount++
		if callCount == 1 {
			return 0, fmt.Errorf("exec: myapp: not found")
		}
		return 42, nil
	})
	handler := s.Handler()

	// First launch fails.
	req := httptest.NewRequest("POST", "/apps/myapp/launch", nil)
	req.RemoteAddr = "127.0.0.1:54321"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("first launch status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	// Verify error state is set.
	stateAfterFail := s.State().Get("myapp")
	if stateAfterFail == nil || stateAfterFail.Error == "" {
		t.Fatal("error state should be set after failed launch")
	}

	// Second launch succeeds.
	req2 := httptest.NewRequest("POST", "/apps/myapp/launch", nil)
	req2.RemoteAddr = "127.0.0.1:54321"
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("second launch status = %d, want %d; body: %s", rec2.Code, http.StatusOK, rec2.Body.String())
	}

	// GET /apps/myapp/state — error should be cleared.
	req3 := httptest.NewRequest("GET", "/apps/myapp/state", nil)
	req3.RemoteAddr = "127.0.0.1:54321"
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req3)

	var state AppState
	if err := json.NewDecoder(rec3.Body).Decode(&state); err != nil {
		t.Fatalf("decoding state: %v", err)
	}

	if state.Error != "" {
		t.Errorf("state.Error = %q, want empty after successful relaunch", state.Error)
	}
	if state.ErrorAt != nil {
		t.Errorf("state.ErrorAt = %v, want nil after successful relaunch", state.ErrorAt)
	}
}

func TestPostTrigger_WithChord_FailureSetsErrorState(t *testing.T) {
	cfg := registry.Config{
		Apps: []registry.AppEntry{
			{Name: "calculator", Chord: "c", Command: "calc --ui", Environment: "wsl"},
		},
	}
	reg := registry.NewRegistry(cfg)
	s := NewServer(0, "v0.0.1", reg, nil)

	s.SetLaunchFunc(func(name string) (int, error) {
		return 0, fmt.Errorf("wsl.exe: command not found")
	})
	handler := s.Handler()

	// Trigger with chord that fails.
	body := strings.NewReader(`{"chord":"c"}`)
	req := httptest.NewRequest("POST", "/trigger", body)
	req.RemoteAddr = "127.0.0.1:54321"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("trigger status = %d, want %d; body: %s", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}

	// Verify error state includes details.
	state := s.State().Get("calculator")
	if state == nil {
		t.Fatal("state should exist for calculator after failed trigger")
	}
	if state.Error == "" {
		t.Error("state.Error should be non-empty after failed trigger launch")
	}
	if !strings.Contains(state.Error, "calc") {
		t.Errorf("state.Error = %q, want it to contain command info", state.Error)
	}
	if !strings.Contains(state.Error, "wsl") {
		t.Errorf("state.Error = %q, want it to contain environment info", state.Error)
	}
	if state.ErrorAt == nil {
		t.Error("state.ErrorAt should be non-nil after failed trigger launch")
	}
}

func TestPostLaunch_ErrorStateIncludesDetails(t *testing.T) {
	cfg := registry.Config{
		Apps: []registry.AppEntry{
			{Name: "editor", Chord: "v", Command: "vim --server", Environment: "wsl"},
		},
	}
	reg := registry.NewRegistry(cfg)
	s := NewServer(0, "v0.0.1", reg, nil)

	s.SetLaunchFunc(func(name string) (int, error) {
		return 0, fmt.Errorf("exec: vim: not found")
	})
	handler := s.Handler()

	req := httptest.NewRequest("POST", "/apps/editor/launch", nil)
	req.RemoteAddr = "127.0.0.1:54321"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// GET state via JSON.
	req2 := httptest.NewRequest("GET", "/apps/editor/state", nil)
	req2.RemoteAddr = "127.0.0.1:54321"
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	// Parse JSON and verify error and error_at fields.
	var raw map[string]any
	if err := json.NewDecoder(rec2.Body).Decode(&raw); err != nil {
		t.Fatalf("decoding state JSON: %v", err)
	}
	errField, ok := raw["error"]
	if !ok || errField == "" {
		t.Fatal("JSON response should have non-empty 'error' field")
	}
	errStr := errField.(string)
	if !strings.Contains(errStr, "vim --server") {
		t.Errorf("error = %q, want it to contain command attempted", errStr)
	}
	if !strings.Contains(errStr, "wsl") {
		t.Errorf("error = %q, want it to contain environment", errStr)
	}
	if !strings.Contains(errStr, "not found") {
		t.Errorf("error = %q, want it to contain error message", errStr)
	}
	if _, ok := raw["error_at"]; !ok {
		t.Error("JSON response should have 'error_at' field")
	}
}
