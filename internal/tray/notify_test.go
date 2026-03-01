package tray

import (
	"testing"
)

func TestHeadlessProvider_NotifyError(t *testing.T) {
	p := &HeadlessProvider{}

	p.NotifyError("Launch Failed", "Failed to launch myapp: command not found")

	notifs := p.Notifications()
	if len(notifs) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(notifs))
	}
	if notifs[0].Title != "Launch Failed" {
		t.Errorf("title = %q, want %q", notifs[0].Title, "Launch Failed")
	}
	if notifs[0].Message != "Failed to launch myapp: command not found" {
		t.Errorf("message = %q, want %q", notifs[0].Message, "Failed to launch myapp: command not found")
	}
}

func TestHeadlessProvider_NotifyError_Multiple(t *testing.T) {
	p := &HeadlessProvider{}

	p.NotifyError("Error 1", "first")
	p.NotifyError("Error 2", "second")
	p.NotifyError("Error 3", "third")

	notifs := p.Notifications()
	if len(notifs) != 3 {
		t.Fatalf("expected 3 notifications, got %d", len(notifs))
	}
}
