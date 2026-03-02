//go:build linux

package launcher

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"github.com/josiahH-cf/orchestratr/internal/registry"
)

// tracked holds the state for a single running process.
type tracked struct {
	cmd  *exec.Cmd
	name string
	pid  int
}

// NativeExecutor launches apps as child processes on the local host
// using bash -c to support shell features (pipes, env vars, etc.).
type NativeExecutor struct {
	mu       sync.Mutex
	procs    map[string]*tracked
	onExit   ExitCallback
	logger   *log.Logger
	shellCmd string // the shell binary, default "bash"
}

// NewPlatformExecutor creates the platform-appropriate Executor. On
// Linux this returns a NativeExecutor.
func NewPlatformExecutor(opts ...Option) Executor {
	return NewNativeExecutor(opts...)
}

// NewNativeExecutor creates a NativeExecutor with the given options.
func NewNativeExecutor(opts ...Option) *NativeExecutor {
	o := buildOptions(opts)
	e := &NativeExecutor{
		procs:    make(map[string]*tracked),
		logger:   o.Logger,
		onExit:   o.OnExit,
		shellCmd: o.Shell,
	}
	if e.logger == nil {
		e.logger = log.New(os.Stderr, "launcher: ", log.LstdFlags)
	}
	return e
}

// Launch starts the app described by entry. It returns
// ErrAlreadyRunning if the app already has a tracked process.
// When entry.Detached is true, the process is started but not
// tracked — no exit callback will fire (OI-5).
func (e *NativeExecutor) Launch(entry registry.AppEntry) (*Result, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, ok := e.procs[entry.Name]; ok {
		return nil, ErrAlreadyRunning
	}

	if entry.Command == "" {
		return nil, fmt.Errorf("app %q has no command", entry.Name)
	}

	cmd := exec.Command(e.shellCmd, "-c", entry.Command)
	// Start in a new process group so we can signal the whole group.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("launching %q: %w", entry.Name, err)
	}

	t := &tracked{
		cmd:  cmd,
		name: entry.Name,
		pid:  cmd.Process.Pid,
	}

	if entry.Detached {
		// Don't track or wait — the process is fire-and-forget.
		// Release so we don't leak a zombie.
		_ = cmd.Process.Release()
		return &Result{Name: entry.Name, PID: t.pid}, nil
	}

	e.procs[entry.Name] = t

	// Monitor the process in a goroutine — no polling, uses waitpid.
	go e.waitForExit(t)

	return &Result{
		Name: entry.Name,
		PID:  t.pid,
	}, nil
}

// waitForExit blocks until the process exits, then updates tracking
// and fires the exit callback.
func (e *NativeExecutor) waitForExit(t *tracked) {
	waitErr := t.cmd.Wait()

	e.mu.Lock()
	// Only remove if it's still the same process (not re-launched).
	if cur, ok := e.procs[t.name]; ok && cur.pid == t.pid {
		delete(e.procs, t.name)
	}
	cb := e.onExit
	e.mu.Unlock()

	if waitErr != nil {
		e.logger.Printf("app %q (PID %d) exited with error: %v", t.name, t.pid, waitErr)
	} else {
		e.logger.Printf("app %q (PID %d) exited normally", t.name, t.pid)
	}

	if cb != nil {
		cb(t.name, waitErr)
	}
}

// Stop sends SIGTERM to the app's process group. Returns
// ErrNotRunning if the app is not tracked.
func (e *NativeExecutor) Stop(name string) error {
	e.mu.Lock()
	t, ok := e.procs[name]
	e.mu.Unlock()

	if !ok {
		return ErrNotRunning
	}

	// Signal the whole process group (negative PID).
	if err := syscall.Kill(-t.pid, syscall.SIGTERM); err != nil {
		return fmt.Errorf("sending SIGTERM to %q (PID %d): %w", name, t.pid, err)
	}
	return nil
}

// StopAll sends SIGTERM to all tracked processes. Errors are logged
// but do not prevent stopping remaining processes.
func (e *NativeExecutor) StopAll() {
	e.mu.Lock()
	names := make([]string, 0, len(e.procs))
	for name := range e.procs {
		names = append(names, name)
	}
	e.mu.Unlock()

	for _, name := range names {
		if err := e.Stop(name); err != nil {
			e.logger.Printf("warning: stopping %q: %v", name, err)
		}
	}
}

// IsRunning reports whether the named app has a tracked running process.
func (e *NativeExecutor) IsRunning(name string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	_, ok := e.procs[name]
	return ok
}

// PID returns the OS process ID for the named app. Returns 0 and
// false if the app is not tracked as running.
func (e *NativeExecutor) PID(name string) (int, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	t, ok := e.procs[name]
	if !ok {
		return 0, false
	}
	return t.pid, true
}
