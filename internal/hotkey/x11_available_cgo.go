//go:build linux && cgo

package hotkey

// newX11ListenerIfAvailable attempts to create an X11 listener.
// Returns nil if DISPLAY is not set or the X11 connection fails.
func newX11ListenerIfAvailable() Listener {
	l := newX11Listener()
	if l == nil {
		return nil
	}
	return l
}
