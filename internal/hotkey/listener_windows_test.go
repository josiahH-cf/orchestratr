//go:build windows

package hotkey

import (
	"testing"
)

func TestWindowsListener_Info(t *testing.T) {
	l := newWindowsListener()
	info := l.Info()
	if info.OS != "windows" {
		t.Errorf("Info().OS = %q, want %q", info.OS, "windows")
	}
	if info.Method != "registerhotkey" {
		t.Errorf("Info().Method = %q, want %q", info.Method, "registerhotkey")
	}
}

func TestWindowsListener_Register(t *testing.T) {
	l := newWindowsListener()

	// ctrl+space is a known conflict.
	warning, err := l.Register(Key{Modifiers: ModCtrl, Code: "space"})
	if err != nil {
		t.Fatalf("Register error: %v", err)
	}
	if warning == "" {
		t.Error("expected conflict warning for ctrl+space")
	}

	// Check that the leader key was stored.
	if !l.leader.Equal(Key{Modifiers: ModCtrl, Code: "space"}) {
		t.Errorf("leader = %v, want ctrl+space", l.leader)
	}
}

func TestWindowsListener_RegisterUnknownKey(t *testing.T) {
	l := newWindowsListener()

	_, err := l.Register(Key{Code: "nonexistent_key"})
	if err == nil {
		t.Error("expected error for unknown key, got nil")
	}
}

func TestWindowsListener_StopIdempotent(t *testing.T) {
	l := newWindowsListener()
	if err := l.Stop(); err != nil {
		t.Errorf("first Stop() error: %v", err)
	}
	if err := l.Stop(); err != nil {
		t.Errorf("second Stop() error: %v", err)
	}
}

func TestWindowsListener_PlatformListenerIsWindows(t *testing.T) {
	l := NewPlatformListener()
	info := l.Info()
	if info.OS != "windows" {
		t.Errorf("Info().OS = %q, want %q", info.OS, "windows")
	}
	if info.Method != "registerhotkey" {
		t.Errorf("Info().Method = %q, want %q", info.Method, "registerhotkey")
	}
}

func TestKeyToVK_MappingCoverage(t *testing.T) {
	tests := []struct {
		code string
		want uint32
	}{
		{"space", _VK_SPACE},
		{"return", _VK_RETURN},
		{"enter", _VK_RETURN},
		{"escape", _VK_ESCAPE},
		{"esc", _VK_ESCAPE},
		{"tab", _VK_TAB},
		{"backspace", _VK_BACK},
		{"delete", _VK_DELETE},
		{"del", _VK_DELETE},
		{"insert", _VK_INSERT},
		{"home", _VK_HOME},
		{"end", _VK_END},
		{"pageup", _VK_PRIOR},
		{"pagedown", _VK_NEXT},
		{"up", _VK_UP},
		{"down", _VK_DOWN},
		{"left", _VK_LEFT},
		{"right", _VK_RIGHT},
		{"f1", _VK_F1},
		{"f12", _VK_F12},
		{"a", 0x41},
		{"z", 0x5A},
		{"0", 0x30},
		{"9", 0x39},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			got, ok := keyToVK(tt.code)
			if !ok {
				t.Fatalf("keyToVK(%q) returned not-ok", tt.code)
			}
			if got != tt.want {
				t.Errorf("keyToVK(%q) = 0x%X, want 0x%X", tt.code, got, tt.want)
			}
		})
	}
}

func TestKeyToVK_Unknown(t *testing.T) {
	_, ok := keyToVK("nonexistent_key")
	if ok {
		t.Error("expected not-ok for unknown key")
	}
}

func TestModifiersToWin32_Mapping(t *testing.T) {
	tests := []struct {
		mod  Modifier
		want uint32
	}{
		{ModCtrl, _MOD_CONTROL},
		{ModAlt, _MOD_ALT},
		{ModShift, _MOD_SHIFT},
		{ModSuper, _MOD_WIN},
		{ModCtrl | ModAlt, _MOD_CONTROL | _MOD_ALT},
		{ModCtrl | ModShift | ModAlt | ModSuper, _MOD_CONTROL | _MOD_SHIFT | _MOD_ALT | _MOD_WIN},
		{0, 0},
	}

	for _, tt := range tests {
		got := modifiersToWin32(tt.mod)
		if got != tt.want {
			t.Errorf("modifiersToWin32(%v) = 0x%X, want 0x%X", tt.mod, got, tt.want)
		}
	}
}

func TestVKToKey_RoundTrip(t *testing.T) {
	// For each known VK code, converting back should produce a valid key.
	codes := []string{
		"space", "return", "escape", "tab", "backspace",
		"delete", "insert", "home", "end", "pageup", "pagedown",
		"up", "down", "left", "right",
		"f1", "f2", "f3", "f4", "f5", "f6",
		"f7", "f8", "f9", "f10", "f11", "f12",
		"a", "m", "z", "0", "5", "9",
	}

	for _, code := range codes {
		vk, ok := keyToVK(code)
		if !ok {
			t.Errorf("keyToVK(%q) not ok", code)
			continue
		}
		back := vkToKeyCode(vk)
		if back == "" {
			t.Errorf("vkToKeyCode(0x%X) returned empty for %q", vk, code)
		}
	}
}
