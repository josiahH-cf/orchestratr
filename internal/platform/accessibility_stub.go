//go:build !darwin

package platform

// CheckAccessibility reports whether the application has accessibility
// permission. On non-macOS platforms, this always returns true since
// accessibility permissions are a macOS-specific concept.
func CheckAccessibility() (bool, error) {
	return true, nil
}

// AccessibilityPrompt returns instructions for granting accessibility
// permission. On non-macOS platforms this returns a generic message.
func AccessibilityPrompt() string {
	return "Accessibility permissions are not required on this platform."
}
