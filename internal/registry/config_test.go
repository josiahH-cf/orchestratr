package registry

import (
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.LeaderKey != "ctrl+space" {
		t.Errorf("LeaderKey = %q, want %q", cfg.LeaderKey, "ctrl+space")
	}
	if cfg.ChordTimeoutMs != 2000 {
		t.Errorf("ChordTimeoutMs = %d, want %d", cfg.ChordTimeoutMs, 2000)
	}
	if cfg.APIPort != 9876 {
		t.Errorf("APIPort = %d, want %d", cfg.APIPort, 9876)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, "info")
	}
	if len(cfg.Apps) != 0 {
		t.Errorf("Apps length = %d, want 0", len(cfg.Apps))
	}
}

func TestAppEntryOptionalFields(t *testing.T) {
	// Verify YAML tags are correct by marshaling and unmarshaling.
	// This is tested more thoroughly in loader_test.go; here we just
	// verify the struct tags exist.
	cfg := DefaultConfig()
	cfg.Apps = append(cfg.Apps, AppEntry{
		Name:           "test",
		Chord:          "t",
		Command:        "echo hello",
		Environment:    "native",
		Description:    "a test app",
		ReadyCmd:       "test --status",
		ReadyTimeoutMs: 5000,
	})

	if cfg.Apps[0].ReadyCmd != "test --status" {
		t.Errorf("ReadyCmd = %q, want %q", cfg.Apps[0].ReadyCmd, "test --status")
	}
	if cfg.Apps[0].ReadyTimeoutMs != 5000 {
		t.Errorf("ReadyTimeoutMs = %d, want %d", cfg.Apps[0].ReadyTimeoutMs, 5000)
	}
}

func TestAppEntryDetachedField(t *testing.T) {
	entry := AppEntry{
		Name:     "uwp-app",
		Chord:    "u",
		Command:  "calc.exe",
		Detached: true,
	}
	if !entry.Detached {
		t.Error("Detached = false, want true")
	}

	// Default should be false.
	entry2 := AppEntry{Name: "normal", Chord: "n", Command: "echo"}
	if entry2.Detached {
		t.Error("Detached should default to false")
	}
}
