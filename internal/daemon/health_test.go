package daemon

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestHealthServer_ReturnsOK(t *testing.T) {
	// Find a free port.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	srv := NewHealthServer(port)
	go func() { _ = srv.Start() }()
	defer srv.Stop()

	// Wait for the server to be ready.
	url := fmt.Sprintf("http://127.0.0.1:%d/health", port)
	var resp *http.Response
	for i := 0; i < 50; i++ {
		resp, err = http.Get(url)
		if err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("GET /health error after retries: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decoding body: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("status = %q, want %q", body["status"], "ok")
	}
}

func TestHealthServer_LocalhostOnly(t *testing.T) {
	srv := NewHealthServer(0)

	addr := srv.ListenAddr()
	if addr != "127.0.0.1" {
		t.Errorf("listen addr = %q, want 127.0.0.1", addr)
	}
}

func TestWritePortFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "port")

	if err := WritePortFile(path, 9876); err != nil {
		t.Fatalf("WritePortFile() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "9876" {
		t.Errorf("port file = %q, want %q", string(data), "9876")
	}
}

func TestRemovePortFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "port")

	if err := WritePortFile(path, 9876); err != nil {
		t.Fatal(err)
	}

	RemovePortFile(path)

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("port file should be removed")
	}
}
