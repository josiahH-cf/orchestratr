//go:build linux

package launcher

import (
	"fmt"
	"os/exec"
	"strings"
)

// FocusWindow attempts to bring the window belonging to the given PID
// to the foreground. On Linux/X11 this uses xdotool to search for the
// window and activate it. Returns an error if xdotool is not installed,
// no window is found, or the activation fails. This is best-effort on
// Wayland compositors.
func FocusWindow(pid int) error {
	if pid <= 0 {
		return fmt.Errorf("invalid PID %d", pid)
	}

	xdotool, err := exec.LookPath("xdotool")
	if err != nil {
		return fmt.Errorf("%w: xdotool not found: %v", ErrFocusNotSupported, err)
	}

	// Search for windows owned by this PID.
	out, err := exec.Command(xdotool, "search", "--pid", fmt.Sprintf("%d", pid)).Output()
	if err != nil {
		return fmt.Errorf("xdotool search --pid %d failed: %w", pid, err)
	}

	wids := strings.Fields(strings.TrimSpace(string(out)))
	if len(wids) == 0 {
		return fmt.Errorf("no windows found for PID %d", pid)
	}

	// Activate the first window found.
	if err := exec.Command(xdotool, "windowactivate", "--sync", wids[0]).Run(); err != nil {
		return fmt.Errorf("xdotool windowactivate %s failed: %w", wids[0], err)
	}

	return nil
}
