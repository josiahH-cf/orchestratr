// Package gui provides a web-based management GUI for editing the
// orchestratr configuration. It starts a localhost HTTP server, serves
// an embedded HTML/JS single-page app, and opens the user's default
// browser. The GUI reads and writes the same YAML config file that can
// also be edited by hand.
package gui

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/josiahH-cf/orchestratr/internal/registry"
)

//go:embed static/*
var staticFiles embed.FS

// Server is the GUI web server.
type Server struct {
	cfgPath string
	apiPort int // daemon API port (0 = daemon not running)
	logger  *log.Logger

	mu     sync.Mutex
	server *http.Server
	ln     net.Listener
	port   int
}

// NewServer creates a GUI server that edits the config at cfgPath.
// apiPort is the daemon's API port (for triggering reloads and
// fetching live state); pass 0 if the daemon is not running.
func NewServer(cfgPath string, apiPort int, logger *log.Logger) *Server {
	if logger == nil {
		logger = log.New(os.Stderr, "gui: ", log.LstdFlags)
	}
	return &Server{
		cfgPath: cfgPath,
		apiPort: apiPort,
		logger:  logger,
	}
}

// Start begins listening on an ephemeral port and serving the GUI.
// It does not block — call Stop to shut down.
func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.ln != nil {
		return fmt.Errorf("gui server already started")
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("gui listen: %w", err)
	}
	s.ln = ln
	s.port = ln.Addr().(*net.TCPAddr).Port

	mux := http.NewServeMux()
	s.registerRoutes(mux)

	s.server = &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		if err := s.server.Serve(ln); err != nil && err != http.ErrServerClosed {
			s.logger.Printf("gui server error: %v", err)
		}
	}()

	return nil
}

// Port returns the port the server is listening on.
func (s *Server) Port() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.port
}

// URL returns the full URL for the GUI.
func (s *Server) URL() string {
	return fmt.Sprintf("http://127.0.0.1:%d", s.Port())
}

// Stop gracefully shuts down the GUI server.
func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.server == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := s.server.Shutdown(ctx)
	s.server = nil
	s.ln = nil
	return err
}

// OpenBrowser opens the GUI URL in the user's default browser.
// Returns an error if the browser cannot be opened (e.g., headless).
func (s *Server) OpenBrowser() error {
	url := s.URL()
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform %s — open %s manually", runtime.GOOS, url)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("opening browser: %w (open %s manually)", err, url)
	}
	return nil
}

// Handler returns the HTTP handler for testing.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	s.registerRoutes(mux)
	return mux
}

// registerRoutes sets up the HTTP routes.
func (s *Server) registerRoutes(mux *http.ServeMux) {
	// Serve embedded static files.
	staticSub, _ := fs.Sub(staticFiles, "static")
	mux.Handle("/", http.FileServer(http.FS(staticSub)))

	// Config API.
	mux.HandleFunc("/api/config", s.handleConfig)
	mux.HandleFunc("/api/daemon-info", s.handleDaemonInfo)
}

// handleConfig handles GET (read config) and PUT (write config).
func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGetConfig(w, r)
	case http.MethodPut:
		s.handlePutConfig(w, r)
	default:
		w.Header().Set("Allow", "GET, PUT")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetConfig reads the config file and returns it as JSON.
func (s *Server) handleGetConfig(w http.ResponseWriter, _ *http.Request) {
	cfg, err := registry.LoadAndValidate(s.cfgPath)
	if err != nil {
		// Try loading without validation.
		cfg2, err2 := registry.Load(s.cfgPath)
		if err2 != nil {
			def := registry.DefaultConfig()
			cfg = &def
		} else {
			cfg = cfg2
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cfg)
}

// handlePutConfig receives a JSON config, validates it, writes YAML,
// and triggers a daemon reload if the daemon is running.
func (s *Server) handlePutConfig(w http.ResponseWriter, r *http.Request) {
	var cfg registry.Config
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	// Normalize empty environments.
	for i := range cfg.Apps {
		if cfg.Apps[i].Environment == "" {
			cfg.Apps[i].Environment = "native"
		}
	}

	// Validate.
	if errs := registry.ValidateConfig(&cfg); len(errs) > 0 {
		writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("validation: %v", errs[0]))
		return
	}

	// Write YAML.
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "marshal: "+err.Error())
		return
	}
	if err := os.WriteFile(s.cfgPath, data, 0o644); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "write: "+err.Error())
		return
	}

	// Trigger daemon reload if available.
	reloaded := false
	if s.apiPort > 0 {
		reloaded = s.triggerDaemonReload()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status":   "ok",
		"reloaded": reloaded,
		"apps":     len(cfg.Apps),
	})
}

// handleDaemonInfo returns information about the daemon connection.
func (s *Server) handleDaemonInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	info := map[string]any{
		"api_port":  s.apiPort,
		"connected": false,
	}

	if s.apiPort > 0 {
		client := &http.Client{Timeout: 2 * time.Second}
		resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/health", s.apiPort))
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				info["connected"] = true
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// triggerDaemonReload sends POST /reload to the daemon. Returns true
// on success.
func (s *Server) triggerDaemonReload() bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Post(
		fmt.Sprintf("http://127.0.0.1:%d/reload", s.apiPort),
		"application/json", nil)
	if err != nil {
		s.logger.Printf("daemon reload failed: %v", err)
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// writeJSONError writes a JSON error response.
func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
