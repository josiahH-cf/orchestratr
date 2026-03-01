package daemon

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestAcquireLock_Success(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "orchestratr.pid")

	lock, err := AcquireLock(path)
	if err != nil {
		t.Fatalf("AcquireLock() error = %v", err)
	}
	defer lock.Release()

	// PID file should contain our PID.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading PID file: %v", err)
	}
	pid, err := strconv.Atoi(string(data))
	if err != nil {
		t.Fatalf("parsing PID: %v", err)
	}
	if pid != os.Getpid() {
		t.Errorf("PID = %d, want %d", pid, os.Getpid())
	}
}

func TestAcquireLock_RejectsDuplicate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "orchestratr.pid")

	lock1, err := AcquireLock(path)
	if err != nil {
		t.Fatalf("first AcquireLock() error = %v", err)
	}
	defer lock1.Release()

	_, err = AcquireLock(path)
	if err == nil {
		t.Fatal("second AcquireLock() should fail when lock is held")
	}
}

func TestAcquireLock_RecoversStaleLock(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "orchestratr.pid")

	// Write a PID that doesn't exist (stale lock).
	if err := os.WriteFile(path, []byte("999999999"), 0o644); err != nil {
		t.Fatal(err)
	}

	lock, err := AcquireLock(path)
	if err != nil {
		t.Fatalf("AcquireLock() should recover stale lock, got: %v", err)
	}
	defer lock.Release()
}

func TestLock_Release(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "orchestratr.pid")

	lock, err := AcquireLock(path)
	if err != nil {
		t.Fatalf("AcquireLock() error = %v", err)
	}

	if err := lock.Release(); err != nil {
		t.Fatalf("Release() error = %v", err)
	}

	// PID file should be removed.
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("PID file should be removed after Release()")
	}

	// Should be able to re-acquire.
	lock2, err := AcquireLock(path)
	if err != nil {
		t.Fatalf("re-AcquireLock() error = %v", err)
	}
	defer lock2.Release()
}
