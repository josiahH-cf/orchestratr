package launcher

import (
	"context"
	"log"
	"os/exec"
	"time"

	"github.com/josiahH-cf/orchestratr/internal/api"
	"github.com/josiahH-cf/orchestratr/internal/registry"
)

// defaultReadyTimeout is used when ReadyTimeoutMs is zero or negative.
const defaultReadyTimeout = 5 * time.Second

// pollInterval is the time between readiness probe attempts.
const pollInterval = 500 * time.Millisecond

// cmdTimeout is the maximum duration for a single ready_cmd execution.
const cmdTimeout = 5 * time.Second

// PollReadiness checks an app's readiness after launch. If ready_cmd
// is empty the app transitions to ready immediately. Otherwise the
// command is polled at 500ms intervals until it exits 0, the timeout
// elapses, the process exits, or the context is cancelled. On timeout
// or process exit the app remains in "launched" state.
func PollReadiness(ctx context.Context, entry registry.AppEntry, tracker *api.StateTracker, executor Executor, logger *log.Logger) {
	if entry.ReadyCmd == "" {
		tracker.SetReady(entry.Name)
		return
	}

	timeout := time.Duration(entry.ReadyTimeoutMs) * time.Millisecond
	if timeout <= 0 {
		timeout = defaultReadyTimeout
	}

	deadline := time.NewTimer(timeout)
	defer deadline.Stop()

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Printf("readiness poll for %q cancelled", entry.Name)
			return
		case <-deadline.C:
			logger.Printf("WARN: %s did not become ready within %v", entry.Name, timeout)
			return
		case <-ticker.C:
			if !executor.IsRunning(entry.Name) {
				logger.Printf("readiness poll for %q: process exited before ready", entry.Name)
				return
			}
			if runReadyCmd(ctx, entry.ReadyCmd) == nil {
				tracker.SetReady(entry.Name)
				logger.Printf("app %q is ready", entry.Name)
				return
			}
		}
	}
}

// runReadyCmd executes the ready_cmd and returns nil on exit code 0.
func runReadyCmd(ctx context.Context, readyCmd string) error {
	cmdCtx, cancel := context.WithTimeout(ctx, cmdTimeout)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "bash", "-c", readyCmd)
	// Discard stdout/stderr — only exit code matters.
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}
