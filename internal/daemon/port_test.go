package daemon

import (
	"os"
	"path/filepath"
	"testing"
)

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

func TestDefaultPortFilePath(t *testing.T) {
	p := DefaultPortFilePath()
	if p == "" {
		t.Error("DefaultPortFilePath() should not be empty")
	}
}
