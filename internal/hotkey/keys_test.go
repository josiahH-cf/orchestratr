package hotkey

import "testing"

func TestParseKey_Simple(t *testing.T) {
	tests := []struct {
		input   string
		wantMod Modifier
		wantKey string
	}{
		{"a", 0, "a"},
		{"space", 0, "space"},
		{"f1", 0, "f1"},
		{"escape", 0, "escape"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			k, err := ParseKey(tt.input)
			if err != nil {
				t.Fatalf("ParseKey(%q) error: %v", tt.input, err)
			}
			if k.Modifiers != tt.wantMod {
				t.Errorf("modifiers = %v, want %v", k.Modifiers, tt.wantMod)
			}
			if k.Code != tt.wantKey {
				t.Errorf("code = %q, want %q", k.Code, tt.wantKey)
			}
		})
	}
}

func TestParseKey_WithModifiers(t *testing.T) {
	tests := []struct {
		input   string
		wantMod Modifier
		wantKey string
	}{
		{"ctrl+space", ModCtrl, "space"},
		{"Ctrl+Space", ModCtrl, "space"},
		{"CTRL+SPACE", ModCtrl, "space"},
		{"ctrl+shift+a", ModCtrl | ModShift, "a"},
		{"alt+f4", ModAlt, "f4"},
		{"super+e", ModSuper, "e"},
		{"cmd+space", ModSuper, "space"}, // macOS alias
		{"win+r", ModSuper, "r"},         // Windows alias
		{"option+tab", ModAlt, "tab"},    // macOS alias
		{"control+c", ModCtrl, "c"},      // full name alias
		{"ctrl+alt+shift+super+x", ModCtrl | ModAlt | ModShift | ModSuper, "x"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			k, err := ParseKey(tt.input)
			if err != nil {
				t.Fatalf("ParseKey(%q) error: %v", tt.input, err)
			}
			if k.Modifiers != tt.wantMod {
				t.Errorf("modifiers = %v, want %v", k.Modifiers, tt.wantMod)
			}
			if k.Code != tt.wantKey {
				t.Errorf("code = %q, want %q", k.Code, tt.wantKey)
			}
		})
	}
}

func TestParseKey_Whitespace(t *testing.T) {
	k, err := ParseKey("ctrl + space")
	if err != nil {
		t.Fatalf("ParseKey with spaces error: %v", err)
	}
	if k.Modifiers != ModCtrl || k.Code != "space" {
		t.Errorf("got %v, want ctrl+space", k)
	}
}

func TestParseKey_Errors(t *testing.T) {
	tests := []struct {
		input string
		want  string // substring of error message
	}{
		{"", "empty"},
		{"ctrl+shift", "no base key"},
		{"ctrl+a+b", "multiple base keys"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := ParseKey(tt.input)
			if err == nil {
				t.Fatalf("ParseKey(%q) expected error containing %q", tt.input, tt.want)
			}
			if !containsLower(err.Error(), tt.want) {
				t.Errorf("error = %q, want substring %q", err.Error(), tt.want)
			}
		})
	}
}

func TestKeyString(t *testing.T) {
	tests := []struct {
		key  Key
		want string
	}{
		{Key{Code: "a"}, "a"},
		{Key{Modifiers: ModCtrl, Code: "space"}, "ctrl+space"},
		{Key{Modifiers: ModCtrl | ModShift, Code: "a"}, "ctrl+shift+a"},
		{Key{Modifiers: ModCtrl | ModAlt | ModShift | ModSuper, Code: "x"}, "ctrl+alt+shift+super+x"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.key.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestKeyEqual(t *testing.T) {
	a := Key{Modifiers: ModCtrl, Code: "space"}
	b := Key{Modifiers: ModCtrl, Code: "space"}
	c := Key{Modifiers: ModAlt, Code: "space"}
	d := Key{Modifiers: ModCtrl, Code: "a"}

	if !a.Equal(b) {
		t.Error("a should equal b")
	}
	if a.Equal(c) {
		t.Error("a should not equal c (different modifier)")
	}
	if a.Equal(d) {
		t.Error("a should not equal d (different code)")
	}
}

func TestParseKeyRoundTrip(t *testing.T) {
	inputs := []string{"ctrl+space", "alt+f4", "shift+a", "super+e", "ctrl+alt+shift+super+x"}
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			k, err := ParseKey(input)
			if err != nil {
				t.Fatalf("ParseKey(%q) error: %v", input, err)
			}
			reparsed, err := ParseKey(k.String())
			if err != nil {
				t.Fatalf("ParseKey(%q) round-trip error: %v", k.String(), err)
			}
			if !k.Equal(reparsed) {
				t.Errorf("round-trip mismatch: %v != %v", k, reparsed)
			}
		})
	}
}

func TestCheckConflicts(t *testing.T) {
	// Known conflict
	k, _ := ParseKey("ctrl+space")
	warn := CheckConflicts(k)
	if warn == "" {
		t.Error("expected conflict warning for ctrl+space")
	}
	if !containsLower(warn, "autocomplete") && !containsLower(warn, "input method") {
		t.Errorf("warning = %q, expected reference to autocomplete or input method", warn)
	}

	// No conflict
	k2, _ := ParseKey("ctrl+alt+o")
	warn2 := CheckConflicts(k2)
	if warn2 != "" {
		t.Errorf("unexpected conflict warning for ctrl+alt+o: %q", warn2)
	}
}

func TestModifierString(t *testing.T) {
	tests := []struct {
		mod  Modifier
		want string
	}{
		{0, ""},
		{ModCtrl, "ctrl"},
		{ModAlt, "alt"},
		{ModCtrl | ModAlt, "ctrl+alt"},
		{ModCtrl | ModAlt | ModShift | ModSuper, "ctrl+alt+shift+super"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.mod.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

// containsLower checks if s contains substr (case-insensitive).
func containsLower(s, substr string) bool {
	s = toLower(s)
	substr = toLower(substr)
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
