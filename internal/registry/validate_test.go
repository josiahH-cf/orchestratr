package registry

import (
	"strings"
	"testing"
)

func TestValidateConfig_ValidMinimal(t *testing.T) {
	cfg := &Config{
		Apps: []AppEntry{
			{Name: "app1", Chord: "a", Command: "echo hi", Environment: "native"},
		},
	}
	errs := ValidateConfig(cfg)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidateConfig_EmptyApps(t *testing.T) {
	cfg := &Config{Apps: []AppEntry{}}
	errs := ValidateConfig(cfg)
	if len(errs) != 0 {
		t.Errorf("expected no errors for empty apps, got %v", errs)
	}
}

func TestValidateConfig_MissingRequiredFields(t *testing.T) {
	tests := []struct {
		name    string
		app     AppEntry
		wantErr string
	}{
		{
			name:    "missing name",
			app:     AppEntry{Chord: "a", Command: "echo hi"},
			wantErr: "name is required",
		},
		{
			name:    "missing chord",
			app:     AppEntry{Name: "app1", Command: "echo hi"},
			wantErr: "chord is required",
		},
		{
			name:    "missing command",
			app:     AppEntry{Name: "app1", Chord: "a"},
			wantErr: "command is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Apps: []AppEntry{tt.app}}
			errs := ValidateConfig(cfg)
			if len(errs) == 0 {
				t.Fatal("expected errors, got none")
			}
			found := false
			for _, err := range errs {
				if strings.Contains(err.Error(), tt.wantErr) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected error containing %q, got %v", tt.wantErr, errs)
			}
		})
	}
}

func TestValidateConfig_AllFieldsMissing(t *testing.T) {
	cfg := &Config{Apps: []AppEntry{{}}}
	errs := ValidateConfig(cfg)
	if len(errs) < 3 {
		t.Errorf("expected at least 3 errors (name, chord, command), got %d: %v", len(errs), errs)
	}
}

func TestValidateConfig_DuplicateChords(t *testing.T) {
	cfg := &Config{
		Apps: []AppEntry{
			{Name: "app1", Chord: "a", Command: "cmd1"},
			{Name: "app2", Chord: "a", Command: "cmd2"},
		},
	}
	errs := ValidateConfig(cfg)
	found := false
	for _, err := range errs {
		if strings.Contains(err.Error(), "duplicate chord") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected duplicate chord error, got %v", errs)
	}
}

func TestValidateConfig_DuplicateChordsCaseInsensitive(t *testing.T) {
	cfg := &Config{
		Apps: []AppEntry{
			{Name: "app1", Chord: "A", Command: "cmd1"},
			{Name: "app2", Chord: "a", Command: "cmd2"},
		},
	}
	errs := ValidateConfig(cfg)
	found := false
	for _, err := range errs {
		if strings.Contains(err.Error(), "duplicate chord") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected duplicate chord error for case-insensitive match, got %v", errs)
	}
}

func TestValidateConfig_ReservedChords(t *testing.T) {
	tests := []struct {
		chord string
	}{
		{"?"},
		{"space"},
	}

	for _, tt := range tests {
		t.Run("reserved_"+tt.chord, func(t *testing.T) {
			cfg := &Config{
				Apps: []AppEntry{
					{Name: "app1", Chord: tt.chord, Command: "cmd1"},
				},
			}
			errs := ValidateConfig(cfg)
			found := false
			for _, err := range errs {
				if strings.Contains(err.Error(), "reserved") {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected reserved chord error for %q, got %v", tt.chord, errs)
			}
		})
	}
}

func TestValidateChord_ValidChords(t *testing.T) {
	validChords := []string{
		"a", "z", "1", "9", "e", "!", "/",
		"f1", "F1", "f12", "tab", "enter", "escape",
	}
	for _, chord := range validChords {
		t.Run(chord, func(t *testing.T) {
			if err := validateChord(chord); err != nil {
				t.Errorf("validateChord(%q) = %v, want nil", chord, err)
			}
		})
	}
}

func TestValidateChord_InvalidChords(t *testing.T) {
	invalidChords := []string{
		"ab",     // multi-char, not a key name
		"ctrl",   // not in well-known keys
		"alt",    // not in well-known keys
		"foobar", // random string
		"ctrl+a", // modifier combo, not a chord
	}
	for _, chord := range invalidChords {
		t.Run(chord, func(t *testing.T) {
			if err := validateChord(chord); err == nil {
				t.Errorf("validateChord(%q) = nil, want error", chord)
			}
		})
	}
}

func TestValidateEnvironment(t *testing.T) {
	tests := []struct {
		env     string
		wantErr bool
	}{
		{"native", false},
		{"wsl", false},
		{"wsl:Ubuntu-22.04", false},
		{"wsl:Debian", false},
		{"wsl:", true},   // empty distro
		{"docker", true}, // unknown
		{"", false},      // empty is handled separately (optional field)
		{"NATIVE", true}, // case-sensitive
	}

	for _, tt := range tests {
		t.Run(tt.env, func(t *testing.T) {
			err := validateEnvironment(tt.env)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateEnvironment(%q) error = %v, wantErr %v", tt.env, err, tt.wantErr)
			}
		})
	}
}

func TestValidateConfig_InvalidEnvironment(t *testing.T) {
	cfg := &Config{
		Apps: []AppEntry{
			{Name: "app1", Chord: "a", Command: "cmd1", Environment: "docker"},
		},
	}
	errs := ValidateConfig(cfg)
	found := false
	for _, err := range errs {
		if strings.Contains(err.Error(), "not valid") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected environment error, got %v", errs)
	}
}

func TestValidateConfig_MultipleErrors(t *testing.T) {
	cfg := &Config{
		Apps: []AppEntry{
			{Name: "", Chord: "", Command: ""},
			{Name: "app2", Chord: "a", Command: "cmd"},
			{Name: "app3", Chord: "a", Command: "cmd"},
		},
	}
	errs := ValidateConfig(cfg)
	// First app: missing name, chord, command = 3 errors
	// Second+third: duplicate chord = 1 error
	// Total: at least 4
	if len(errs) < 4 {
		t.Errorf("expected at least 4 errors, got %d: %v", len(errs), errs)
	}
}
