package autostart

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLinuxManagerInstall(t *testing.T) {
	tmpDir := t.TempDir()
	m := &LinuxManager{ConfigDir: tmpDir}

	binaryPath := "/usr/local/bin/orchestratr"
	if err := m.Install(binaryPath); err != nil {
		t.Fatalf("Install() error: %v", err)
	}

	// Service file should exist.
	serviceFile := filepath.Join(tmpDir, "systemd", "user", "orchestratr.service")
	data, err := os.ReadFile(serviceFile)
	if err != nil {
		t.Fatalf("reading service file: %v", err)
	}
	content := string(data)

	// Verify key content.
	if !strings.Contains(content, "ExecStart=/usr/local/bin/orchestratr start --foreground") {
		t.Errorf("service file missing ExecStart line:\n%s", content)
	}
	if !strings.Contains(content, "[Unit]") {
		t.Errorf("service file missing [Unit] section:\n%s", content)
	}
	if !strings.Contains(content, "[Install]") {
		t.Errorf("service file missing [Install] section:\n%s", content)
	}
	if !strings.Contains(content, "WantedBy=default.target") {
		t.Errorf("service file missing WantedBy:\n%s", content)
	}
	if !strings.Contains(content, "Restart=on-failure") {
		t.Errorf("service file missing Restart=on-failure:\n%s", content)
	}
}

func TestLinuxManagerInstallIdempotent(t *testing.T) {
	tmpDir := t.TempDir()
	m := &LinuxManager{ConfigDir: tmpDir}

	binaryPath := "/usr/local/bin/orchestratr"

	// Install twice — should not error.
	if err := m.Install(binaryPath); err != nil {
		t.Fatalf("first Install() error: %v", err)
	}
	if err := m.Install(binaryPath); err != nil {
		t.Fatalf("second Install() error: %v", err)
	}

	// Should still have exactly one service file with correct content.
	serviceFile := filepath.Join(tmpDir, "systemd", "user", "orchestratr.service")
	data, err := os.ReadFile(serviceFile)
	if err != nil {
		t.Fatalf("reading service file: %v", err)
	}
	if !strings.Contains(string(data), "ExecStart=/usr/local/bin/orchestratr start --foreground") {
		t.Error("service file content incorrect after idempotent install")
	}
}

func TestLinuxManagerInstallUpdatesPath(t *testing.T) {
	tmpDir := t.TempDir()
	m := &LinuxManager{ConfigDir: tmpDir}

	// Install with one path.
	if err := m.Install("/old/path/orchestratr"); err != nil {
		t.Fatalf("first Install() error: %v", err)
	}

	// Install again with a new path — should update.
	if err := m.Install("/new/path/orchestratr"); err != nil {
		t.Fatalf("second Install() error: %v", err)
	}

	serviceFile := filepath.Join(tmpDir, "systemd", "user", "orchestratr.service")
	data, err := os.ReadFile(serviceFile)
	if err != nil {
		t.Fatalf("reading service file: %v", err)
	}
	content := string(data)
	if strings.Contains(content, "/old/path/") {
		t.Error("service file still references old path after update")
	}
	if !strings.Contains(content, "ExecStart=/new/path/orchestratr start --foreground") {
		t.Errorf("service file missing updated ExecStart:\n%s", content)
	}
}

func TestLinuxManagerIsInstalled(t *testing.T) {
	tmpDir := t.TempDir()
	m := &LinuxManager{ConfigDir: tmpDir}

	if m.IsInstalled() {
		t.Error("IsInstalled() = true before install")
	}

	if err := m.Install("/usr/local/bin/orchestratr"); err != nil {
		t.Fatalf("Install() error: %v", err)
	}

	if !m.IsInstalled() {
		t.Error("IsInstalled() = false after install")
	}
}

func TestLinuxManagerUninstall(t *testing.T) {
	tmpDir := t.TempDir()
	m := &LinuxManager{ConfigDir: tmpDir}

	// Uninstall without install should return ErrNotInstalled.
	if err := m.Uninstall(); err != ErrNotInstalled {
		t.Errorf("Uninstall() before install = %v, want ErrNotInstalled", err)
	}

	// Install then uninstall.
	if err := m.Install("/usr/local/bin/orchestratr"); err != nil {
		t.Fatalf("Install() error: %v", err)
	}
	if err := m.Uninstall(); err != nil {
		t.Fatalf("Uninstall() error: %v", err)
	}

	// Service file should be gone.
	serviceFile := filepath.Join(tmpDir, "systemd", "user", "orchestratr.service")
	if _, err := os.Stat(serviceFile); !os.IsNotExist(err) {
		t.Error("service file still exists after uninstall")
	}

	if m.IsInstalled() {
		t.Error("IsInstalled() = true after uninstall")
	}
}

func TestLinuxManagerDescription(t *testing.T) {
	tmpDir := t.TempDir()
	m := &LinuxManager{ConfigDir: tmpDir}

	desc := m.Description()
	if desc == "" {
		t.Error("Description() returned empty string")
	}
	if !strings.Contains(desc, "systemd") {
		t.Errorf("Description() = %q, expected to mention systemd", desc)
	}
}
