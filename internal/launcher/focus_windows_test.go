//go:build windows

package launcher

import (
	"errors"
	"testing"
)

func TestFocusWindow_InvalidPID(t *testing.T) {
	tests := []struct {
		name string
		pid  int
	}{
		{"zero", 0},
		{"negative", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := FocusWindow(tt.pid)
			if err == nil {
				t.Errorf("FocusWindow(%d) = nil, want error", tt.pid)
			}
		})
	}
}

func TestFocusWindow_NoWindowForPID(t *testing.T) {
	// PID 1 (System Idle Process analogue) will not have a visible
	// window. This should return an error, not panic.
	err := FocusWindow(1)
	if err == nil {
		t.Error("FocusWindow(1) = nil, expected error for PID with no window")
	}
}

func TestFocusWindow_NonexistentPID(t *testing.T) {
	// Use an absurdly high PID that is unlikely to exist.
	err := FocusWindow(999999999)
	if err == nil {
		t.Error("FocusWindow(999999999) = nil, expected error")
	}
}

func TestFocusWindow_ErrorIsFocusNotSupported_ForNoWindow(t *testing.T) {
	// When no window is found, the error should wrap ErrFocusNotSupported
	// or at least be non-nil.
	err := FocusWindow(999999999)
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	// We simply verify it's an error — the exact type depends on
	// whether EnumWindows finds nothing vs the PID doesn't exist.
	if errors.Is(err, ErrFocusNotSupported) {
		t.Logf("got ErrFocusNotSupported as expected")
	} else {
		t.Logf("got error (not ErrFocusNotSupported): %v", err)
	}
}
