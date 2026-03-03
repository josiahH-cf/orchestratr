//go:build linux

package launcher

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/josiahH-cf/orchestratr/internal/api"
	"github.com/josiahH-cf/orchestratr/internal/registry"
)

func TestPollReadiness_EmptyReadyCmd(t *testing.T) {
	// When ready_cmd is empty, app should transition to ready immediately.
	tracker := api.NewStateTracker()
	tracker.SetLaunched("myapp")

	entry := registry.AppEntry{
		Name:    "myapp",
		Command: "sleep 60",
	}

	exec := NewNativeExecutor()
	logger := log.New(os.Stderr, "test: ", log.LstdFlags)

	PollReadiness(context.Background(), entry, tracker, exec, logger)

	state := tracker.Get("myapp")
	if state == nil {
		t.Fatal("expected state for myapp")
	}
	if !state.Ready {
		t.Error("expected Ready = true when ready_cmd is empty")
	}
	if state.ReadyAt == nil {
		t.Error("expected ReadyAt to be set")
	}
}

func TestPollReadiness_ReadyCmdSuccess(t *testing.T) {
	// ready_cmd exits 0 → app becomes ready.
	tracker := api.NewStateTracker()
	tracker.SetLaunched("myapp")

	exec := NewNativeExecutor()
	defer exec.StopAll()

	// Launch a real process so IsRunning returns true.
	entry := registry.AppEntry{
		Name:        "myapp",
		Command:     "sleep 60",
		ReadyCmd:    "true", // exits 0 immediately
		Environment: "native",
	}
	_, err := exec.Launch(entry)
	if err != nil {
		t.Fatalf("Launch failed: %v", err)
	}

	logger := log.New(os.Stderr, "test: ", log.LstdFlags)

	PollReadiness(context.Background(), entry, tracker, exec, logger)

	state := tracker.Get("myapp")
	if state == nil {
		t.Fatal("expected state for myapp")
	}
	if !state.Ready {
		t.Error("expected Ready = true after ready_cmd exits 0")
	}
}

func TestPollReadiness_ReadyCmdTimeout(t *testing.T) {
	// ready_cmd always fails → times out, stays in launched state.
	tracker := api.NewStateTracker()
	tracker.SetLaunched("myapp")

	exec := NewNativeExecutor()
	defer exec.StopAll()

	entry := registry.AppEntry{
		Name:           "myapp",
		Command:        "sleep 60",
		ReadyCmd:       "false", // always exits non-zero
		ReadyTimeoutMs: 600,     // short timeout for test speed
		Environment:    "native",
	}
	_, err := exec.Launch(entry)
	if err != nil {
		t.Fatalf("Launch failed: %v", err)
	}

	logger := log.New(os.Stderr, "test: ", log.LstdFlags)

	start := time.Now()
	PollReadiness(context.Background(), entry, tracker, exec, logger)
	elapsed := time.Since(start)

	state := tracker.Get("myapp")
	if state == nil {
		t.Fatal("expected state for myapp")
	}
	if state.Ready {
		t.Error("expected Ready = false after timeout")
	}
	// Should have taken at least the timeout duration.
	if elapsed < 500*time.Millisecond {
		t.Errorf("polling ended too soon: %v (expected >= 500ms)", elapsed)
	}
}

func TestPollReadiness_ProcessExitsBeforeReady(t *testing.T) {
	// Process exits before ready_cmd succeeds → polling stops.
	tracker := api.NewStateTracker()
	tracker.SetLaunched("myapp")

	exec := NewNativeExecutor()
	defer exec.StopAll()

	// Launch a short-lived process.
	entry := registry.AppEntry{
		Name:           "myapp",
		Command:        "sleep 0.1",
		ReadyCmd:       "false", // always fails
		ReadyTimeoutMs: 10000,   // long timeout
		Environment:    "native",
	}
	_, err := exec.Launch(entry)
	if err != nil {
		t.Fatalf("Launch failed: %v", err)
	}

	logger := log.New(os.Stderr, "test: ", log.LstdFlags)

	// Wait for the process to exit.
	waitFor(t, 3*time.Second, "process to exit", func() bool {
		return !exec.IsRunning("myapp")
	})

	PollReadiness(context.Background(), entry, tracker, exec, logger)

	state := tracker.Get("myapp")
	if state == nil {
		t.Fatal("expected state for myapp")
	}
	if state.Ready {
		t.Error("expected Ready = false when process exited before ready")
	}
}

func TestPollReadiness_ContextCancellation(t *testing.T) {
	// Context cancelled → polling stops promptly.
	tracker := api.NewStateTracker()
	tracker.SetLaunched("myapp")

	exec := NewNativeExecutor()
	defer exec.StopAll()

	entry := registry.AppEntry{
		Name:           "myapp",
		Command:        "sleep 60",
		ReadyCmd:       "false", // always fails
		ReadyTimeoutMs: 30000,   // very long timeout
		Environment:    "native",
	}
	_, err := exec.Launch(entry)
	if err != nil {
		t.Fatalf("Launch failed: %v", err)
	}

	logger := log.New(os.Stderr, "test: ", log.LstdFlags)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		PollReadiness(ctx, entry, tracker, exec, logger)
		close(done)
	}()

	// Cancel after a short delay.
	time.Sleep(200 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// OK — polling stopped.
	case <-time.After(3 * time.Second):
		t.Fatal("PollReadiness did not return after context cancellation")
	}

	state := tracker.Get("myapp")
	if state == nil {
		t.Fatal("expected state for myapp")
	}
	if state.Ready {
		t.Error("expected Ready = false after context cancellation")
	}
}

func TestPollReadiness_DefaultTimeout(t *testing.T) {
	// When ReadyTimeoutMs is 0, default of 5000ms should be used.
	// We just verify it doesn't crash or use an invalid value.
	tracker := api.NewStateTracker()
	tracker.SetLaunched("myapp")

	exec := NewNativeExecutor()
	defer exec.StopAll()

	entry := registry.AppEntry{
		Name:           "myapp",
		Command:        "sleep 60",
		ReadyCmd:       "true", // exits 0 immediately
		ReadyTimeoutMs: 0,      // should use default
		Environment:    "native",
	}
	_, err := exec.Launch(entry)
	if err != nil {
		t.Fatalf("Launch failed: %v", err)
	}

	logger := log.New(os.Stderr, "test: ", log.LstdFlags)

	PollReadiness(context.Background(), entry, tracker, exec, logger)

	state := tracker.Get("myapp")
	if state == nil {
		t.Fatal("expected state for myapp")
	}
	if !state.Ready {
		t.Error("expected Ready = true")
	}
}

func TestPollReadiness_NegativeTimeout(t *testing.T) {
	// Negative ReadyTimeoutMs should use default.
	tracker := api.NewStateTracker()
	tracker.SetLaunched("myapp")

	exec := NewNativeExecutor()
	defer exec.StopAll()

	entry := registry.AppEntry{
		Name:           "myapp",
		Command:        "sleep 60",
		ReadyCmd:       "true",
		ReadyTimeoutMs: -100,
		Environment:    "native",
	}
	_, err := exec.Launch(entry)
	if err != nil {
		t.Fatalf("Launch failed: %v", err)
	}

	logger := log.New(os.Stderr, "test: ", log.LstdFlags)

	PollReadiness(context.Background(), entry, tracker, exec, logger)

	state := tracker.Get("myapp")
	if state == nil {
		t.Fatal("expected state for myapp")
	}
	if !state.Ready {
		t.Error("expected Ready = true")
	}
}
