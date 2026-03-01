//go:build linux && cgo

package hotkey

import (
	"os"
	"testing"
)

func TestNewX11Listener_NoDisplay(t *testing.T) {
	old := os.Getenv("DISPLAY")
	os.Unsetenv("DISPLAY")
	defer os.Setenv("DISPLAY", old)

	l := newX11Listener()
	if l != nil {
		t.Error("expected nil when DISPLAY is not set")
		l.Stop()
	}
}

func TestNewX11Listener_WithDisplay(t *testing.T) {
	if os.Getenv("DISPLAY") == "" {
		t.Skip("DISPLAY not set")
	}

	l := newX11Listener()
	if l == nil {
		t.Skip("X11 connection failed (no real X server)")
	}
	defer l.Stop()

	info := l.Info()
	if info.Method != "x11_xgrabkey" {
		t.Errorf("Info().Method = %q, want %q", info.Method, "x11_xgrabkey")
	}
}

func TestX11Listener_Register(t *testing.T) {
	if os.Getenv("DISPLAY") == "" {
		t.Skip("DISPLAY not set")
	}

	l := newX11Listener()
	if l == nil {
		t.Skip("X11 connection failed (no real X server)")
	}
	defer l.Stop()

	warning, err := l.Register(Key{Modifiers: ModCtrl, Code: "space"})
	if err != nil {
		t.Fatalf("Register error: %v", err)
	}
	// ctrl+space is a known conflict
	if warning == "" {
		t.Error("expected conflict warning for ctrl+space")
	}
}

func TestX11Listener_StopIdempotent(t *testing.T) {
	if os.Getenv("DISPLAY") == "" {
		t.Skip("DISPLAY not set")
	}

	l := newX11Listener()
	if l == nil {
		t.Skip("X11 connection failed (no real X server)")
	}

	if err := l.Stop(); err != nil {
		t.Errorf("first Stop() error: %v", err)
	}
	if err := l.Stop(); err != nil {
		t.Errorf("second Stop() error: %v", err)
	}
}

func TestKeyToKeySym_KnownKeys(t *testing.T) {
	// Verify key mapping for common keys.
	tests := []struct {
		code string
		want bool // true = should map to a valid keysym
	}{
		{"space", true},
		{"return", true},
		{"escape", true},
		{"tab", true},
		{"a", true},
		{"z", true},
		{"0", true},
		{"9", true},
		{"f1", true},
		{"f12", true},
		{"nonexistent_key_xyz", false},
	}
	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			sym := keyToKeySym(tt.code)
			if tt.want && sym == 0 {
				t.Errorf("keyToKeySym(%q) = NoSymbol, want valid", tt.code)
			}
			if !tt.want && sym != 0 {
				t.Errorf("keyToKeySym(%q) = %d, want NoSymbol", tt.code, sym)
			}
		})
	}
}

func TestModifiersToX11Mask(t *testing.T) {
	tests := []struct {
		mod  Modifier
		desc string
	}{
		{ModCtrl, "ctrl"},
		{ModAlt, "alt"},
		{ModShift, "shift"},
		{ModSuper, "super"},
		{ModCtrl | ModAlt, "ctrl+alt"},
		{0, "none"},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			mask := modifiersToX11Mask(tt.mod)
			if tt.mod == 0 && mask != 0 {
				t.Errorf("expected 0 mask for no modifiers, got %d", mask)
			}
			if tt.mod != 0 && mask == 0 {
				t.Errorf("expected non-zero mask for %s", tt.desc)
			}
		})
	}
}
