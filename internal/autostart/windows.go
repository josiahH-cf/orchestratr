package autostart

import (
	"fmt"
	"os"
	"path/filepath"
)

const regFileName = "orchestratr.reg"

// WindowsManager manages autostart via a registry command file.
// On real Windows, this would use the registry API directly
// (HKCU\Software\Microsoft\Windows\CurrentVersion\Run). For
// portability and testability, this implementation writes a
// registry-format file that records the intended autostart command.
type WindowsManager struct {
	// RegistryDir overrides the directory for the registry file.
	// If empty, defaults to the user's AppData\Roaming\orchestratr.
	RegistryDir string
}

// regFilePath returns the path to the registry simulation file.
func (m *WindowsManager) regFilePath() string {
	base := m.RegistryDir
	if base == "" {
		home, _ := os.UserHomeDir()
		base = filepath.Join(home, "AppData", "Roaming", "orchestratr")
	}
	return filepath.Join(base, regFileName)
}

// Install creates or updates the autostart registry entry.
func (m *WindowsManager) Install(binaryPath string) error {
	dir := filepath.Dir(m.regFilePath())
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating registry dir: %w", err)
	}

	// The registry value: "C:\path\to\orchestratr.exe" start --foreground
	value := fmt.Sprintf(`"%s" start --foreground`, binaryPath)
	content := fmt.Sprintf(
		"Windows Registry Editor Version 5.00\n\n"+
			"[HKEY_CURRENT_USER\\Software\\Microsoft\\Windows\\CurrentVersion\\Run]\n"+
			"\"orchestratr\"=\"%s\"\n",
		value,
	)

	if err := os.WriteFile(m.regFilePath(), []byte(content), 0o644); err != nil {
		return fmt.Errorf("writing registry file: %w", err)
	}
	return nil
}

// Uninstall removes the autostart registry entry.
func (m *WindowsManager) Uninstall() error {
	if !m.IsInstalled() {
		return ErrNotInstalled
	}
	if err := os.Remove(m.regFilePath()); err != nil {
		return fmt.Errorf("removing registry file: %w", err)
	}
	return nil
}

// IsInstalled reports whether the autostart registry entry exists.
func (m *WindowsManager) IsInstalled() bool {
	_, err := os.Stat(m.regFilePath())
	return err == nil
}

// Description returns a human-readable description of the autostart method.
func (m *WindowsManager) Description() string {
	return "Windows registry key HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Run"
}
