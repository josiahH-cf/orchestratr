package registry

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad_ValidConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")

	content := `
leader_key: "ctrl+space"
chord_timeout_ms: 2000
api_port: 9876
log_level: info

apps:
  - name: espansr
    chord: e
    command: "espansr gui"
    environment: wsl
    description: "Espanso template manager"
    ready_cmd: "espansr status --json"
    ready_timeout_ms: 3000
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

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
	if len(cfg.Apps) != 1 {
		t.Fatalf("Apps len = %d, want 1", len(cfg.Apps))
	}

	app := cfg.Apps[0]
	if app.Name != "espansr" {
		t.Errorf("app.Name = %q, want %q", app.Name, "espansr")
	}
	if app.Chord != "e" {
		t.Errorf("app.Chord = %q, want %q", app.Chord, "e")
	}
	if app.Environment != "wsl" {
		t.Errorf("app.Environment = %q, want %q", app.Environment, "wsl")
	}
	if app.ReadyCmd != "espansr status --json" {
		t.Errorf("app.ReadyCmd = %q, want %q", app.ReadyCmd, "espansr status --json")
	}
	if app.ReadyTimeoutMs != 3000 {
		t.Errorf("app.ReadyTimeoutMs = %d, want %d", app.ReadyTimeoutMs, 3000)
	}
}

func TestLoad_EmptyApps(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")

	content := `
leader_key: "ctrl+space"
apps: []
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(cfg.Apps) != 0 {
		t.Errorf("Apps len = %d, want 0", len(cfg.Apps))
	}
}

func TestLoad_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")

	if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Empty file should produce default values.
	if cfg.LeaderKey != "ctrl+space" {
		t.Errorf("LeaderKey = %q, want default %q", cfg.LeaderKey, "ctrl+space")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yml")
	if err == nil {
		t.Fatal("Load() should return error for nonexistent file")
	}
	if !strings.Contains(err.Error(), "reading config") {
		t.Errorf("error = %q, want it to contain 'reading config'", err.Error())
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")

	content := `
leader_key: "ctrl+space"
apps:
  - name: "missing closing quote
    chord: e
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("Load() should return error for invalid YAML")
	}
	if !strings.Contains(err.Error(), "parsing config") {
		t.Errorf("error = %q, want it to contain 'parsing config'", err.Error())
	}
}

func TestLoadAndValidate_Valid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")

	content := `
apps:
  - name: app1
    chord: a
    command: "echo hi"
    environment: native
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadAndValidate(path)
	if err != nil {
		t.Fatalf("LoadAndValidate() error = %v", err)
	}
	if len(cfg.Apps) != 1 {
		t.Errorf("Apps len = %d, want 1", len(cfg.Apps))
	}
}

func TestLoadAndValidate_ValidationError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")

	content := `
apps:
  - name: app1
    chord: a
    command: "cmd1"
  - name: app2
    chord: a
    command: "cmd2"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadAndValidate(path)
	if err == nil {
		t.Fatal("LoadAndValidate() should return error for duplicate chords")
	}
	if !strings.Contains(err.Error(), "duplicate chord") {
		t.Errorf("error = %q, want it to contain 'duplicate chord'", err.Error())
	}
}

func TestLoadAndValidate_MissingFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")

	content := `
apps:
  - description: "no name, chord, or command"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadAndValidate(path)
	if err == nil {
		t.Fatal("LoadAndValidate() should return error for missing fields")
	}
	if !strings.Contains(err.Error(), "name is required") {
		t.Errorf("error should mention missing name, got %q", err.Error())
	}
}

func TestDefaultConfigPath(t *testing.T) {
	path := DefaultConfigPath()
	if path == "" {
		t.Fatal("DefaultConfigPath() returned empty string")
	}
	if !strings.HasSuffix(path, filepath.Join("orchestratr", "config.yml")) {
		t.Errorf("DefaultConfigPath() = %q, want suffix %q", path, filepath.Join("orchestratr", "config.yml"))
	}
}

func TestEnsureDefaults_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "config.yml")

	result, err := EnsureDefaults(path)
	if err != nil {
		t.Fatalf("EnsureDefaults() error = %v", err)
	}
	if result != path {
		t.Errorf("EnsureDefaults() = %q, want %q", result, path)
	}

	// File should exist.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading created file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "leader_key") {
		t.Error("default config should contain leader_key")
	}
	if !strings.Contains(content, "ctrl+space") {
		t.Error("default config should contain ctrl+space")
	}

	// Should be valid YAML that loads successfully.
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() on default config: %v", err)
	}
	if cfg.LeaderKey != "ctrl+space" {
		t.Errorf("LeaderKey = %q, want %q", cfg.LeaderKey, "ctrl+space")
	}
}

func TestEnsureDefaults_ExistingFileUntouched(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")

	original := "leader_key: custom\n"
	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := EnsureDefaults(path)
	if err != nil {
		t.Fatalf("EnsureDefaults() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != original {
		t.Errorf("EnsureDefaults modified existing file: %q", string(data))
	}
}

func TestLoad_DefaultsAppliedForMissingFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")

	// Only set apps, leaving all other fields to defaults.
	content := `
apps:
  - name: test
    chord: t
    command: "echo test"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Defaults should be applied for unset fields.
	if cfg.LeaderKey != "ctrl+space" {
		t.Errorf("LeaderKey = %q, want default %q", cfg.LeaderKey, "ctrl+space")
	}
	if cfg.APIPort != 9876 {
		t.Errorf("APIPort = %d, want default %d", cfg.APIPort, 9876)
	}
	if cfg.ChordTimeoutMs != 2000 {
		t.Errorf("ChordTimeoutMs = %d, want default %d", cfg.ChordTimeoutMs, 2000)
	}
}

func TestLoad_NormalizesEmptyEnvironment(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")

	content := `
apps:
  - name: app1
    chord: a
    command: "echo hi"
  - name: app2
    chord: b
    command: "echo bye"
    environment: wsl
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Apps[0].Environment != "native" {
		t.Errorf("Apps[0].Environment = %q, want %q", cfg.Apps[0].Environment, "native")
	}
	if cfg.Apps[1].Environment != "wsl" {
		t.Errorf("Apps[1].Environment = %q, want %q (should be preserved)", cfg.Apps[1].Environment, "wsl")
	}
}
