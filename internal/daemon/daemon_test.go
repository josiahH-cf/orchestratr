package daemon

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestDaemon_StartStop(t *testing.T) {
	d := New(Config{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)

	var runErr error
	go func() {
		defer wg.Done()
		runErr = d.Run(ctx)
	}()

	// Give the daemon a moment to start.
	time.Sleep(50 * time.Millisecond)

	if d.State() != StateRunning {
		t.Errorf("State() = %v, want %v", d.State(), StateRunning)
	}

	cancel()
	wg.Wait()

	if runErr != nil {
		t.Errorf("Run() error = %v", runErr)
	}

	if d.State() != StateStopped {
		t.Errorf("State() = %v, want %v after stop", d.State(), StateStopped)
	}
}

func TestDaemon_PauseResume(t *testing.T) {
	d := New(Config{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = d.Run(ctx)
	}()

	time.Sleep(50 * time.Millisecond)

	// Pause.
	if err := d.Pause(); err != nil {
		t.Fatalf("Pause() error = %v", err)
	}
	if d.State() != StatePaused {
		t.Errorf("State() = %v, want %v", d.State(), StatePaused)
	}

	// Resume.
	if err := d.Resume(); err != nil {
		t.Fatalf("Resume() error = %v", err)
	}
	if d.State() != StateRunning {
		t.Errorf("State() = %v, want %v", d.State(), StateRunning)
	}

	cancel()
	wg.Wait()
}

func TestDaemon_PauseWhenNotRunning(t *testing.T) {
	d := New(Config{})
	if err := d.Pause(); err == nil {
		t.Error("Pause() should error when daemon is not running")
	}
}

func TestDaemon_ResumeWhenNotPaused(t *testing.T) {
	d := New(Config{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = d.Run(ctx)
	}()

	time.Sleep(50 * time.Millisecond)

	if err := d.Resume(); err == nil {
		t.Error("Resume() should error when daemon is not paused")
	}

	cancel()
	wg.Wait()
}

func TestDaemon_StatesAreCorrect(t *testing.T) {
	tests := []struct {
		state State
		want  string
	}{
		{StateStopped, "stopped"},
		{StateRunning, "running"},
		{StatePaused, "paused"},
	}
	for _, tt := range tests {
		if got := tt.state.String(); got != tt.want {
			t.Errorf("State(%d).String() = %q, want %q", tt.state, got, tt.want)
		}
	}
}
