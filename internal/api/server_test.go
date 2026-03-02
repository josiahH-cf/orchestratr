package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/josiahH-cf/orchestratr/internal/registry"
)

func TestNewServer(t *testing.T) {
	s := NewServer(0, "v0.1.0-test", nil, nil)
	if s == nil {
		t.Fatal("NewServer returned nil")
	}
	if s.version != "v0.1.0-test" {
		t.Errorf("version = %q, want %q", s.version, "v0.1.0-test")
	}
}

func TestLocalhostMiddleware_AllowsLocalhost(t *testing.T) {
	handler := localhostOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		name       string
		remoteAddr string
		wantCode   int
	}{
		{"127.0.0.1 with port", "127.0.0.1:12345", http.StatusOK},
		{"127.0.0.1 no port", "127.0.0.1", http.StatusOK},
		{"::1 with port", "[::1]:12345", http.StatusOK},
		{"::1 no port", "::1", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.remoteAddr
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != tt.wantCode {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantCode)
			}
		})
	}
}

func TestLocalhostMiddleware_RejectsNonLocal(t *testing.T) {
	handler := localhostOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		name       string
		remoteAddr string
	}{
		{"external IPv4", "192.168.1.100:12345"},
		{"external IPv4 no port", "10.0.0.1"},
		{"public IP", "8.8.8.8:443"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.remoteAddr
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != http.StatusForbidden {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
			}
			// Verify error JSON format.
			var errResp ErrorResponse
			if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
				t.Fatalf("decoding error response: %v", err)
			}
			if errResp.Code == "" {
				t.Error("error response missing code field")
			}
			if errResp.Error == "" {
				t.Error("error response missing error field")
			}
		})
	}
}

func TestHealthEndpoint(t *testing.T) {
	s := NewServer(0, "v1.2.3", nil, nil)
	handler := s.Handler()

	req := httptest.NewRequest("GET", "/health", nil)
	req.RemoteAddr = "127.0.0.1:54321"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp HealthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decoding health response: %v", err)
	}
	if resp.Status != "ok" {
		t.Errorf("status = %q, want %q", resp.Status, "ok")
	}
	if resp.Version != "v1.2.3" {
		t.Errorf("version = %q, want %q", resp.Version, "v1.2.3")
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}
}

func TestHealthEndpoint_WrongMethod(t *testing.T) {
	s := NewServer(0, "v0.0.1", nil, nil)
	handler := s.Handler()

	req := httptest.NewRequest("POST", "/health", nil)
	req.RemoteAddr = "127.0.0.1:54321"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}

	// RFC 7231 §6.5.5: 405 MUST include Allow header.
	allow := rec.Header().Get("Allow")
	if allow == "" {
		t.Error("405 response missing Allow header")
	}
	if allow != "GET" {
		t.Errorf("Allow = %q, want %q", allow, "GET")
	}
}

func TestUnknownRoute_Returns404(t *testing.T) {
	s := NewServer(0, "v0.0.1", nil, nil)
	handler := s.Handler()

	req := httptest.NewRequest("GET", "/nonexistent", nil)
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

func TestWriteJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	writeJSON(rec, http.StatusCreated, map[string]string{"key": "value"})

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decoding body: %v", err)
	}
	if body["key"] != "value" {
		t.Errorf("key = %q, want %q", body["key"], "value")
	}
}

func TestWriteError(t *testing.T) {
	rec := httptest.NewRecorder()
	writeError(rec, http.StatusBadRequest, "bad_request", "something went wrong")

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	var errResp ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
		t.Fatalf("decoding error response: %v", err)
	}
	if errResp.Code != "bad_request" {
		t.Errorf("code = %q, want %q", errResp.Code, "bad_request")
	}
	if errResp.Error != "something went wrong" {
		t.Errorf("error = %q, want %q", errResp.Error, "something went wrong")
	}
}

func TestServerStartStop(t *testing.T) {
	s := NewServer(0, "v0.0.1", nil, nil)

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.Start()
	}()

	// Wait for the server to be ready.
	if !s.WaitReady(2) {
		t.Fatal("server did not become ready")
	}

	port := s.Port()
	if port == 0 {
		t.Fatal("port should not be 0 after start")
	}

	s.Stop()

	err := <-errCh
	if err != nil && err != http.ErrServerClosed {
		t.Fatalf("unexpected server error: %v", err)
	}
}

func TestServerDoubleStart(t *testing.T) {
	s := NewServer(0, "v0.0.1", nil, nil)

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.Start()
	}()

	if !s.WaitReady(2) {
		t.Fatal("server did not become ready")
	}
	defer s.Stop()

	// Second Start should return an error, not panic.
	err := s.Start()
	if err == nil {
		t.Fatal("expected error on double start")
	}
}

func TestAppAction_NilRegistry(t *testing.T) {
	s := NewServer(0, "v0.0.1", nil, nil)
	handler := s.Handler()

	req := httptest.NewRequest("POST", "/apps/anything/launched", nil)
	req.RemoteAddr = "127.0.0.1:54321"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}

	var errResp ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
		t.Fatalf("decoding error response: %v", err)
	}
	if errResp.Code != "unavailable" {
		t.Errorf("code = %q, want %q", errResp.Code, "unavailable")
	}
}

func TestAppAction_InvalidName(t *testing.T) {
	s := NewServer(0, "v0.0.1", nil, nil)
	handler := s.Handler()

	tests := []struct {
		name     string
		path     string
		wantCode int
	}{
		// Go's http.ServeMux auto-redirects /.. and /. with 301 —
		// path traversal is safe via stdlib before we even see it.
		{"dot dot redirect", "/apps/../launched", http.StatusMovedPermanently},
		{"single dot redirect", "/apps/./launched", http.StatusMovedPermanently},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", tt.path, nil)
			req.RemoteAddr = "127.0.0.1:54321"
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantCode {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantCode)
			}
		})
	}
}

func TestStateTracker_AllDeterministic(t *testing.T) {
	st := NewStateTracker()
	st.SetLaunched("zebra")
	st.SetLaunched("alpha")
	st.SetLaunched("middle")

	all := st.All()
	if len(all) != 3 {
		t.Fatalf("got %d states, want 3", len(all))
	}
	if all[0].Name != "alpha" {
		t.Errorf("all[0].Name = %q, want %q", all[0].Name, "alpha")
	}
	if all[1].Name != "middle" {
		t.Errorf("all[1].Name = %q, want %q", all[1].Name, "middle")
	}
	if all[2].Name != "zebra" {
		t.Errorf("all[2].Name = %q, want %q", all[2].Name, "zebra")
	}
}
func TestTrigger_NoHandler(t *testing.T) {
	s := NewServer(0, "test", nil, nil)
	handler := s.Handler()

	req := httptest.NewRequest(http.MethodPost, "/trigger", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}
}

func TestTrigger_Success(t *testing.T) {
	s := NewServer(0, "test", nil, nil)
	called := false
	s.SetTriggerFunc(func() error {
		called = true
		return nil
	})

	handler := s.Handler()
	req := httptest.NewRequest(http.MethodPost, "/trigger", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if !called {
		t.Error("trigger func was not called")
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp["status"] != "ok" {
		t.Errorf("status = %q, want %q", resp["status"], "ok")
	}
}

func TestTrigger_MethodNotAllowed(t *testing.T) {
	s := NewServer(0, "test", nil, nil)
	handler := s.Handler()

	req := httptest.NewRequest(http.MethodGet, "/trigger", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
	if w.Header().Get("Allow") != http.MethodPost {
		t.Errorf("Allow = %q, want %q", w.Header().Get("Allow"), http.MethodPost)
	}
}

func TestLaunch_NoHandler(t *testing.T) {
	cfg := registry.Config{
		Apps: []registry.AppEntry{{Name: "testapp", Chord: "t", Command: "echo hi"}},
	}
	reg := registry.NewRegistry(cfg)
	s := NewServer(0, "test", reg, nil)
	handler := s.Handler()

	req := httptest.NewRequest(http.MethodPost, "/apps/testapp/launch", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}
}

func TestLaunch_Success(t *testing.T) {
	cfg := registry.Config{
		Apps: []registry.AppEntry{{Name: "testapp", Chord: "t", Command: "echo hi"}},
	}
	reg := registry.NewRegistry(cfg)
	s := NewServer(0, "test", reg, nil)

	called := false
	s.SetLaunchFunc(func(name string) (int, error) {
		called = true
		if name != "testapp" {
			t.Errorf("launch name = %q, want %q", name, "testapp")
		}
		return 12345, nil
	})

	handler := s.Handler()
	req := httptest.NewRequest(http.MethodPost, "/apps/testapp/launch", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if !called {
		t.Error("launch func was not called")
	}

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp["status"] != "ok" {
		t.Errorf("status = %q, want %q", resp["status"], "ok")
	}
	if pid, ok := resp["pid"].(float64); !ok || pid != 12345 {
		t.Errorf("pid = %v, want 12345", resp["pid"])
	}
}

func TestLaunch_AppNotFound(t *testing.T) {
	cfg := registry.Config{
		Apps: []registry.AppEntry{{Name: "testapp", Chord: "t", Command: "echo hi"}},
	}
	reg := registry.NewRegistry(cfg)
	s := NewServer(0, "test", reg, nil)
	s.SetLaunchFunc(func(name string) (int, error) {
		return 0, nil
	})
	handler := s.Handler()

	req := httptest.NewRequest(http.MethodPost, "/apps/nonexistent/launch", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestLaunch_MethodNotAllowed(t *testing.T) {
	cfg := registry.Config{
		Apps: []registry.AppEntry{{Name: "testapp", Chord: "t", Command: "echo hi"}},
	}
	reg := registry.NewRegistry(cfg)
	s := NewServer(0, "test", reg, nil)
	handler := s.Handler()

	req := httptest.NewRequest(http.MethodGet, "/apps/testapp/launch", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}
