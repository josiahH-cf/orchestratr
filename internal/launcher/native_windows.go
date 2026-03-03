//go:build windows

package launcher

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"

	"github.com/josiahH-cf/orchestratr/internal/registry"
)

// WindowsExecutor launches apps as child processes on Windows. It
// routes between native Windows apps (cmd.exe /c) and WSL2 apps
// (wsl.exe -d <distro>) based on the AppEntry.Environment field.
type WindowsExecutor struct {
	mu       sync.Mutex
	procs    map[string]*tracked
	onExit   ExitCallback
	logger   *log.Logger
	shellCmd string // override for native shell, default "cmd.exe"

	// cachedDefaultDistro is resolved once from `wsl --list --quiet`
	// and cached for the lifetime of the executor.
	cachedDefaultDistro string
	distroResolved      bool
}

// NewPlatformExecutor creates the platform-appropriate Executor. On
// Windows this returns a WindowsExecutor.
func NewPlatformExecutor(opts ...Option) Executor {
	return NewWindowsExecutor(opts...)
}

// NewWindowsExecutor creates a WindowsExecutor with the given options.
func NewWindowsExecutor(opts ...Option) *WindowsExecutor {
	o := buildOptions(opts)
	e := &WindowsExecutor{
		procs:    make(map[string]*tracked),
		logger:   o.Logger,
		onExit:   o.OnExit,
		shellCmd: o.Shell,
	}
	if e.shellCmd == "bash" {
		// buildOptions defaults to "bash", but on Windows the native
		// shell is cmd.exe.
		e.shellCmd = "cmd.exe"
	}
	if e.logger == nil {
		e.logger = log.New(os.Stderr, "launcher: ", log.LstdFlags)
	}
	return e
}

// Launch starts the app described by entry. It returns
// ErrAlreadyRunning if the app already has a tracked process.
// Environment routing:
//   - "" or "native": cmd.exe /c <command>
//   - "wsl":          wsl.exe -d <default> -- bash -lc <command>
//   - "wsl:<distro>": wsl.exe -d <distro> -- bash -lc <command>
func (e *WindowsExecutor) Launch(entry registry.AppEntry) (*Result, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, ok := e.procs[entry.Name]; ok {
		return nil, ErrAlreadyRunning
	}

	if entry.Command == "" {
		return nil, fmt.Errorf("app %q has no command", entry.Name)
	}

	cmd, err := e.buildCommand(entry)
	if err != nil {
		return nil, fmt.Errorf("building command for %q: %w", entry.Name, err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("launching %q: %w", entry.Name, err)
	}

	t := &tracked{
		cmd:  cmd,
		name: entry.Name,
		pid:  cmd.Process.Pid,
	}

	if entry.Detached {
		// Don't track or wait — fire-and-forget.
		_ = cmd.Process.Release()
		return &Result{Name: entry.Name, PID: t.pid}, nil
	}

	e.procs[entry.Name] = t

	// Monitor the process in a goroutine.
	go e.waitForExit(t)

	return &Result{
		Name: entry.Name,
		PID:  t.pid,
	}, nil
}

// buildCommand creates the exec.Cmd for the given entry based on its
// Environment field.
func (e *WindowsExecutor) buildCommand(entry registry.AppEntry) (*exec.Cmd, error) {
	env := strings.TrimSpace(entry.Environment)

	switch {
	case env == "" || env == "native":
		cmd := exec.Command(e.shellCmd, "/c", entry.Command)
		cmd.SysProcAttr = &syscall.SysProcAttr{
			CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
		}
		return cmd, nil

	case env == "wsl":
		distro, err := e.resolveDefaultWSLDistro()
		if err != nil {
			return nil, fmt.Errorf("resolving default WSL distro: %w", err)
		}
		cmd := exec.Command("wsl.exe", "-d", distro, "--", "bash", "-lc", entry.Command)
		cmd.SysProcAttr = &syscall.SysProcAttr{
			CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
		}
		return cmd, nil

	case strings.HasPrefix(env, "wsl:"):
		distro := strings.TrimPrefix(env, "wsl:")
		if distro == "" {
			return nil, fmt.Errorf("empty distro name in environment %q", env)
		}
		cmd := exec.Command("wsl.exe", "-d", distro, "--", "bash", "-lc", entry.Command)
		cmd.SysProcAttr = &syscall.SysProcAttr{
			CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
		}
		return cmd, nil

	default:
		return nil, fmt.Errorf("unsupported environment %q", env)
	}
}

// resolveDefaultWSLDistro runs `wsl.exe --list --quiet` and returns
// the first non-empty line (the default distribution). The result is
// cached for the lifetime of the executor.
func (e *WindowsExecutor) resolveDefaultWSLDistro() (string, error) {
	if e.distroResolved {
		if e.cachedDefaultDistro == "" {
			return "", fmt.Errorf("no WSL distributions found (cached)")
		}
		return e.cachedDefaultDistro, nil
	}

	out, err := exec.Command("wsl.exe", "--list", "--quiet").Output()
	if err != nil {
		e.distroResolved = true
		return "", fmt.Errorf("running wsl --list --quiet: %w", err)
	}

	// wsl.exe may output UTF-16LE on some systems; handle both.
	text := cleanWSLOutput(string(out))
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			e.cachedDefaultDistro = line
			e.distroResolved = true
			return line, nil
		}
	}

	e.distroResolved = true
	return "", fmt.Errorf("no WSL distributions found")
}

// cleanWSLOutput strips null bytes and carriage returns from wsl.exe
// output, handling potential UTF-16LE encoding.
func cleanWSLOutput(s string) string {
	s = strings.ReplaceAll(s, "\x00", "")
	s = strings.ReplaceAll(s, "\r", "")
	return s
}

// waitForExit blocks until the process exits, then updates tracking
// and fires the exit callback.
func (e *WindowsExecutor) waitForExit(t *tracked) {
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

// Stop terminates the named app's process. On Windows, we use
// Process.Kill() since there is no SIGTERM equivalent.
func (e *WindowsExecutor) Stop(name string) error {
	e.mu.Lock()
	t, ok := e.procs[name]
	e.mu.Unlock()

	if !ok {
		return ErrNotRunning
	}

	if err := t.cmd.Process.Kill(); err != nil {
		return fmt.Errorf("killing %q (PID %d): %w", name, t.pid, err)
	}
	return nil
}

// StopAll terminates all tracked processes. Errors are logged but do
// not prevent stopping remaining processes.
func (e *WindowsExecutor) StopAll() {
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
func (e *WindowsExecutor) IsRunning(name string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	_, ok := e.procs[name]
	return ok
}

// PID returns the OS process ID for the named app. Returns 0 and
// false if the app is not tracked as running.
func (e *WindowsExecutor) PID(name string) (int, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	t, ok := e.procs[name]
	if !ok {
		return 0, false
	}
	return t.pid, true
}
