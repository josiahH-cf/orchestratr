package autostart

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWindowsManagerInstall(t *testing.T) {
	tmpDir := t.TempDir()
	m := &WindowsManager{RegistryDir: tmpDir}

	binaryPath := `C:\Program Files\orchestratr\orchestratr.exe`
	if err := m.Install(binaryPath); err != nil {
		t.Fatalf("Install() error: %v", err)
	}

	// Registry simulation file should exist.
	regFile := filepath.Join(tmpDir, "orchestratr.reg")
	data, err := os.ReadFile(regFile)
	if err != nil {
		t.Fatalf("reading reg file: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, binaryPath) {
		t.Errorf("reg file missing binary path:\n%s", content)
	}
	if !strings.Contains(content, "start") {
		t.Errorf("reg file missing 'start' argument:\n%s", content)
	}
}

func TestWindowsManagerInstallIdempotent(t *testing.T) {
	tmpDir := t.TempDir()
	m := &WindowsManager{RegistryDir: tmpDir}

	binaryPath := `C:\orchestratr.exe`
	if err := m.Install(binaryPath); err != nil {
		t.Fatalf("first Install() error: %v", err)
	}
	if err := m.Install(binaryPath); err != nil {
		t.Fatalf("second Install() error: %v", err)
	}
}

func TestWindowsManagerInstallUpdatesPath(t *testing.T) {
	tmpDir := t.TempDir()
	m := &WindowsManager{RegistryDir: tmpDir}

	if err := m.Install(`C:\old\orchestratr.exe`); err != nil {
		t.Fatalf("first Install() error: %v", err)
	}
	if err := m.Install(`C:\new\orchestratr.exe`); err != nil {
		t.Fatalf("second Install() error: %v", err)
	}

	regFile := filepath.Join(tmpDir, "orchestratr.reg")
	data, err := os.ReadFile(regFile)
	if err != nil {
		t.Fatalf("reading reg file: %v", err)
	}
	content := string(data)
	if strings.Contains(content, `C:\old\`) {
		t.Error("reg file still references old path")
	}
	if !strings.Contains(content, `C:\new\orchestratr.exe`) {
		t.Error("reg file missing updated path")
	}
}

func TestWindowsManagerIsInstalled(t *testing.T) {
	tmpDir := t.TempDir()
	m := &WindowsManager{RegistryDir: tmpDir}

	if m.IsInstalled() {
		t.Error("IsInstalled() = true before install")
	}

	if err := m.Install(`C:\orchestratr.exe`); err != nil {
		t.Fatalf("Install() error: %v", err)
	}

	if !m.IsInstalled() {
		t.Error("IsInstalled() = false after install")
	}
}

func TestWindowsManagerUninstall(t *testing.T) {
	tmpDir := t.TempDir()
	m := &WindowsManager{RegistryDir: tmpDir}

	if err := m.Uninstall(); err != ErrNotInstalled {
		t.Errorf("Uninstall() before install = %v, want ErrNotInstalled", err)
	}

	if err := m.Install(`C:\orchestratr.exe`); err != nil {
		t.Fatalf("Install() error: %v", err)
	}
	if err := m.Uninstall(); err != nil {
		t.Fatalf("Uninstall() error: %v", err)
	}

	regFile := filepath.Join(tmpDir, "orchestratr.reg")
	if _, err := os.Stat(regFile); !os.IsNotExist(err) {
		t.Error("reg file still exists after uninstall")
	}

	if m.IsInstalled() {
		t.Error("IsInstalled() = true after uninstall")
	}
}

func TestWindowsManagerDescription(t *testing.T) {
	tmpDir := t.TempDir()
	m := &WindowsManager{RegistryDir: tmpDir}

	desc := m.Description()
	if desc == "" {
		t.Error("Description() returned empty string")
	}
	if !strings.Contains(strings.ToLower(desc), "registry") && !strings.Contains(strings.ToLower(desc), "windows") {
		t.Errorf("Description() = %q, expected to mention registry or windows", desc)
	}
}
