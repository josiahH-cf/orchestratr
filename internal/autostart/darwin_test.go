package autostart

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDarwinManagerInstall(t *testing.T) {
	tmpDir := t.TempDir()
	m := &DarwinManager{LaunchAgentsDir: tmpDir}

	binaryPath := "/usr/local/bin/orchestratr"
	if err := m.Install(binaryPath); err != nil {
		t.Fatalf("Install() error: %v", err)
	}

	// Plist file should exist.
	plistFile := filepath.Join(tmpDir, "com.orchestratr.daemon.plist")
	data, err := os.ReadFile(plistFile)
	if err != nil {
		t.Fatalf("reading plist file: %v", err)
	}
	content := string(data)

	// Verify key content.
	if !strings.Contains(content, "<string>com.orchestratr.daemon</string>") {
		t.Errorf("plist missing Label:\n%s", content)
	}
	if !strings.Contains(content, "<string>/usr/local/bin/orchestratr</string>") {
		t.Errorf("plist missing binary path:\n%s", content)
	}
	if !strings.Contains(content, "<string>start</string>") {
		t.Errorf("plist missing 'start' argument:\n%s", content)
	}
	if !strings.Contains(content, "<string>--foreground</string>") {
		t.Errorf("plist missing '--foreground' argument:\n%s", content)
	}
	if !strings.Contains(content, "<key>RunAtLoad</key>") {
		t.Errorf("plist missing RunAtLoad:\n%s", content)
	}
	if !strings.Contains(content, "<key>KeepAlive</key>") {
		t.Errorf("plist missing KeepAlive:\n%s", content)
	}
}

func TestDarwinManagerInstallIdempotent(t *testing.T) {
	tmpDir := t.TempDir()
	m := &DarwinManager{LaunchAgentsDir: tmpDir}

	binaryPath := "/usr/local/bin/orchestratr"
	if err := m.Install(binaryPath); err != nil {
		t.Fatalf("first Install() error: %v", err)
	}
	if err := m.Install(binaryPath); err != nil {
		t.Fatalf("second Install() error: %v", err)
	}

	plistFile := filepath.Join(tmpDir, "com.orchestratr.daemon.plist")
	data, err := os.ReadFile(plistFile)
	if err != nil {
		t.Fatalf("reading plist file: %v", err)
	}
	if !strings.Contains(string(data), "/usr/local/bin/orchestratr") {
		t.Error("plist content incorrect after idempotent install")
	}
}

func TestDarwinManagerInstallUpdatesPath(t *testing.T) {
	tmpDir := t.TempDir()
	m := &DarwinManager{LaunchAgentsDir: tmpDir}

	if err := m.Install("/old/path/orchestratr"); err != nil {
		t.Fatalf("first Install() error: %v", err)
	}
	if err := m.Install("/new/path/orchestratr"); err != nil {
		t.Fatalf("second Install() error: %v", err)
	}

	plistFile := filepath.Join(tmpDir, "com.orchestratr.daemon.plist")
	data, err := os.ReadFile(plistFile)
	if err != nil {
		t.Fatalf("reading plist file: %v", err)
	}
	content := string(data)
	if strings.Contains(content, "/old/path/") {
		t.Error("plist still references old path after update")
	}
	if !strings.Contains(content, "/new/path/orchestratr") {
		t.Error("plist missing updated binary path")
	}
}

func TestDarwinManagerIsInstalled(t *testing.T) {
	tmpDir := t.TempDir()
	m := &DarwinManager{LaunchAgentsDir: tmpDir}

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

func TestDarwinManagerUninstall(t *testing.T) {
	tmpDir := t.TempDir()
	m := &DarwinManager{LaunchAgentsDir: tmpDir}

	if err := m.Uninstall(); err != ErrNotInstalled {
		t.Errorf("Uninstall() before install = %v, want ErrNotInstalled", err)
	}

	if err := m.Install("/usr/local/bin/orchestratr"); err != nil {
		t.Fatalf("Install() error: %v", err)
	}
	if err := m.Uninstall(); err != nil {
		t.Fatalf("Uninstall() error: %v", err)
	}

	plistFile := filepath.Join(tmpDir, "com.orchestratr.daemon.plist")
	if _, err := os.Stat(plistFile); !os.IsNotExist(err) {
		t.Error("plist file still exists after uninstall")
	}

	if m.IsInstalled() {
		t.Error("IsInstalled() = true after uninstall")
	}
}

func TestDarwinManagerDescription(t *testing.T) {
	tmpDir := t.TempDir()
	m := &DarwinManager{LaunchAgentsDir: tmpDir}

	desc := m.Description()
	if desc == "" {
		t.Error("Description() returned empty string")
	}
	if !strings.Contains(strings.ToLower(desc), "launch agent") {
		t.Errorf("Description() = %q, expected to mention Launch Agent", desc)
	}
}
