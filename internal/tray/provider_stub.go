//go:build !(linux || windows) || notray

package tray

// NewPlatformProvider returns a HeadlessProvider on platforms where
// the native system tray is not yet implemented (Darwin) and when
// the notray build tag is set (for CGo-free CI builds).
func NewPlatformProvider() Provider {
	return &HeadlessProvider{}
}
