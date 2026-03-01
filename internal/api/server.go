package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/josiahH-cf/orchestratr/internal/registry"
)

// ReloadFunc is called when POST /reload is received. It should
// reload the config and return an error if the reload failed.
type ReloadFunc func() (*registry.Config, error)

// Server is the localhost HTTP API server for orchestratr.
type Server struct {
	port     int
	version  string
	reg      *registry.Registry
	reloadFn ReloadFunc
	state    *StateTracker

	mu     sync.Mutex
	server *http.Server
	ln     net.Listener
	ready  chan struct{}
}

// NewServer creates a new API server. Use port 0 for automatic
// assignment. The registry and reloadFn may be nil for basic
// health-only operation.
func NewServer(port int, version string, reg *registry.Registry, reloadFn ReloadFunc) *Server {
	return &Server{
		port:     port,
		version:  version,
		reg:      reg,
		reloadFn: reloadFn,
		state:    NewStateTracker(),
		ready:    make(chan struct{}),
	}
}

// Handler returns the HTTP handler with all routes and middleware.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/apps", s.handleApps)
	mux.HandleFunc("/apps/", s.handleAppAction)
	mux.HandleFunc("/reload", s.handleReload)

	// Catch-all for unknown routes.
	mux.HandleFunc("/", s.handleNotFound)

	return localhostOnly(mux)
}

// Start begins listening on 127.0.0.1 and serving requests. It
// blocks until the server is stopped. Returns http.ErrServerClosed
// on graceful shutdown.
func (s *Server) Start() error {
	addr := fmt.Sprintf("127.0.0.1:%d", s.port)

	handler := s.Handler()
	s.mu.Lock()
	s.server = &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	var err error
	s.ln, err = net.Listen("tcp", addr)
	if err != nil {
		s.mu.Unlock()
		return fmt.Errorf("api server listen: %w", err)
	}

	// Update port in case 0 was used.
	s.port = s.ln.Addr().(*net.TCPAddr).Port
	s.mu.Unlock()

	close(s.ready)

	return s.server.Serve(s.ln)
}

// Stop gracefully shuts down the server.
func (s *Server) Stop() {
	s.mu.Lock()
	srv := s.server
	s.mu.Unlock()

	if srv != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}
}

// Port returns the port the server is listening on.
func (s *Server) Port() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.port
}

// WaitReady blocks until the server is listening or the timeout
// (in seconds) expires. Returns true if the server is ready.
func (s *Server) WaitReady(timeoutSec int) bool {
	select {
	case <-s.ready:
		return true
	case <-time.After(time.Duration(timeoutSec) * time.Second):
		return false
	}
}

// localhostOnly is middleware that rejects requests not originating
// from 127.0.0.1 or ::1 with a 403 Forbidden response.
func localhostOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := r.RemoteAddr
		// Strip port if present.
		if h, _, err := net.SplitHostPort(host); err == nil {
			host = h
		}

		if host != "127.0.0.1" && host != "::1" {
			writeError(w, http.StatusForbidden, "forbidden", "requests from non-localhost origins are rejected")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// handleHealth responds with the server's health status and version.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "use GET")
		return
	}
	writeJSON(w, http.StatusOK, HealthResponse{
		Status:  "ok",
		Version: s.version,
	})
}

// handleApps returns the app registry as a JSON array.
func (s *Server) handleApps(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "use GET")
		return
	}
	if s.reg == nil {
		writeJSON(w, http.StatusOK, []any{})
		return
	}
	writeJSON(w, http.StatusOK, s.reg.Apps())
}

// handleAppAction routes /apps/{name}/{action} requests.
func (s *Server) handleAppAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "use POST")
		return
	}

	// Parse /apps/{name}/{action}
	path := strings.TrimPrefix(r.URL.Path, "/apps/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "expected /apps/{name}/{action}")
		return
	}

	name := parts[0]
	action := parts[1]

	// Verify app exists in registry.
	if s.reg != nil {
		if _, found := s.reg.FindByName(name); !found {
			writeError(w, http.StatusNotFound, "not_found", fmt.Sprintf("app %q not found", name))
			return
		}
	}

	switch action {
	case "launched":
		s.handleLaunched(w, r, name)
	case "ready":
		s.handleReady(w, r, name)
	default:
		writeError(w, http.StatusNotFound, "not_found", fmt.Sprintf("unknown action %q", action))
	}
}

// handleLaunched records that an app has been launched.
func (s *Server) handleLaunched(w http.ResponseWriter, _ *http.Request, name string) {
	if s.state != nil {
		s.state.SetLaunched(name)
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "app": name, "state": "launched"})
}

// handleReady records that an app is fully initialized.
func (s *Server) handleReady(w http.ResponseWriter, _ *http.Request, name string) {
	if s.state != nil {
		s.state.SetReady(name)
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "app": name, "state": "ready"})
}

// handleReload triggers a config hot-reload.
func (s *Server) handleReload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "use POST")
		return
	}
	if s.reloadFn == nil {
		writeError(w, http.StatusServiceUnavailable, "unavailable", "reload not configured")
		return
	}
	cfg, err := s.reloadFn()
	if err != nil {
		writeError(w, http.StatusBadRequest, "reload_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"apps":   cfg.Apps,
	})
}

// handleNotFound is the catch-all handler for unknown routes.
func (s *Server) handleNotFound(w http.ResponseWriter, _ *http.Request) {
	writeError(w, http.StatusNotFound, "not_found", "endpoint not found")
}
