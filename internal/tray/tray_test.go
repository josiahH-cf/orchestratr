package tray

import (
	"testing"
)

func TestHeadlessProvider_Implements_Provider(t *testing.T) {
	var _ Provider = (*HeadlessProvider)(nil)
}

func TestHeadlessProvider_Setup(t *testing.T) {
	p := &HeadlessProvider{}
	if err := p.Setup(); err != nil {
		t.Fatalf("Setup() error = %v", err)
	}
}

func TestHeadlessProvider_SetState(t *testing.T) {
	p := &HeadlessProvider{}
	_ = p.Setup()

	tests := []string{"running", "paused", "listening"}
	for _, state := range tests {
		if err := p.SetState(state); err != nil {
			t.Errorf("SetState(%q) error = %v", state, err)
		}
	}

	if p.LastState() != "listening" {
		t.Errorf("LastState() = %q, want %q", p.LastState(), "listening")
	}
}

func TestHeadlessProvider_Quit(t *testing.T) {
	p := &HeadlessProvider{}
	_ = p.Setup()
	p.Quit()
	// Should not panic.
}
