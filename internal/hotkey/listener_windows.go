//go:build windows

package hotkey

// NewPlatformListener returns the best available Listener for Windows.
// The full implementation will use RegisterHotKey for the leader key
// and a low-level keyboard hook for chord capture. This returns a stub
// for now.
func NewPlatformListener() Listener {
	return &stubListener{
		info: PlatformInfo{
			OS:     "windows",
			Method: "stub",
		},
	}
}
