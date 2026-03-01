package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
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
