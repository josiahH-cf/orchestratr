package autostart

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

const serviceName = "orchestratr.service"

const systemdServiceTmpl = `[Unit]
Description=orchestratr — app launcher daemon
After=graphical-session.target

[Service]
ExecStart={{.BinaryPath}} start --foreground
Restart=on-failure
RestartSec=5

[Install]
WantedBy=default.target
`

// LinuxManager manages autostart via a systemd user service.
type LinuxManager struct {
	// ConfigDir overrides the base config directory. If empty,
	// defaults to ~/.config.
	ConfigDir string
}

// servicePath returns the full path to the systemd user service file.
func (m *LinuxManager) servicePath() string {
	base := m.ConfigDir
	if base == "" {
		home, _ := os.UserHomeDir()
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "systemd", "user", serviceName)
}

// Install creates or updates the systemd user service file.
func (m *LinuxManager) Install(binaryPath string) error {
	path := m.servicePath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating systemd user dir: %w", err)
	}

	tmpl, err := template.New("service").Parse(systemdServiceTmpl)
	if err != nil {
		return fmt.Errorf("parsing service template: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating service file: %w", err)
	}
	defer f.Close()

	data := struct{ BinaryPath string }{BinaryPath: binaryPath}
	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("writing service file: %w", err)
	}

	return nil
}

// Uninstall removes the systemd user service file.
func (m *LinuxManager) Uninstall() error {
	if !m.IsInstalled() {
		return ErrNotInstalled
	}
	if err := os.Remove(m.servicePath()); err != nil {
		return fmt.Errorf("removing service file: %w", err)
	}
	return nil
}

// IsInstalled reports whether the systemd user service file exists.
func (m *LinuxManager) IsInstalled() bool {
	_, err := os.Stat(m.servicePath())
	return err == nil
}

// Description returns a human-readable description of the autostart method.
func (m *LinuxManager) Description() string {
	return fmt.Sprintf("systemd user service at %s", m.servicePath())
}
