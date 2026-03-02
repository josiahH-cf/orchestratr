package autostart

import (
	"fmt"
	"os"
	"path/filepath"
)

const regFileName = "orchestratr.reg"

// WindowsManager manages autostart via a registry command file.
// NOTE: This is a file-based stub that generates a .reg file for
// import. A production Windows deployment should use
// golang.org/x/sys/windows/registry to write directly to
// HKCU\Software\Microsoft\Windows\CurrentVersion\Run.
// The file-based approach is used for cross-platform testability.
type WindowsManager struct {
	// RegistryDir overrides the directory for the registry file.
	// If empty, defaults to the user's AppData\Roaming\orchestratr.
	RegistryDir string
}

// regFilePath returns the path to the registry simulation file.
func (m *WindowsManager) regFilePath() (string, error) {
	base := m.RegistryDir
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("detecting home directory: %w", err)
		}
		base = filepath.Join(home, "AppData", "Roaming", "orchestratr")
	}
	return filepath.Join(base, regFileName), nil
}

// Install creates or updates the autostart registry entry.
func (m *WindowsManager) Install(binaryPath string) error {
	path, err := m.regFilePath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
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

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("writing registry file: %w", err)
	}
	return nil
}

// Uninstall removes the autostart registry entry.
func (m *WindowsManager) Uninstall() error {
	if !m.IsInstalled() {
		return ErrNotInstalled
	}
	path, err := m.regFilePath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("removing registry file: %w", err)
	}
	return nil
}

// IsInstalled reports whether the autostart registry entry exists.
func (m *WindowsManager) IsInstalled() bool {
	path, err := m.regFilePath()
	if err != nil {
		return false
	}
	_, statErr := os.Stat(path)
	return statErr == nil
}

// Description returns a human-readable description of the autostart method.
func (m *WindowsManager) Description() string {
	return "Windows registry key HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Run"
}
