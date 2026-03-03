//go:build windows

package autostart

import (
	"strings"
	"testing"

	"golang.org/x/sys/windows/registry"
)

func TestRegistryManagerInstall(t *testing.T) {
	m := &RegistryManager{}
	defer cleanup(t)

	binaryPath := `C:\Program Files\orchestratr\orchestratr.exe`
	if err := m.Install(binaryPath); err != nil {
		t.Fatalf("Install() error: %v", err)
	}

	if !m.IsInstalled() {
		t.Error("IsInstalled() = false after Install")
	}

	// Verify the registry value directly.
	val := readRegistryValue(t)
	if !strings.Contains(val, binaryPath) {
		t.Errorf("registry value %q does not contain binary path", val)
	}
	if !strings.Contains(val, "start --foreground") {
		t.Errorf("registry value %q does not contain 'start --foreground'", val)
	}
}

func TestRegistryManagerInstallIdempotent(t *testing.T) {
	m := &RegistryManager{}
	defer cleanup(t)

	if err := m.Install(`C:\orchestratr.exe`); err != nil {
		t.Fatalf("first Install() error: %v", err)
	}
	if err := m.Install(`C:\orchestratr.exe`); err != nil {
		t.Fatalf("second Install() error: %v", err)
	}
}

func TestRegistryManagerInstallUpdatesPath(t *testing.T) {
	m := &RegistryManager{}
	defer cleanup(t)

	if err := m.Install(`C:\old\orchestratr.exe`); err != nil {
		t.Fatalf("first Install() error: %v", err)
	}
	if err := m.Install(`C:\new\orchestratr.exe`); err != nil {
		t.Fatalf("second Install() error: %v", err)
	}

	val := readRegistryValue(t)
	if strings.Contains(val, `C:\old\`) {
		t.Error("registry still references old path")
	}
	if !strings.Contains(val, `C:\new\orchestratr.exe`) {
		t.Error("registry missing updated path")
	}
}

func TestRegistryManagerUninstall(t *testing.T) {
	m := &RegistryManager{}
	defer cleanup(t)

	// Uninstall before install should return ErrNotInstalled.
	if err := m.Uninstall(); err != ErrNotInstalled {
		t.Errorf("Uninstall() before install = %v, want ErrNotInstalled", err)
	}

	if err := m.Install(`C:\orchestratr.exe`); err != nil {
		t.Fatalf("Install() error: %v", err)
	}
	if err := m.Uninstall(); err != nil {
		t.Fatalf("Uninstall() error: %v", err)
	}

	if m.IsInstalled() {
		t.Error("IsInstalled() = true after Uninstall")
	}
}

func TestRegistryManagerDescription(t *testing.T) {
	m := &RegistryManager{}
	desc := m.Description()
	if desc == "" {
		t.Error("Description() returned empty string")
	}
	if !strings.Contains(strings.ToLower(desc), "registry") {
		t.Errorf("Description() = %q, expected to mention registry", desc)
	}
}

func TestNewPlatformManager_IsRegistry(t *testing.T) {
	m := newPlatformManager()
	if _, ok := m.(*RegistryManager); !ok {
		t.Errorf("newPlatformManager() returned %T, want *RegistryManager", m)
	}
}

// --- helpers ---

// cleanup removes the test registry entry if it exists.
func cleanup(t *testing.T) {
	t.Helper()
	key, err := registry.OpenKey(
		registry.CURRENT_USER,
		autostartKeyPath,
		registry.SET_VALUE,
	)
	if err != nil {
		return
	}
	defer key.Close()
	_ = key.DeleteValue(autostartValue)
}

// readRegistryValue reads the orchestratr autostart value directly.
func readRegistryValue(t *testing.T) string {
	t.Helper()
	key, err := registry.OpenKey(
		registry.CURRENT_USER,
		autostartKeyPath,
		registry.QUERY_VALUE,
	)
	if err != nil {
		t.Fatalf("opening registry key: %v", err)
	}
	defer key.Close()

	val, _, err := key.GetStringValue(autostartValue)
	if err != nil {
		t.Fatalf("reading registry value: %v", err)
	}
	return val
}
