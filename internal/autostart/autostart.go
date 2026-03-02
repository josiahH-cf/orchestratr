// Package autostart provides cross-platform autostart configuration
// for the orchestratr daemon. Each platform implements the Manager
// interface to install, uninstall, and query autostart entries.
package autostart

import "errors"

// ErrNotInstalled is returned when Uninstall is called but no
// autostart entry exists.
var ErrNotInstalled = errors.New("autostart entry not installed")

// ErrNotImplemented is returned on platforms where autostart
// configuration has not been implemented.
var ErrNotImplemented = errors.New("autostart not implemented for this platform")

// Manager configures platform-specific autostart for orchestratr.
type Manager interface {
	// Install creates or updates the autostart entry so orchestratr
	// starts at user login. It is idempotent — calling Install when
	// already installed updates the entry without duplicating it.
	Install(binaryPath string) error

	// Uninstall removes the autostart entry. Returns ErrNotInstalled
	// if no entry exists.
	Uninstall() error

	// IsInstalled reports whether an autostart entry currently exists.
	IsInstalled() bool

	// Description returns a human-readable string describing what
	// the autostart entry does (e.g. "systemd user service at ...").
	Description() string
}
