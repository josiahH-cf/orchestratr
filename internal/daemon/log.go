package daemon

import (
	"fmt"
	"os"
	"path/filepath"
)

// DefaultLogPath returns the default log file location.
func DefaultLogPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = filepath.Join(os.Getenv("HOME"), ".config")
	}
	return filepath.Join(dir, "orchestratr", "orchestratr.log")
}

// SetupLogFile opens (or creates) the log file for writing. The caller is
// responsible for closing the returned file. The file is opened in append
// mode so existing logs are preserved.
func SetupLogFile(path string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("creating log directory: %w", err)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("opening log file: %w", err)
	}
	return f, nil
}
