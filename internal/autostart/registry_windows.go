//go:build windows

package autostart

import (
	"fmt"

	"golang.org/x/sys/windows/registry"
)

const (
	autostartKeyPath = `Software\Microsoft\Windows\CurrentVersion\Run`
	autostartValue   = "orchestratr"
)

// RegistryManager manages autostart via the Windows registry at
// HKCU\Software\Microsoft\Windows\CurrentVersion\Run. This is the
// production implementation that writes directly to the registry.
type RegistryManager struct{}

// Install creates or updates the autostart registry entry so
// orchestratr starts at user login.
func (m *RegistryManager) Install(binaryPath string) error {
	key, _, err := registry.CreateKey(
		registry.CURRENT_USER,
		autostartKeyPath,
		registry.SET_VALUE,
	)
	if err != nil {
		return fmt.Errorf("opening registry key: %w", err)
	}
	defer key.Close()

	value := fmt.Sprintf(`"%s" start --foreground`, binaryPath)
	if err := key.SetStringValue(autostartValue, value); err != nil {
		return fmt.Errorf("setting registry value: %w", err)
	}
	return nil
}

// Uninstall removes the autostart registry entry. Returns
// ErrNotInstalled if no entry exists.
func (m *RegistryManager) Uninstall() error {
	if !m.IsInstalled() {
		return ErrNotInstalled
	}

	key, err := registry.OpenKey(
		registry.CURRENT_USER,
		autostartKeyPath,
		registry.SET_VALUE,
	)
	if err != nil {
		return fmt.Errorf("opening registry key: %w", err)
	}
	defer key.Close()

	if err := key.DeleteValue(autostartValue); err != nil {
		return fmt.Errorf("deleting registry value: %w", err)
	}
	return nil
}

// IsInstalled reports whether the autostart registry entry exists.
func (m *RegistryManager) IsInstalled() bool {
	key, err := registry.OpenKey(
		registry.CURRENT_USER,
		autostartKeyPath,
		registry.QUERY_VALUE,
	)
	if err != nil {
		return false
	}
	defer key.Close()

	_, _, err = key.GetStringValue(autostartValue)
	return err == nil
}

// Description returns a human-readable description of the autostart method.
func (m *RegistryManager) Description() string {
	return `Windows registry key HKCU\Software\Microsoft\Windows\CurrentVersion\Run`
}

// newPlatformManager returns the Windows-specific RegistryManager.
// This is called from NewManager() on Windows builds.
func newPlatformManager() Manager {
	return &RegistryManager{}
}
