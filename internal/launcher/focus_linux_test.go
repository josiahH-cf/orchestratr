//go:build linux

package launcher

import (
	"os/exec"
	"testing"

	"github.com/josiahH-cf/orchestratr/internal/registry"
)

func TestFocusWindow_PIDZero(t *testing.T) {
	err := FocusWindow(0)
	if err == nil {
		t.Error("expected error for PID 0")
	}
}

func TestFocusWindow_NegativePID(t *testing.T) {
	err := FocusWindow(-1)
	if err == nil {
		t.Error("expected error for negative PID")
	}
}

func TestFocusWindow_NonexistentPID(t *testing.T) {
	// PID 999999999 should not have a window.
	err := FocusWindow(999999999)
	if err == nil {
		t.Error("expected error for nonexistent PID")
	}
}

func TestFocusWindow_NoXdotool(t *testing.T) {
	// Skip if xdotool is actually available — this tests the
	// not-installed path.
	if _, err := exec.LookPath("xdotool"); err == nil {
		t.Skip("xdotool is available; skipping no-xdotool test")
	}
	err := FocusWindow(1)
	if err == nil {
		t.Error("expected error when xdotool is not available")
	}
}

func TestPIDRunningApp(t *testing.T) {
	e := NewNativeExecutor()
	defer e.StopAll()

	entry := registryEntry("pidtest", "sleep 60")
	_, err := e.Launch(entry)
	if err != nil {
		t.Fatalf("Launch failed: %v", err)
	}

	pid, ok := e.PID("pidtest")
	if !ok {
		t.Error("expected PID to be found for running app")
	}
	if pid <= 0 {
		t.Errorf("expected PID > 0, got %d", pid)
	}
}

func TestPIDNotRunning(t *testing.T) {
	e := NewNativeExecutor()

	pid, ok := e.PID("ghost")
	if ok {
		t.Error("expected ok=false for non-running app")
	}
	if pid != 0 {
		t.Errorf("expected PID=0, got %d", pid)
	}
}

// registryEntry is a test helper to build an AppEntry.
func registryEntry(name, command string) registry.AppEntry {
	return registry.AppEntry{
		Name:        name,
		Command:     command,
		Environment: "native",
	}
}
