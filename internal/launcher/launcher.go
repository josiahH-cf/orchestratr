// Package launcher provides process spawning, PID tracking, and
// lifecycle management for registered apps. Platform-specific
// implementations satisfy the Executor interface.
package launcher

import (
	"errors"
	"log"

	"github.com/josiahH-cf/orchestratr/internal/registry"
)

// ErrAlreadyRunning is returned when Launch is called for an app
// that already has a running process.
var ErrAlreadyRunning = errors.New("app is already running")

// ErrNotRunning is returned when Stop is called for an app that
// is not currently running.
var ErrNotRunning = errors.New("app is not running")

// ErrNotImplemented is returned by stub executors on platforms
// where launching is not yet supported.
var ErrNotImplemented = errors.New("launcher not implemented on this platform")

// ErrFocusNotSupported is returned when bring-to-front is not
// available on the current platform or display server.
var ErrFocusNotSupported = errors.New("window focus not supported")

// ExitCallback is called when a tracked process exits. The name
// is the app name and err is the process exit error (nil on
// success, non-nil on non-zero exit or signal).
type ExitCallback func(name string, err error)

// Result holds the outcome of a successful Launch call.
type Result struct {
	// Name is the app name from the registry entry.
	Name string
	// PID is the OS process ID of the spawned process.
	PID int
}

// Executor launches and tracks app processes. Implementations are
// platform-specific.
type Executor interface {
	// Launch starts the app described by entry. Returns
	// ErrAlreadyRunning if the app is already tracked as running.
	// The ExitCallback (if set) is called asynchronously when the
	// process exits.
	Launch(entry registry.AppEntry) (*Result, error)

	// Stop sends a termination signal to the named app's process.
	// Returns ErrNotRunning if the app is not tracked.
	Stop(name string) error

	// StopAll terminates all tracked processes. Best-effort: errors
	// are logged but do not prevent stopping remaining processes.
	StopAll()

	// IsRunning reports whether the named app has a tracked running
	// process.
	IsRunning(name string) bool

	// PID returns the OS process ID for the named app. Returns 0 and
	// false if the app is not tracked as running.
	PID(name string) (int, bool)
}

// Options holds configuration for the platform executor.
type Options struct {
	OnExit ExitCallback
	Logger *log.Logger
	Shell  string // override the default shell binary
}

// Option configures a platform executor.
type Option func(*Options)

// WithExitCallback sets the function called when a tracked process
// exits.
func WithExitCallback(fn ExitCallback) Option {
	return func(o *Options) { o.OnExit = fn }
}

// WithLogger sets the executor's logger.
func WithLogger(l *log.Logger) Option {
	return func(o *Options) { o.Logger = l }
}

// WithShell overrides the default shell binary ("bash"). Useful for
// testing with a known shell path.
func WithShell(shell string) Option {
	return func(o *Options) { o.Shell = shell }
}

// buildOptions applies option functions to a default Options value.
func buildOptions(opts []Option) Options {
	o := Options{Shell: "bash"}
	for _, fn := range opts {
		fn(&o)
	}
	return o
}
