package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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

// TriggerFunc is called when POST /trigger is received, simulating
// a leader key press. This enables the Wayland manual keybinding
// fallback (orchestratr trigger).
type TriggerFunc func() error

// LaunchFunc is called when POST /apps/{name}/launch is received.
// It launches (or focuses) the named app and returns the PID on
// success. This enables launching apps without the hotkey engine
// (e.g., on WSL, Wayland, or via CLI).
type LaunchFunc func(name string) (pid int, err error)

// maxRequestBody is the maximum allowed request body size (1 MB).
const maxRequestBody = 1 << 20

// Server is the localhost HTTP API server for orchestratr.
type Server struct {
	port      int
	version   string
	reg       *registry.Registry
	reloadFn  ReloadFunc
	triggerFn TriggerFunc
	launchFn  LaunchFunc
	state     *StateTracker
	logger    *log.Logger

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
		logger:   log.New(log.Default().Writer(), "api: ", log.LstdFlags),
		ready:    make(chan struct{}),
	}
}

// SetLogger replaces the server's logger.
func (s *Server) SetLogger(l *log.Logger) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logger = l
}

// SetTriggerFunc sets the function called when POST /trigger is received.
func (s *Server) SetTriggerFunc(fn TriggerFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.triggerFn = fn
}

// SetLaunchFunc sets the function called when POST /apps/{name}/launch
// is received. This allows launching apps directly without the hotkey
// engine.
func (s *Server) SetLaunchFunc(fn LaunchFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.launchFn = fn
}

// State returns the server's StateTracker for external callers that
// need to synchronize state (e.g., during config reload).
func (s *Server) State() *StateTracker {
	return s.state
}

// Handler returns the HTTP handler with all routes and middleware.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/apps", s.handleApps)
	mux.HandleFunc("/apps/", s.handleAppAction)
	mux.HandleFunc("/reload", s.handleReload)
	mux.HandleFunc("/trigger", s.handleTrigger)

	// Catch-all for unknown routes.
	mux.HandleFunc("/", s.handleNotFound)

	return s.requestLogger(localhostOnly(mux))
}

// Start begins listening on 127.0.0.1 and serving requests. It
// blocks until the server is stopped. Returns http.ErrServerClosed
// on graceful shutdown.
func (s *Server) Start() error {
	addr := fmt.Sprintf("127.0.0.1:%d", s.port)

	handler := s.Handler()
	s.mu.Lock()

	// Guard against double-start: if ready is already closed, reject.
	select {
	case <-s.ready:
		s.mu.Unlock()
		return fmt.Errorf("api server already started")
	default:
	}

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

// requestLogger is middleware that logs each request's method, path,
// and response status.
func (s *Server) requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		s.logger.Printf("%s %s %d %s", r.Method, r.URL.Path, sw.status, time.Since(start).Round(time.Microsecond))
	})
}

// statusWriter wraps http.ResponseWriter to capture the status code.
type statusWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (sw *statusWriter) WriteHeader(code int) {
	if !sw.wroteHeader {
		sw.status = code
		sw.wroteHeader = true
	}
	sw.ResponseWriter.WriteHeader(code)
}

// methodNotAllowed writes a 405 response with the required Allow header.
func methodNotAllowed(w http.ResponseWriter, allowed string) {
	w.Header().Set("Allow", allowed)
	writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "use "+allowed)
}

// handleHealth responds with the server's health status and version.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
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
		methodNotAllowed(w, http.MethodGet)
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
	// Limit request body size.
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBody)

	// Parse /apps/{name}/{action}
	path := strings.TrimPrefix(r.URL.Path, "/apps/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "expected /apps/{name}/{action}")
		return
	}

	name := parts[0]
	action := parts[1]

	// Validate app name: must be non-empty, no slashes, no path traversal.
	if strings.ContainsAny(name, "/\\") || name == "." || name == ".." {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid app name")
		return
	}

	// Verify app exists in registry. If registry is nil, we cannot
	// validate app names so reject all lifecycle requests.
	if s.reg == nil {
		writeError(w, http.StatusServiceUnavailable, "unavailable", "registry not loaded")
		return
	}
	if _, found := s.reg.FindByName(name); !found {
		writeError(w, http.StatusNotFound, "not_found", fmt.Sprintf("app %q not found", name))
		return
	}

	switch action {
	case "launched":
		if r.Method != http.MethodPost {
			methodNotAllowed(w, http.MethodPost)
			return
		}
		s.handleLaunched(w, r, name)
	case "ready":
		if r.Method != http.MethodPost {
			methodNotAllowed(w, http.MethodPost)
			return
		}
		s.handleReady(w, r, name)
	case "stopped":
		if r.Method != http.MethodPost {
			methodNotAllowed(w, http.MethodPost)
			return
		}
		s.handleStopped(w, r, name)
	case "state":
		if r.Method != http.MethodGet {
			methodNotAllowed(w, http.MethodGet)
			return
		}
		s.handleState(w, r, name)
	case "launch":
		if r.Method != http.MethodPost {
			methodNotAllowed(w, http.MethodPost)
			return
		}
		s.handleLaunch(w, r, name)
	default:
		writeError(w, http.StatusNotFound, "not_found", fmt.Sprintf("unknown action %q", action))
	}
}

// handleLaunched records that an app has been launched.
func (s *Server) handleLaunched(w http.ResponseWriter, _ *http.Request, name string) {
	s.state.SetLaunched(name)
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "app": name, "state": "launched"})
}

// handleReady records that an app is fully initialized.
func (s *Server) handleReady(w http.ResponseWriter, _ *http.Request, name string) {
	s.state.SetReady(name)
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "app": name, "state": "ready"})
}

// handleStopped marks an app as stopped (not launched, not ready).
func (s *Server) handleStopped(w http.ResponseWriter, _ *http.Request, name string) {
	s.state.SetStopped(name)
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "app": name, "state": "stopped"})
}

// handleState returns the lifecycle state for a single app.
func (s *Server) handleState(w http.ResponseWriter, _ *http.Request, name string) {
	appState := s.state.Get(name)
	if appState == nil {
		appState = &AppState{Name: name}
	}
	writeJSON(w, http.StatusOK, appState)
}

// handleLaunch launches or focuses the named app.
func (s *Server) handleLaunch(w http.ResponseWriter, _ *http.Request, name string) {
	s.mu.Lock()
	fn := s.launchFn
	s.mu.Unlock()

	if fn == nil {
		writeError(w, http.StatusServiceUnavailable, "unavailable", "launcher not configured")
		return
	}
	pid, err := fn(name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "launch_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "app": name, "pid": pid})
}

// handleReload triggers a config hot-reload.
func (s *Server) handleReload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
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

// handleTrigger simulates a leader key press via the API. This is
// used by the Wayland manual keybinding fallback.
//
// When the request body contains a "chord" field, the handler looks
// up the chord in the registry and launches the matching app directly
// without requiring the hotkey engine (OI-9, OI-12). This enables
// compositor keybindings like: orchestratr trigger c
func (s *Server) handleTrigger(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}

	// Check for an optional chord in the request body.
	var body struct {
		Chord string `json:"chord"`
	}
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}

	// If a chord is provided, bypass the hotkey engine and launch
	// the matching app directly.
	if body.Chord != "" {
		if s.reg == nil {
			writeError(w, http.StatusServiceUnavailable, "unavailable", "registry not loaded")
			return
		}
		app, found := s.reg.FindByChord(body.Chord)
		if !found {
			writeError(w, http.StatusNotFound, "not_found", fmt.Sprintf("no app with chord %q", body.Chord))
			return
		}

		s.mu.Lock()
		fn := s.launchFn
		s.mu.Unlock()

		if fn == nil {
			writeError(w, http.StatusServiceUnavailable, "unavailable", "launcher not configured")
			return
		}
		pid, err := fn(app.Name)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "launch_failed", err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "app": app.Name, "pid": pid})
		return
	}

	// No chord: simulate a leader key press (original behavior).
	s.mu.Lock()
	fn := s.triggerFn
	s.mu.Unlock()

	if fn == nil {
		writeError(w, http.StatusServiceUnavailable, "unavailable", "hotkey engine not running")
		return
	}
	if err := fn(); err != nil {
		writeError(w, http.StatusInternalServerError, "trigger_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleNotFound is the catch-all handler for unknown routes.
func (s *Server) handleNotFound(w http.ResponseWriter, _ *http.Request) {
	writeError(w, http.StatusNotFound, "not_found", "endpoint not found")
}
