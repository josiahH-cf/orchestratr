package daemon

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

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
