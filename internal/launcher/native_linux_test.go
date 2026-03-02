//go:build linux

package launcher

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/josiahH-cf/orchestratr/internal/registry"
)

// helper waits for a condition to become true within a timeout.
func waitFor(t *testing.T, timeout time.Duration, desc string, fn func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if fn() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for: %s", desc)
}

func TestLaunchAndIsRunning(t *testing.T) {
	e := NewNativeExecutor()
	defer e.StopAll()

	entry := registry.AppEntry{
		Name:        "sleeper",
		Command:     "sleep 60",
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

func TestLaunchAlreadyRunning(t *testing.T) {
	e := NewNativeExecutor()
	defer e.StopAll()

	entry := registry.AppEntry{
		Name:    "sleeper",
		Command: "sleep 60",
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

func TestStopRunningApp(t *testing.T) {
	e := NewNativeExecutor()

	entry := registry.AppEntry{
		Name:    "sleeper",
		Command: "sleep 60",
	}
	_, err := e.Launch(entry)
	if err != nil {
		t.Fatalf("Launch failed: %v", err)
	}

	if err := e.Stop("sleeper"); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	waitFor(t, 3*time.Second, "sleeper to exit", func() bool {
		return !e.IsRunning("sleeper")
	})
}

func TestStopNotRunning(t *testing.T) {
	e := NewNativeExecutor()

	err := e.Stop("ghost")
	if !errors.Is(err, ErrNotRunning) {
		t.Errorf("Stop(ghost): got %v, want ErrNotRunning", err)
	}
}

func TestNaturalExit(t *testing.T) {
	var mu sync.Mutex
	var exitName string
	var exitErr error
	exitCalled := make(chan struct{})

	e := NewNativeExecutor(WithExitCallback(func(name string, err error) {
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

func TestNaturalExitWithError(t *testing.T) {
	var mu sync.Mutex
	var exitErr error
	exitCalled := make(chan struct{})

	e := NewNativeExecutor(WithExitCallback(func(name string, err error) {
		mu.Lock()
		defer mu.Unlock()
		exitErr = err
		close(exitCalled)
	}))

	entry := registry.AppEntry{
		Name:    "failing",
		Command: "exit 1",
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
	if exitErr == nil {
		t.Error("expected non-nil exit error for 'exit 1'")
	}
}

func TestLaunchBadCommand(t *testing.T) {
	e := NewNativeExecutor(WithShell("/nonexistent-shell"))

	entry := registry.AppEntry{
		Name:    "badcmd",
		Command: "whatever",
	}
	_, err := e.Launch(entry)
	if err == nil {
		t.Fatal("expected error for bad shell")
	}
}

func TestStopAll(t *testing.T) {
	e := NewNativeExecutor()

	for _, name := range []string{"a", "b", "c"} {
		entry := registry.AppEntry{
			Name:    name,
			Command: "sleep 60",
		}
		if _, err := e.Launch(entry); err != nil {
			t.Fatalf("Launch(%q) failed: %v", name, err)
		}
	}

	if !e.IsRunning("a") || !e.IsRunning("b") || !e.IsRunning("c") {
		t.Fatal("expected all apps running before StopAll")
	}

	e.StopAll()

	waitFor(t, 5*time.Second, "all apps to exit", func() bool {
		return !e.IsRunning("a") && !e.IsRunning("b") && !e.IsRunning("c")
	})
}

func TestRelaunchAfterExit(t *testing.T) {
	exitCalled := make(chan struct{}, 1)
	e := NewNativeExecutor(WithExitCallback(func(name string, err error) {
		exitCalled <- struct{}{}
	}))
	defer e.StopAll()

	entry := registry.AppEntry{
		Name:    "relaunch",
		Command: "echo done",
	}

	// First launch — should exit quickly.
	_, err := e.Launch(entry)
	if err != nil {
		t.Fatalf("first Launch failed: %v", err)
	}

	select {
	case <-exitCalled:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for first exit")
	}

	// Should be launchable again.
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

func TestLaunchDetached_NotTracked(t *testing.T) {
	exitCalled := make(chan struct{}, 1)
	e := NewNativeExecutor(WithExitCallback(func(name string, err error) {
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

	// Detached apps should NOT be tracked.
	if e.IsRunning("detached") {
		t.Error("detached app should not be tracked as running")
	}

	// Exit callback should NOT be called for detached apps.
	select {
	case <-exitCalled:
		t.Error("exit callback should not fire for detached app")
	case <-time.After(500 * time.Millisecond):
		// Expected: no callback.
	}
}
