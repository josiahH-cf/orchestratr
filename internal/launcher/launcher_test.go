//go:build linux

package launcher

import (
	"testing"

	"github.com/josiahH-cf/orchestratr/internal/registry"
)

// TestExecutorInterface verifies that both NativeExecutor (on linux)
// and StubExecutor satisfy the Executor interface at compile time.
var _ Executor = (*NativeExecutor)(nil)

func TestErrSentinels(t *testing.T) {
	// Ensure sentinel errors have useful messages.
	tests := []struct {
		err  error
		want string
	}{
		{ErrAlreadyRunning, "app is already running"},
		{ErrNotRunning, "app is not running"},
		{ErrNotImplemented, "launcher not implemented on this platform"},
	}
	for _, tt := range tests {
		if tt.err.Error() != tt.want {
			t.Errorf("got %q, want %q", tt.err.Error(), tt.want)
		}
	}
}

func TestResultFields(t *testing.T) {
	r := &Result{Name: "myapp", PID: 42}
	if r.Name != "myapp" {
		t.Errorf("Name = %q, want %q", r.Name, "myapp")
	}
	if r.PID != 42 {
		t.Errorf("PID = %d, want %d", r.PID, 42)
	}
}

func TestLaunchEmptyCommand(t *testing.T) {
	e := NewNativeExecutor()
	_, err := e.Launch(registry.AppEntry{Name: "nocommand"})
	if err == nil {
		t.Fatal("expected error for empty command")
	}
}
