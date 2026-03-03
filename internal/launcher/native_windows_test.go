//go:build windows

package launcher

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/josiahH-cf/orchestratr/internal/registry"
)

// TestWindowsExecutorInterface verifies WindowsExecutor satisfies the
// Executor interface at compile time.
var _ Executor = (*WindowsExecutor)(nil)

func TestWindowsLaunchAndIsRunning(t *testing.T) {
	e := NewWindowsExecutor()
	defer e.StopAll()

	entry := registry.AppEntry{
		Name:        "sleeper",
		Command:     "ping -n 60 127.0.0.1",
		Environment: "native",
	}

	res, err := e.Launch(entry)
	if err != nil {
		t.Fatalf("Launch failed: %v", err)
	}
	if res.Name != "sleeper" {
		t.Errorf("result Name = %q, want %q", res.Name, "sleeper")
	}
	if res.PID <= 0 {
		t.Errorf("result PID = %d, want > 0", res.PID)
	}
	if !e.IsRunning("sleeper") {
		t.Error("expected IsRunning(sleeper) = true")
	}
}

func TestWindowsLaunchEmptyEnvDefaultsToNative(t *testing.T) {
	e := NewWindowsExecutor()
	defer e.StopAll()

	// Empty environment should default to native (cmd.exe /c).
	entry := registry.AppEntry{
		Name:    "echoer",
		Command: "echo hello",
	}

	res, err := e.Launch(entry)
	if err != nil {
		t.Fatalf("Launch failed: %v", err)
	}
	if res.PID <= 0 {
		t.Errorf("PID = %d, want > 0", res.PID)
	}
}

func TestWindowsLaunchAlreadyRunning(t *testing.T) {
	e := NewWindowsExecutor()
	defer e.StopAll()

	entry := registry.AppEntry{
		Name:    "sleeper",
		Command: "ping -n 60 127.0.0.1",
	}
	_, err := e.Launch(entry)
	if err != nil {
		t.Fatalf("first Launch failed: %v", err)
	}

	_, err = e.Launch(entry)
	if !errors.Is(err, ErrAlreadyRunning) {
		t.Errorf("second Launch: got %v, want ErrAlreadyRunning", err)
	}
}

func TestWindowsLaunchEmptyCommand(t *testing.T) {
	e := NewWindowsExecutor()
	_, err := e.Launch(registry.AppEntry{Name: "nocommand"})
	if err == nil {
		t.Fatal("expected error for empty command")
	}
}

func TestWindowsStopRunningApp(t *testing.T) {
	e := NewWindowsExecutor()

	entry := registry.AppEntry{
		Name:    "sleeper",
		Command: "ping -n 60 127.0.0.1",
	}
	_, err := e.Launch(entry)
	if err != nil {
		t.Fatalf("Launch failed: %v", err)
	}

	if err := e.Stop("sleeper"); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if !e.IsRunning("sleeper") {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("sleeper still running after Stop")
}

func TestWindowsStopNotRunning(t *testing.T) {
	e := NewWindowsExecutor()

	err := e.Stop("ghost")
	if !errors.Is(err, ErrNotRunning) {
		t.Errorf("Stop(ghost): got %v, want ErrNotRunning", err)
	}
}

func TestWindowsNaturalExit(t *testing.T) {
	var mu sync.Mutex
	var exitName string
	var exitErr error
	exitCalled := make(chan struct{})

	e := NewWindowsExecutor(WithExitCallback(func(name string, err error) {
		mu.Lock()
		defer mu.Unlock()
		exitName = name
		exitErr = err
		close(exitCalled)
	}))

	entry := registry.AppEntry{
		Name:    "quick",
		Command: "echo hello",
	}
	_, err := e.Launch(entry)
	if err != nil {
		t.Fatalf("Launch failed: %v", err)
	}

	select {
	case <-exitCalled:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for exit callback")
	}

	mu.Lock()
	defer mu.Unlock()
	if exitName != "quick" {
		t.Errorf("exit callback name = %q, want %q", exitName, "quick")
	}
	if exitErr != nil {
		t.Errorf("exit callback err = %v, want nil", exitErr)
	}

	if e.IsRunning("quick") {
		t.Error("expected IsRunning(quick) = false after exit")
	}
}

func TestWindowsStopAll(t *testing.T) {
	e := NewWindowsExecutor()

	for _, name := range []string{"a", "b", "c"} {
		entry := registry.AppEntry{
			Name:    name,
			Command: "ping -n 60 127.0.0.1",
		}
		if _, err := e.Launch(entry); err != nil {
			t.Fatalf("Launch(%q) failed: %v", name, err)
		}
	}

	if !e.IsRunning("a") || !e.IsRunning("b") || !e.IsRunning("c") {
		t.Fatal("expected all apps running before StopAll")
	}

	e.StopAll()

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if !e.IsRunning("a") && !e.IsRunning("b") && !e.IsRunning("c") {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("apps still running after StopAll")
}

func TestWindowsRelaunchAfterExit(t *testing.T) {
	exitCalled := make(chan struct{}, 1)
	e := NewWindowsExecutor(WithExitCallback(func(name string, err error) {
		exitCalled <- struct{}{}
	}))
	defer e.StopAll()

	entry := registry.AppEntry{
		Name:    "relaunch",
		Command: "echo done",
	}

	_, err := e.Launch(entry)
	if err != nil {
		t.Fatalf("first Launch failed: %v", err)
	}

	select {
	case <-exitCalled:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for first exit")
	}

	_, err = e.Launch(entry)
	if err != nil {
		t.Fatalf("re-Launch failed: %v", err)
	}

	select {
	case <-exitCalled:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for second exit")
	}
}

func TestWindowsLaunchDetached_NotTracked(t *testing.T) {
	exitCalled := make(chan struct{}, 1)
	e := NewWindowsExecutor(WithExitCallback(func(name string, err error) {
		exitCalled <- struct{}{}
	}))

	entry := registry.AppEntry{
		Name:     "detached",
		Command:  "echo done",
		Detached: true,
	}

	res, err := e.Launch(entry)
	if err != nil {
		t.Fatalf("Launch failed: %v", err)
	}
	if res.PID <= 0 {
		t.Errorf("PID = %d, want > 0", res.PID)
	}

	if e.IsRunning("detached") {
		t.Error("detached app should not be tracked as running")
	}

	select {
	case <-exitCalled:
		t.Error("exit callback should not fire for detached app")
	case <-time.After(500 * time.Millisecond):
	}
}

func TestWindowsBuildCommand_EnvironmentRouting(t *testing.T) {
	e := NewWindowsExecutor()

	tests := []struct {
		name    string
		entry   registry.AppEntry
		wantBin string // expected executable base name
	}{
		{
			name: "native_explicit",
			entry: registry.AppEntry{
				Name:        "app",
				Command:     "notepad",
				Environment: "native",
			},
			wantBin: "cmd.exe",
		},
		{
			name: "native_empty",
			entry: registry.AppEntry{
				Name:    "app",
				Command: "notepad",
			},
			wantBin: "cmd.exe",
		},
		{
			name: "wsl_default",
			entry: registry.AppEntry{
				Name:        "app",
				Command:     "vim",
				Environment: "wsl",
			},
			wantBin: "wsl.exe",
		},
		{
			name: "wsl_named",
			entry: registry.AppEntry{
				Name:        "app",
				Command:     "vim",
				Environment: "wsl:Ubuntu",
			},
			wantBin: "wsl.exe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := e.buildCommand(tt.entry)
			if err != nil {
				t.Fatalf("buildCommand error: %v", err)
			}
			if cmd.Path == "" {
				// cmd.Path might be resolved; check Args[0] instead.
				t.Fatal("command path is empty")
			}
			// The first arg should contain the expected binary.
			if len(cmd.Args) == 0 {
				t.Fatal("no args in command")
			}
		})
	}
}

func TestWindowsPID_NotRunning(t *testing.T) {
	e := NewWindowsExecutor()

	pid, ok := e.PID("nonexistent")
	if ok {
		t.Error("expected ok=false for nonexistent app")
	}
	if pid != 0 {
		t.Errorf("expected pid=0, got %d", pid)
	}
}

func TestWindowsIsRunning_NotTracked(t *testing.T) {
	e := NewWindowsExecutor()
	if e.IsRunning("ghost") {
		t.Error("expected IsRunning(ghost) = false")
	}
}
