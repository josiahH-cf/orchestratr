package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun_Install(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	cfgPath := filepath.Join(configDir, "config.yml")

	// Write a minimal valid config.
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `leader_key: "ctrl+space"
apps:
  - name: test
    chord: t
    command: "echo test"
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("ORCHESTRATR_CONFIG", cfgPath)
	// Override autostart dir so we don't touch real systemd/launchd.
	t.Setenv("ORCHESTRATR_AUTOSTART_DIR", filepath.Join(tmpDir, "autostart"))

	var stdout, stderr bytes.Buffer
	err := run([]string{"install"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run(install) error = %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "autostart") {
		t.Errorf("install output should mention autostart, got: %q", out)
	}
}

func TestRun_InstallIdempotent(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	cfgPath := filepath.Join(configDir, "config.yml")

	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `leader_key: "ctrl+space"
apps: []
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("ORCHESTRATR_CONFIG", cfgPath)
	t.Setenv("ORCHESTRATR_AUTOSTART_DIR", filepath.Join(tmpDir, "autostart"))

	// Run install twice — both should succeed.
	var stdout1, stderr1 bytes.Buffer
	if err := run([]string{"install"}, &stdout1, &stderr1); err != nil {
		t.Fatalf("first install error = %v", err)
	}

	var stdout2, stderr2 bytes.Buffer
	if err := run([]string{"install"}, &stdout2, &stderr2); err != nil {
		t.Fatalf("second install error = %v", err)
	}
}

func TestRun_InstallShowsWSL2Warning(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yml")

	content := `leader_key: "ctrl+space"
apps: []
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create a fake /proc/version with WSL2 content.
	procDir := filepath.Join(tmpDir, "proc")
	if err := os.MkdirAll(procDir, 0o755); err != nil {
		t.Fatal(err)
	}
	procVersion := filepath.Join(procDir, "version")
	if err := os.WriteFile(procVersion, []byte("Linux 5.15-microsoft-standard-WSL2"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("ORCHESTRATR_CONFIG", cfgPath)
	t.Setenv("ORCHESTRATR_AUTOSTART_DIR", filepath.Join(tmpDir, "autostart"))
	t.Setenv("ORCHESTRATR_PROC_VERSION", procVersion)

	var stdout, stderr bytes.Buffer
	err := run([]string{"install"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run(install) error = %v", err)
	}

	combined := stdout.String() + stderr.String()
	if !strings.Contains(combined, "WSL2") {
		t.Errorf("install in WSL2 should warn about WSL2, got: %q", combined)
	}
}

func TestRun_Uninstall(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yml")

	content := `leader_key: "ctrl+space"
apps: []
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("ORCHESTRATR_CONFIG", cfgPath)
	autostartDir := filepath.Join(tmpDir, "autostart")
	t.Setenv("ORCHESTRATR_AUTOSTART_DIR", autostartDir)

	// Install first.
	var stdout1, stderr1 bytes.Buffer
	if err := run([]string{"install"}, &stdout1, &stderr1); err != nil {
		t.Fatalf("install error = %v", err)
	}

	// Uninstall.
	var stdout2, stderr2 bytes.Buffer
	err := run([]string{"uninstall"}, &stdout2, &stderr2)
	if err != nil {
		t.Fatalf("run(uninstall) error = %v", err)
	}

	out := stdout2.String()
	if !strings.Contains(out, "autostart") {
		t.Errorf("uninstall output should mention autostart, got: %q", out)
	}
}

func TestRun_UninstallWhenNotInstalled(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("ORCHESTRATR_AUTOSTART_DIR", filepath.Join(tmpDir, "autostart"))

	var stdout, stderr bytes.Buffer
	err := run([]string{"uninstall"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run(uninstall) error = %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "not installed") {
		t.Errorf("uninstall when not installed should say 'not installed', got: %q", out)
	}
}

func TestRun_InstallCreatesDefaultConfig(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "orchestratr", "config.yml")

	// Config does not exist yet.
	t.Setenv("ORCHESTRATR_CONFIG", cfgPath)
	t.Setenv("ORCHESTRATR_AUTOSTART_DIR", filepath.Join(tmpDir, "autostart"))

	var stdout, stderr bytes.Buffer
	err := run([]string{"install"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run(install) error = %v", err)
	}

	// Default config should have been created.
	if _, statErr := os.Stat(cfgPath); os.IsNotExist(statErr) {
		t.Error("install should create default config file")
	}
}

func TestRun_UsageIncludesInstallUninstall(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run(nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "install") {
		t.Error("usage should mention 'install'")
	}
	if !strings.Contains(out, "uninstall") {
		t.Error("usage should mention 'uninstall'")
	}
}
