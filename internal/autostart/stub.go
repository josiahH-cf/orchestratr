package autostart

import "runtime"

// NewManager returns the platform-appropriate autostart Manager.
func NewManager() Manager {
	switch runtime.GOOS {
	case "linux":
		return &LinuxManager{}
	case "darwin":
		return &DarwinManager{}
	case "windows":
		return &WindowsManager{}
	default:
		return &StubManager{}
	}
}

// StubManager is a no-op Manager for unsupported platforms.
type StubManager struct{}

// Install returns ErrNotImplemented.
func (s *StubManager) Install(_ string) error { return ErrNotImplemented }

// Uninstall returns ErrNotImplemented.
func (s *StubManager) Uninstall() error { return ErrNotImplemented }

// IsInstalled always returns false.
func (s *StubManager) IsInstalled() bool { return false }

// Description returns a description indicating no autostart support.
func (s *StubManager) Description() string {
	return "autostart not supported on " + runtime.GOOS
}
