//go:build (linux || windows) && !notray

package tray

import (
	"os"
	"runtime"
)

// NewPlatformProvider returns the platform-appropriate tray Provider.
// On Linux it returns a SystrayProvider when a D-Bus session is
// detected; otherwise it returns a HeadlessProvider so the daemon
// can start without a desktop environment. On Windows it always
// returns a SystrayProvider (the Win32 notification area is always
// available).
func NewPlatformProvider() Provider {
	if !isDisplayAvailable() {
		return &HeadlessProvider{}
	}
	return &SystrayProvider{}
}

// isDisplayAvailable returns true when a tray backend is expected to
// be accessible at runtime.
//
//   - Linux: at least one of DBUS_SESSION_BUS_ADDRESS, DISPLAY, or
//     WAYLAND_DISPLAY must be set. Without these the process is
//     effectively headless (CI, SSH session, container).
//   - Windows: the notification area is always available.
func isDisplayAvailable() bool {
	switch runtime.GOOS {
	case "windows":
		return true
	case "linux":
		return os.Getenv("DBUS_SESSION_BUS_ADDRESS") != "" ||
			os.Getenv("DISPLAY") != "" ||
			os.Getenv("WAYLAND_DISPLAY") != ""
	default:
		return false
	}
}
