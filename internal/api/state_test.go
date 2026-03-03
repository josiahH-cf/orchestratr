package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/josiahH-cf/orchestratr/internal/registry"
)

func TestStateTracker_SetLaunched(t *testing.T) {
	st := NewStateTracker()
	st.SetLaunched("testapp")

	state := st.Get("testapp")
	if state == nil {
		t.Fatal("expected state for testapp, got nil")
	}
	if !state.Launched {
		t.Error("expected Launched = true")
	}
	if state.LaunchedAt == nil {
		t.Error("expected LaunchedAt to be set")
	}
	if state.Ready {
		t.Error("expected Ready = false after launch")
	}
}

func TestStateTracker_SetReady(t *testing.T) {
	st := NewStateTracker()
	st.SetLaunched("testapp")
	st.SetReady("testapp")

	state := st.Get("testapp")
	if state == nil {
		t.Fatal("expected state for testapp, got nil")
	}
	if !state.Launched {
		t.Error("expected Launched = true")
	}
	if !state.Ready {
		t.Error("expected Ready = true")
	}
	if state.ReadyAt == nil {
		t.Error("expected ReadyAt to be set")
	}
}

func TestStateTracker_RelaunchResetsReady(t *testing.T) {
	st := NewStateTracker()
	st.SetLaunched("testapp")
	st.SetReady("testapp")

	// Re-launch should reset ready state.
	st.SetLaunched("testapp")

	state := st.Get("testapp")
	if state == nil {
		t.Fatal("expected state for testapp, got nil")
	}
	if !state.Launched {
		t.Error("expected Launched = true")
	}
	if state.Ready {
		t.Error("expected Ready = false after re-launch")
	}
	if state.ReadyAt != nil {
		t.Error("expected ReadyAt = nil after re-launch")
	}
}

func TestStateTracker_GetUnknown(t *testing.T) {
	st := NewStateTracker()
	if st.Get("nonexistent") != nil {
		t.Error("expected nil for unknown app")
	}
}

func TestStateTracker_All(t *testing.T) {
	st := NewStateTracker()
	st.SetLaunched("app1")
	st.SetLaunched("app2")

	all := st.All()
	if len(all) != 2 {
		t.Errorf("got %d states, want 2", len(all))
	}
}

func TestStateTracker_SetStopped(t *testing.T) {
	st := NewStateTracker()
	st.SetLaunched("testapp")
	st.SetReady("testapp")

	st.SetStopped("testapp")

	state := st.Get("testapp")
	if state == nil {
		t.Fatal("expected state for testapp, got nil")
	}
	if state.Launched {
		t.Error("expected Launched = false after SetStopped")
	}
	if state.Ready {
		t.Error("expected Ready = false after SetStopped")
	}
	if state.LaunchedAt != nil {
		t.Error("expected LaunchedAt = nil after SetStopped")
	}
	if state.ReadyAt != nil {
		t.Error("expected ReadyAt = nil after SetStopped")
	}
}

func TestStateTracker_SetStopped_Idempotent(t *testing.T) {
	st := NewStateTracker()
	// SetStopped on unknown app should not panic.
	st.SetStopped("nonexistent")

	// SetStopped on already-stopped app is a no-op.
	st.SetLaunched("testapp")
	st.SetStopped("testapp")
	st.SetStopped("testapp") // second call should be fine
	state := st.Get("testapp")
	if state.Launched {
		t.Error("expected Launched = false")
	}
}

func TestStateTracker_Sync(t *testing.T) {
	st := NewStateTracker()
	st.SetLaunched("app1")
	st.SetLaunched("app2")
	st.SetLaunched("app3")

	// Sync to only keep app1 and app3.
	st.Sync([]string{"app1", "app3"})

	if st.Get("app1") == nil {
		t.Error("app1 should be preserved")
	}
	if st.Get("app2") != nil {
		t.Error("app2 should be removed")
	}
	if st.Get("app3") == nil {
		t.Error("app3 should be preserved")
	}

	all := st.All()
	if len(all) != 2 {
		t.Errorf("got %d states after sync, want 2", len(all))
	}
}

func TestStateTracker_Sync_EmptyList(t *testing.T) {
	st := NewStateTracker()
	st.SetLaunched("app1")
	st.Sync([]string{})

	all := st.All()
	if len(all) != 0 {
		t.Errorf("got %d states after sync with empty list, want 0", len(all))
	}
}

func TestStateTracker_SetError(t *testing.T) {
	st := NewStateTracker()
	st.SetLaunched("testapp")
	st.SetError("testapp", "exec: command not found")

	state := st.Get("testapp")
	if state == nil {
		t.Fatal("expected state for testapp, got nil")
	}
	if state.Error != "exec: command not found" {
		t.Errorf("Error = %q, want %q", state.Error, "exec: command not found")
	}
	if state.ErrorAt == nil {
		t.Error("expected ErrorAt to be set")
	}
	// App should still be in launched state (error is supplementary info).
	if !state.Launched {
		t.Error("expected Launched = true")
	}
}

func TestStateTracker_SetError_CreatesEntry(t *testing.T) {
	st := NewStateTracker()
	// SetError on unknown app should create the entry.
	st.SetError("newapp", "something failed")

	state := st.Get("newapp")
	if state == nil {
		t.Fatal("expected state for newapp, got nil")
	}
	if state.Error != "something failed" {
		t.Errorf("Error = %q, want %q", state.Error, "something failed")
	}
}

func TestStateTracker_SetLaunched_ClearsError(t *testing.T) {
	st := NewStateTracker()
	st.SetLaunched("testapp")
	st.SetError("testapp", "previous failure")

	// Re-launch should clear the error.
	st.SetLaunched("testapp")

	state := st.Get("testapp")
	if state == nil {
		t.Fatal("expected state for testapp, got nil")
	}
	if state.Error != "" {
		t.Errorf("Error = %q, want empty after re-launch", state.Error)
	}
	if state.ErrorAt != nil {
		t.Error("expected ErrorAt = nil after re-launch")
	}
}

func TestStateTracker_SetStopped_ClearsError(t *testing.T) {
	st := NewStateTracker()
	st.SetLaunched("testapp")
	st.SetError("testapp", "some error")
	st.SetStopped("testapp")

	state := st.Get("testapp")
	if state == nil {
		t.Fatal("expected state for testapp, got nil")
	}
	if state.Error != "" {
		t.Errorf("Error = %q, want empty after stop", state.Error)
	}
	if state.ErrorAt != nil {
		t.Error("expected ErrorAt = nil after stop")
	}
}

func TestStateTracker_ClearError(t *testing.T) {
	st := NewStateTracker()
	st.SetLaunched("testapp")
	st.SetError("testapp", "bad stuff")
	st.ClearError("testapp")

	state := st.Get("testapp")
	if state == nil {
		t.Fatal("expected state for testapp, got nil")
	}
	if state.Error != "" {
		t.Errorf("Error = %q, want empty after ClearError", state.Error)
	}
	if state.ErrorAt != nil {
		t.Error("expected ErrorAt = nil after ClearError")
	}
}

func TestStateTracker_ClearError_Idempotent(t *testing.T) {
	st := NewStateTracker()
	// ClearError on unknown app should not panic.
	st.ClearError("nonexistent")

	// ClearError on app with no error should be a no-op.
	st.SetLaunched("testapp")
	st.ClearError("testapp")

	state := st.Get("testapp")
	if state.Error != "" {
		t.Errorf("Error = %q, want empty", state.Error)
	}
}

func TestStateTracker_ErrorInJSON(t *testing.T) {
	st := NewStateTracker()
	st.SetLaunched("testapp")
	st.SetError("testapp", "exec failed")

	state := st.Get("testapp")

	data, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	if decoded["error"] != "exec failed" {
		t.Errorf("JSON error = %v, want %q", decoded["error"], "exec failed")
	}
	if decoded["error_at"] == nil {
		t.Error("expected error_at in JSON")
	}
}

func TestStateTracker_NoErrorOmittedInJSON(t *testing.T) {
	st := NewStateTracker()
	st.SetLaunched("testapp")

	state := st.Get("testapp")
	data, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	// error and error_at should be omitted when empty.
	if _, exists := decoded["error"]; exists {
		t.Error("expected error to be omitted from JSON when empty")
	}
	if _, exists := decoded["error_at"]; exists {
		t.Error("expected error_at to be omitted from JSON when empty")
	}
}

func TestPostStopped_Success(t *testing.T) {
	cfg := registry.Config{
		Apps: []registry.AppEntry{
			{Name: "espansr", Chord: "e", Command: "espansr", Environment: "native"},
		},
	}
	reg := registry.NewRegistry(cfg)
	s := NewServer(0, "v0.0.1", reg, nil)
	handler := s.Handler()

	// First launch and ready.
	req := httptest.NewRequest("POST", "/apps/espansr/launched", nil)
	req.RemoteAddr = "127.0.0.1:54321"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	req = httptest.NewRequest("POST", "/apps/espansr/ready", nil)
	req.RemoteAddr = "127.0.0.1:54321"
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Now stop.
	req = httptest.NewRequest("POST", "/apps/espansr/stopped", nil)
	req.RemoteAddr = "127.0.0.1:54321"
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if resp["status"] != "ok" {
		t.Errorf("status = %q, want %q", resp["status"], "ok")
	}
	if resp["state"] != "stopped" {
		t.Errorf("state = %q, want %q", resp["state"], "stopped")
	}

	// Verify state tracker was updated.
	appState := s.state.Get("espansr")
	if appState == nil {
		t.Fatal("expected state for espansr")
	}
	if appState.Launched {
		t.Error("expected Launched = false after stopped")
	}
	if appState.Ready {
		t.Error("expected Ready = false after stopped")
	}
}

func TestPostStopped_UnknownApp(t *testing.T) {
	cfg := registry.Config{
		Apps: []registry.AppEntry{
			{Name: "espansr", Chord: "e", Command: "espansr", Environment: "native"},
		},
	}
	reg := registry.NewRegistry(cfg)
	s := NewServer(0, "v0.0.1", reg, nil)
	handler := s.Handler()

	req := httptest.NewRequest("POST", "/apps/unknown/stopped", nil)
	req.RemoteAddr = "127.0.0.1:54321"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestGetState_WithState(t *testing.T) {
	cfg := registry.Config{
		Apps: []registry.AppEntry{
			{Name: "espansr", Chord: "e", Command: "espansr", Environment: "native"},
		},
	}
	reg := registry.NewRegistry(cfg)
	s := NewServer(0, "v0.0.1", reg, nil)
	handler := s.Handler()

	// Launch and ready the app.
	req := httptest.NewRequest("POST", "/apps/espansr/launched", nil)
	req.RemoteAddr = "127.0.0.1:54321"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	req = httptest.NewRequest("POST", "/apps/espansr/ready", nil)
	req.RemoteAddr = "127.0.0.1:54321"
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Get state.
	req = httptest.NewRequest("GET", "/apps/espansr/state", nil)
	req.RemoteAddr = "127.0.0.1:54321"
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var resp AppState
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if resp.Name != "espansr" {
		t.Errorf("name = %q, want %q", resp.Name, "espansr")
	}
	if !resp.Launched {
		t.Error("expected Launched = true")
	}
	if !resp.Ready {
		t.Error("expected Ready = true")
	}
	if resp.LaunchedAt == nil {
		t.Error("expected LaunchedAt to be set")
	}
	if resp.ReadyAt == nil {
		t.Error("expected ReadyAt to be set")
	}
}

func TestGetState_NoLifecycleYet(t *testing.T) {
	cfg := registry.Config{
		Apps: []registry.AppEntry{
			{Name: "espansr", Chord: "e", Command: "espansr", Environment: "native"},
		},
	}
	reg := registry.NewRegistry(cfg)
	s := NewServer(0, "v0.0.1", reg, nil)
	handler := s.Handler()

	// Get state for registered app that has no lifecycle state.
	req := httptest.NewRequest("GET", "/apps/espansr/state", nil)
	req.RemoteAddr = "127.0.0.1:54321"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var resp AppState
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if resp.Name != "espansr" {
		t.Errorf("name = %q, want %q", resp.Name, "espansr")
	}
	if resp.Launched {
		t.Error("expected Launched = false for untracked app")
	}
	if resp.Ready {
		t.Error("expected Ready = false for untracked app")
	}
}

func TestGetState_UnknownApp(t *testing.T) {
	cfg := registry.Config{
		Apps: []registry.AppEntry{
			{Name: "espansr", Chord: "e", Command: "espansr", Environment: "native"},
		},
	}
	reg := registry.NewRegistry(cfg)
	s := NewServer(0, "v0.0.1", reg, nil)
	handler := s.Handler()

	req := httptest.NewRequest("GET", "/apps/unknown/state", nil)
	req.RemoteAddr = "127.0.0.1:54321"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestGetState_WrongMethod(t *testing.T) {
	cfg := registry.Config{
		Apps: []registry.AppEntry{
			{Name: "espansr", Chord: "e", Command: "espansr", Environment: "native"},
		},
	}
	reg := registry.NewRegistry(cfg)
	s := NewServer(0, "v0.0.1", reg, nil)
	handler := s.Handler()

	req := httptest.NewRequest("POST", "/apps/espansr/state", nil)
	req.RemoteAddr = "127.0.0.1:54321"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}

	allow := rec.Header().Get("Allow")
	if allow != "GET" {
		t.Errorf("Allow = %q, want %q", allow, "GET")
	}
}

func TestStopped_WrongMethod(t *testing.T) {
	cfg := registry.Config{
		Apps: []registry.AppEntry{
			{Name: "espansr", Chord: "e", Command: "espansr", Environment: "native"},
		},
	}
	reg := registry.NewRegistry(cfg)
	s := NewServer(0, "v0.0.1", reg, nil)
	handler := s.Handler()

	req := httptest.NewRequest("GET", "/apps/espansr/stopped", nil)
	req.RemoteAddr = "127.0.0.1:54321"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}

	allow := rec.Header().Get("Allow")
	if allow != "POST" {
		t.Errorf("Allow = %q, want %q", allow, "POST")
	}
}

func TestPostLaunched_Success(t *testing.T) {
	cfg := registry.Config{
		Apps: []registry.AppEntry{
			{Name: "espansr", Chord: "e", Command: "espansr", Environment: "native"},
		},
	}
	reg := registry.NewRegistry(cfg)
	s := NewServer(0, "v0.0.1", reg, nil)
	handler := s.Handler()

	req := httptest.NewRequest("POST", "/apps/espansr/launched", nil)
	req.RemoteAddr = "127.0.0.1:54321"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if resp["status"] != "ok" {
		t.Errorf("status = %q, want %q", resp["status"], "ok")
	}
	if resp["app"] != "espansr" {
		t.Errorf("app = %q, want %q", resp["app"], "espansr")
	}
	if resp["state"] != "launched" {
		t.Errorf("state = %q, want %q", resp["state"], "launched")
	}

	// Verify state tracker was updated.
	appState := s.state.Get("espansr")
	if appState == nil {
		t.Fatal("expected state for espansr")
	}
	if !appState.Launched {
		t.Error("expected Launched = true in state tracker")
	}
}

func TestPostReady_Success(t *testing.T) {
	cfg := registry.Config{
		Apps: []registry.AppEntry{
			{Name: "espansr", Chord: "e", Command: "espansr", Environment: "native"},
		},
	}
	reg := registry.NewRegistry(cfg)
	s := NewServer(0, "v0.0.1", reg, nil)
	handler := s.Handler()

	// First launch, then ready.
	req := httptest.NewRequest("POST", "/apps/espansr/launched", nil)
	req.RemoteAddr = "127.0.0.1:54321"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	req = httptest.NewRequest("POST", "/apps/espansr/ready", nil)
	req.RemoteAddr = "127.0.0.1:54321"
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if resp["state"] != "ready" {
		t.Errorf("state = %q, want %q", resp["state"], "ready")
	}

	appState := s.state.Get("espansr")
	if appState == nil {
		t.Fatal("expected state for espansr")
	}
	if !appState.Ready {
		t.Error("expected Ready = true in state tracker")
	}
}

func TestPostLaunched_UnknownApp(t *testing.T) {
	cfg := registry.Config{
		Apps: []registry.AppEntry{
			{Name: "espansr", Chord: "e", Command: "espansr", Environment: "native"},
		},
	}
	reg := registry.NewRegistry(cfg)
	s := NewServer(0, "v0.0.1", reg, nil)
	handler := s.Handler()

	req := httptest.NewRequest("POST", "/apps/unknown/launched", nil)
	req.RemoteAddr = "127.0.0.1:54321"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	var errResp ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
		t.Fatalf("decoding error response: %v", err)
	}
	if errResp.Code != "not_found" {
		t.Errorf("code = %q, want %q", errResp.Code, "not_found")
	}
}

func TestAppAction_BadPath(t *testing.T) {
	cfg := registry.Config{
		Apps: []registry.AppEntry{
			{Name: "espansr", Chord: "e", Command: "espansr", Environment: "native"},
		},
	}
	reg := registry.NewRegistry(cfg)
	s := NewServer(0, "v0.0.1", reg, nil)
	handler := s.Handler()

	tests := []struct {
		name string
		path string
	}{
		{"no action", "/apps/espansr/"},
		{"just name", "/apps/espansr"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", tt.path, nil)
			req.RemoteAddr = "127.0.0.1:54321"
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			// Either 400 bad request or 405 method not allowed is acceptable
			// depending on routing (POST to /apps hits the GET-only apps handler).
			if rec.Code != http.StatusBadRequest && rec.Code != http.StatusMethodNotAllowed {
				t.Errorf("status = %d, want 400 or 405", rec.Code)
			}
		})
	}
}

func TestAppAction_UnknownAction(t *testing.T) {
	cfg := registry.Config{
		Apps: []registry.AppEntry{
			{Name: "espansr", Chord: "e", Command: "espansr", Environment: "native"},
		},
	}
	reg := registry.NewRegistry(cfg)
	s := NewServer(0, "v0.0.1", reg, nil)
	handler := s.Handler()

	req := httptest.NewRequest("POST", "/apps/espansr/unknown", nil)
	req.RemoteAddr = "127.0.0.1:54321"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestAppAction_WrongMethod(t *testing.T) {
	cfg := registry.Config{
		Apps: []registry.AppEntry{
			{Name: "espansr", Chord: "e", Command: "espansr", Environment: "native"},
		},
	}
	reg := registry.NewRegistry(cfg)
	s := NewServer(0, "v0.0.1", reg, nil)
	handler := s.Handler()

	req := httptest.NewRequest("GET", "/apps/espansr/launched", nil)
	req.RemoteAddr = "127.0.0.1:54321"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}

	// RFC 7231: 405 responses MUST include an Allow header.
	allow := rec.Header().Get("Allow")
	if allow == "" {
		t.Error("405 response missing Allow header")
	}
}
