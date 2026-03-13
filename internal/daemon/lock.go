// Package daemon provides the orchestratr background daemon lifecycle,
// single-instance lock, signal handling, and state management.
package daemon

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// Lock represents a PID-based single-instance lock file.
type Lock struct {
	path string
}

// DefaultLockPath returns the default PID file location.
func DefaultLockPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = filepath.Join(os.Getenv("HOME"), ".config")
	}
	return filepath.Join(dir, "orchestratr", "orchestratr.pid")
}

// AcquireLock attempts to create a PID lock file. It returns an error if
// another instance is already running. Stale locks (PID no longer alive)
// are automatically recovered.
func AcquireLock(path string) (*Lock, error) {
	// Ensure directory exists.
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("creating lock directory: %w", err)
	}

	// Check for existing PID file.
	data, err := os.ReadFile(path)
	if err == nil {
		pid, parseErr := strconv.Atoi(string(data))
		if parseErr == nil && processAlive(pid) {
			return nil, fmt.Errorf("another instance is running (PID %d)", pid)
		}
		// Stale lock — remove it.
		_ = os.Remove(path)
	}

	// Write our PID.
	pidStr := strconv.Itoa(os.Getpid())
	if err := os.WriteFile(path, []byte(pidStr), 0o644); err != nil {
		return nil, fmt.Errorf("writing PID file: %w", err)
	}

	return &Lock{path: path}, nil
}

// Release removes the PID lock file.
func (l *Lock) Release() error {
	return os.Remove(l.path)
}

// ProcessAlive checks whether a process with the given PID is running.
func ProcessAlive(pid int) bool {
	return processAlive(pid)
}
