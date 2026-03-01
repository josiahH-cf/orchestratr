package autostart

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

const plistName = "com.orchestratr.daemon.plist"

const launchAgentTmpl = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>com.orchestratr.daemon</string>
  <key>ProgramArguments</key>
  <array>
    <string>{{.BinaryPath}}</string>
    <string>start</string>
    <string>--foreground</string>
  </array>
  <key>RunAtLoad</key>
  <true/>
  <key>KeepAlive</key>
  <true/>
</dict>
</plist>
`

// DarwinManager manages autostart via a macOS Launch Agent plist.
type DarwinManager struct {
	// LaunchAgentsDir overrides the Launch Agents directory. If empty,
	// defaults to ~/Library/LaunchAgents.
	LaunchAgentsDir string
}

// plistPath returns the full path to the Launch Agent plist file.
func (m *DarwinManager) plistPath() string {
	base := m.LaunchAgentsDir
	if base == "" {
		home, _ := os.UserHomeDir()
		base = filepath.Join(home, "Library", "LaunchAgents")
	}
	return filepath.Join(base, plistName)
}

// Install creates or updates the Launch Agent plist file.
func (m *DarwinManager) Install(binaryPath string) error {
	dir := m.LaunchAgentsDir
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, "Library", "LaunchAgents")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating LaunchAgents dir: %w", err)
	}

	tmpl, err := template.New("plist").Parse(launchAgentTmpl)
	if err != nil {
		return fmt.Errorf("parsing plist template: %w", err)
	}

	f, err := os.Create(m.plistPath())
	if err != nil {
		return fmt.Errorf("creating plist file: %w", err)
	}
	defer f.Close()

	data := struct{ BinaryPath string }{BinaryPath: binaryPath}
	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("writing plist file: %w", err)
	}

	return nil
}

// Uninstall removes the Launch Agent plist file.
func (m *DarwinManager) Uninstall() error {
	if !m.IsInstalled() {
		return ErrNotInstalled
	}
	if err := os.Remove(m.plistPath()); err != nil {
		return fmt.Errorf("removing plist file: %w", err)
	}
	return nil
}

// IsInstalled reports whether the Launch Agent plist file exists.
func (m *DarwinManager) IsInstalled() bool {
	_, err := os.Stat(m.plistPath())
	return err == nil
}

// Description returns a human-readable description of the autostart method.
func (m *DarwinManager) Description() string {
	return fmt.Sprintf("macOS Launch Agent at %s", m.plistPath())
}
