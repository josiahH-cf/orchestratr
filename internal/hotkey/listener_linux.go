//go:build linux

package hotkey

import "os"

// NewPlatformListener returns the best available Listener for Linux.
// Detection order:
//  1. X11 (DISPLAY set, cgo available) → x11_xgrabkey
//  2. Wayland (WAYLAND_DISPLAY set) → stub with fallback note
//     (users configure compositor to run "orchestratr trigger")
//  3. Fallback → stub with warning
func NewPlatformListener() Listener {
	// Try X11 first if DISPLAY is set.
	if l := newX11ListenerIfAvailable(); l != nil {
		return l
	}

	// Wayland fallback: the trigger CLI command is the primary
	// mechanism (see decisions/0002-wayland-hotkey-strategy.md).
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		return &stubListener{
			info: PlatformInfo{
				OS:     "linux",
				Method: "wayland_stub",
			},
		}
	}

	return &stubListener{
		info: PlatformInfo{
			OS:     "linux",
			Method: "stub",
		},
	}
}
