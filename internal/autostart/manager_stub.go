//go:build !windows

package autostart

// newPlatformManager returns the file-based WindowsManager on
// non-Windows platforms. This allows cross-platform testing of the
// autostart interface while the real RegistryManager is only
// available on Windows.
func newPlatformManager() Manager {
	return &WindowsManager{}
}
