//go:build darwin

package hotkey

// NewPlatformListener returns the best available Listener for macOS.
// The full implementation will use CGEventTap or NSEvent global monitors
// and requires Accessibility permission. This returns a stub for now.
func NewPlatformListener() Listener {
	return &stubListener{
		info: PlatformInfo{
			OS:     "darwin",
			Method: "stub",
		},
	}
}
