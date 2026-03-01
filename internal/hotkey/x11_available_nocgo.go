//go:build linux && !cgo

package hotkey

// newX11ListenerIfAvailable returns nil when cgo is not available,
// since the X11 listener requires cgo for Xlib bindings.
func newX11ListenerIfAvailable() Listener {
	return nil
}
