//go:build linux

package hotkey

// NewPlatformListener returns the best available Listener for Linux.
// Detection order:
//  1. Wayland (WAYLAND_DISPLAY set) + freedesktop GlobalShortcuts portal → wayland_portal
//  2. X11 (DISPLAY set) → x11_xgrabkey
//  3. Fallback → stub with warning
//
// Full implementations are future work; this returns a stub for now.
func NewPlatformListener() Listener {
	// TODO: detect Wayland/X11 and return real implementations.
	return &stubListener{
		info: PlatformInfo{
			OS:     "linux",
			Method: "stub",
		},
	}
}
