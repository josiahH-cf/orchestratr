package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// HealthServer is a minimal localhost HTTP server that exposes a /health endpoint.
type HealthServer struct {
	port   int
	server *http.Server
	ln     net.Listener
}

// NewHealthServer creates a health server bound to 127.0.0.1 on the given port.
// Use port 0 for automatic assignment.
func NewHealthServer(port int) *HealthServer {
	return &HealthServer{port: port}
}

// ListenAddr returns the bind address (always localhost).
func (h *HealthServer) ListenAddr() string {
	return "127.0.0.1"
}

// Start begins listening and serving. It blocks until the server is stopped.
func (h *HealthServer) Start() error {
	addr := fmt.Sprintf("127.0.0.1:%d", h.port)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", h.handleHealth)

	h.server = &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	var err error
	h.ln, err = net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("health server listen: %w", err)
	}

	// Update port in case 0 was used.
	h.port = h.ln.Addr().(*net.TCPAddr).Port

	return h.server.Serve(h.ln)
}

// Stop gracefully shuts down the health server.
func (h *HealthServer) Stop() {
	if h.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = h.server.Shutdown(ctx)
	}
}

// Port returns the port the server is listening on.
func (h *HealthServer) Port() int {
	return h.port
}

func (h *HealthServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

// DefaultPortFilePath returns the default location for the port discovery file.
func DefaultPortFilePath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = filepath.Join(os.Getenv("HOME"), ".config")
	}
	return filepath.Join(dir, "orchestratr", "port")
}

// WritePortFile writes the API port to a discovery file so other apps can find it.
func WritePortFile(path string, port int) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating port file directory: %w", err)
	}
	return os.WriteFile(path, []byte(strconv.Itoa(port)), 0o644)
}

// RemovePortFile removes the port discovery file on daemon shutdown.
func RemovePortFile(path string) {
	_ = os.Remove(path)
}
