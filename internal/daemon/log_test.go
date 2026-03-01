package daemon

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetupLogFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "orchestratr.log")

	f, err := SetupLogFile(path)
	if err != nil {
		t.Fatalf("SetupLogFile() error = %v", err)
	}
	defer f.Close()

	// Write something to verify the file works.
	_, err = f.WriteString("test log entry\n")
	if err != nil {
		t.Fatalf("writing to log file: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "test log entry") {
		t.Errorf("log file = %q, want it to contain 'test log entry'", string(data))
	}
}

func TestSetupLogFile_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "dir", "orchestratr.log")

	f, err := SetupLogFile(path)
	if err != nil {
		t.Fatalf("SetupLogFile() should create directories, got: %v", err)
	}
	f.Close()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("log file should exist after SetupLogFile")
	}
}

func TestDefaultLogPath(t *testing.T) {
	p := DefaultLogPath()
	if p == "" {
		t.Error("DefaultLogPath() should not be empty")
	}
	if !strings.Contains(p, "orchestratr") {
		t.Errorf("DefaultLogPath() = %q, want it to contain 'orchestratr'", p)
	}
}
