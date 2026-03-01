// Package tray defines the system tray Provider interface and provides
// a headless stub for CI and testing environments.
package tray

// Provider is the interface for platform-specific system tray implementations.
// The daemon calls Setup once at startup, SetState on state changes, and
// Quit on shutdown.
type Provider interface {
	// Setup initializes the tray icon and context menu.
	Setup() error

	// SetState updates the tray icon/tooltip to reflect the daemon state
	// (e.g., "running", "paused", "listening").
	SetState(state string) error

	// Quit removes the tray icon and releases resources.
	Quit()

	// NotifyError displays an error notification via the system tray.
	NotifyError(title, message string)

	// OnPause registers a callback invoked when the user selects Pause.
	OnPause(fn func())

	// OnResume registers a callback invoked when the user selects Resume.
	OnResume(fn func())

	// OnQuit registers a callback invoked when the user selects Quit.
	OnQuit(fn func())

	// OnConfigure registers a callback invoked when the user selects Configure.
	OnConfigure(fn func())
}

// Notification records an error notification for testing.
type Notification struct {
	Title   string
	Message string
}

// HeadlessProvider is a no-op tray implementation for headless, CI, and
// testing environments where no display is available.
type HeadlessProvider struct {
	state  string
	notifs []Notification
}

// Setup is a no-op for headless mode.
func (h *HeadlessProvider) Setup() error { return nil }

// SetState records the state but performs no visual update.
func (h *HeadlessProvider) SetState(state string) error {
	h.state = state
	return nil
}

// Quit is a no-op for headless mode.
func (h *HeadlessProvider) Quit() {}

// OnPause is a no-op for headless mode.
func (h *HeadlessProvider) OnPause(fn func()) {}

// OnResume is a no-op for headless mode.
func (h *HeadlessProvider) OnResume(fn func()) {}

// OnQuit is a no-op for headless mode.
func (h *HeadlessProvider) OnQuit(fn func()) {}

// OnConfigure is a no-op for headless mode.
func (h *HeadlessProvider) OnConfigure(fn func()) {}

// NotifyError records the notification for later inspection (testing).
func (h *HeadlessProvider) NotifyError(title, message string) {
	h.notifs = append(h.notifs, Notification{Title: title, Message: message})
}

// LastState returns the last state set on the headless provider (for testing).
func (h *HeadlessProvider) LastState() string { return h.state }

// Notifications returns all recorded error notifications (for testing).
func (h *HeadlessProvider) Notifications() []Notification { return h.notifs }
