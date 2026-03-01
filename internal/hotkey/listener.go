package hotkey

import "errors"

// ErrNotImplemented is returned by listener stubs on platforms where
// the real hotkey capture API has not yet been implemented.
var ErrNotImplemented = errors.New("hotkey listener not implemented for this platform")

// KeyEvent represents a key press or release detected by a Listener.
type KeyEvent struct {
	Key     Key
	Pressed bool // true = key down, false = key up
}

// PlatformInfo describes the hotkey capture backend in use.
type PlatformInfo struct {
	OS     string // "linux", "darwin", "windows", "unknown"
	Method string // e.g. "x11_xgrabkey", "wayland_portal", "cgeventtap", "registerhotkey", "stub"
}

// String returns a diagnostic string like "linux/x11_xgrabkey".
func (p PlatformInfo) String() string {
	return p.OS + "/" + p.Method
}

// Listener captures global keyboard events at the OS level.
// Each platform provides its own implementation behind build tags.
type Listener interface {
	// Info returns the platform and method this listener uses.
	Info() PlatformInfo

	// Register sets the leader key for global capture. This must be
	// called before Start. It returns a warning string if the key is
	// known to conflict with common shortcuts (empty if no conflict).
	Register(leader Key) (warning string, err error)

	// Start begins listening for key events. Events are sent to the
	// provided channel. Start blocks until Stop is called or an error
	// occurs. The caller owns the channel and must ensure it is read.
	Start(events chan<- KeyEvent) error

	// Stop releases the global hotkey registration and stops listening.
	// It is safe to call Stop multiple times.
	Stop() error
}
