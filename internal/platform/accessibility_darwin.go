//go:build darwin

package platform

import (
	"os/exec"
	"strings"
)

// CheckAccessibility reports whether the application has macOS
// Accessibility permission. It uses the osascript approach to check
// whether the current process can use accessibility features.
// Returns true if permission is granted, false otherwise.
func CheckAccessibility() (bool, error) {
	// Use osascript to check if accessibility is enabled.
	// This returns "true" if the app has permission.
	cmd := exec.Command("osascript", "-e",
		`tell application "System Events" to return (exists process 1)`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		// If osascript fails with an error about assistive access,
		// the permission is not granted.
		if strings.Contains(string(out), "assistive") ||
			strings.Contains(string(out), "accessibility") {
			return false, nil
		}
		// Other errors might be transient — report them.
		return false, err
	}
	return strings.TrimSpace(string(out)) == "true", nil
}

// AccessibilityPrompt returns instructions for granting macOS
// Accessibility permission.
func AccessibilityPrompt() string {
	return `orchestratr needs Accessibility permission to capture global hotkeys.

To grant permission:
  1. Open System Settings → Privacy & Security → Accessibility
  2. Click the lock icon to make changes
  3. Add orchestratr to the allowed list
  4. Restart orchestratr

Without Accessibility permission, orchestratr will run in degraded mode
(system tray only, no hotkey capture).`
}
