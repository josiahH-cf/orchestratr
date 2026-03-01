package registry

import (
	"testing"
)

func newTestRegistry() *Registry {
	cfg := Config{
		LeaderKey:      "ctrl+space",
		ChordTimeoutMs: 2000,
		APIPort:        9876,
		Apps: []AppEntry{
			{Name: "espansr", Chord: "e", Command: "espansr gui", Environment: "wsl"},
			{Name: "browser", Chord: "b", Command: "firefox", Environment: "native"},
		},
	}
	return NewRegistry(cfg)
}

func TestNewRegistry(t *testing.T) {
	r := newTestRegistry()
	if r == nil {
		t.Fatal("NewRegistry returned nil")
	}
	if r.Len() != 2 {
		t.Errorf("Len = %d, want 2", r.Len())
	}
}

func TestRegistry_Apps(t *testing.T) {
	r := newTestRegistry()
	apps := r.Apps()
	if len(apps) != 2 {
		t.Fatalf("Apps() len = %d, want 2", len(apps))
	}
	if apps[0].Name != "espansr" {
		t.Errorf("Apps()[0].Name = %q, want %q", apps[0].Name, "espansr")
	}

	// Verify returned slice is a copy (mutations don't affect registry).
	apps[0].Name = "mutated"
	original := r.Apps()
	if original[0].Name == "mutated" {
		t.Error("Apps() returned a reference instead of a copy")
	}
}

func TestRegistry_Config(t *testing.T) {
	r := newTestRegistry()
	cfg := r.Config()
	if cfg.LeaderKey != "ctrl+space" {
		t.Errorf("Config().LeaderKey = %q, want %q", cfg.LeaderKey, "ctrl+space")
	}
	if len(cfg.Apps) != 2 {
		t.Errorf("Config().Apps len = %d, want 2", len(cfg.Apps))
	}

	// Verify returned config is a copy.
	cfg.Apps[0].Name = "mutated"
	original := r.Config()
	if original.Apps[0].Name == "mutated" {
		t.Error("Config() returned a reference instead of a copy")
	}
}

func TestRegistry_FindByChord(t *testing.T) {
	r := newTestRegistry()

	tests := []struct {
		chord string
		want  string
		found bool
	}{
		{"e", "espansr", true},
		{"E", "espansr", true}, // case-insensitive
		{"b", "browser", true},
		{"x", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.chord, func(t *testing.T) {
			app, ok := r.FindByChord(tt.chord)
			if ok != tt.found {
				t.Fatalf("FindByChord(%q) found = %v, want %v", tt.chord, ok, tt.found)
			}
			if ok && app.Name != tt.want {
				t.Errorf("FindByChord(%q).Name = %q, want %q", tt.chord, app.Name, tt.want)
			}
		})
	}
}

func TestRegistry_FindByName(t *testing.T) {
	r := newTestRegistry()

	tests := []struct {
		name  string
		chord string
		found bool
	}{
		{"espansr", "e", true},
		{"ESPANSR", "e", true}, // case-insensitive
		{"browser", "b", true},
		{"unknown", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, ok := r.FindByName(tt.name)
			if ok != tt.found {
				t.Fatalf("FindByName(%q) found = %v, want %v", tt.name, ok, tt.found)
			}
			if ok && app.Chord != tt.chord {
				t.Errorf("FindByName(%q).Chord = %q, want %q", tt.name, app.Chord, tt.chord)
			}
		})
	}
}

func TestRegistry_Swap(t *testing.T) {
	r := newTestRegistry()
	if r.Len() != 2 {
		t.Fatalf("initial Len = %d, want 2", r.Len())
	}

	newCfg := Config{
		Apps: []AppEntry{
			{Name: "only-one", Chord: "o", Command: "cmd"},
		},
	}
	r.Swap(newCfg)

	if r.Len() != 1 {
		t.Errorf("Len after Swap = %d, want 1", r.Len())
	}
	app, ok := r.FindByChord("o")
	if !ok {
		t.Fatal("FindByChord('o') not found after Swap")
	}
	if app.Name != "only-one" {
		t.Errorf("app.Name = %q, want %q", app.Name, "only-one")
	}
}

func TestRegistry_String(t *testing.T) {
	r := newTestRegistry()
	s := r.String()
	if s != "Registry(2 apps)" {
		t.Errorf("String() = %q, want %q", s, "Registry(2 apps)")
	}
}

func TestRegistry_EmptyConfig(t *testing.T) {
	r := NewRegistry(Config{})
	if r.Len() != 0 {
		t.Errorf("Len = %d, want 0", r.Len())
	}
	apps := r.Apps()
	if len(apps) != 0 {
		t.Errorf("Apps() len = %d, want 0", len(apps))
	}
	_, ok := r.FindByChord("a")
	if ok {
		t.Error("FindByChord should return false for empty registry")
	}
}
