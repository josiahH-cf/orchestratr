package registry

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestAppsDirPath(t *testing.T) {
	tests := []struct {
		configPath string
		want       string
	}{
		{"/home/user/.config/orchestratr/config.yml", "/home/user/.config/orchestratr/apps.d"},
		{"/tmp/test/config.yml", "/tmp/test/apps.d"},
	}
	for _, tt := range tests {
		got := AppsDirPath(tt.configPath)
		if got != tt.want {
			t.Errorf("AppsDirPath(%q) = %q, want %q", tt.configPath, got, tt.want)
		}
	}
}

func TestEnsureAppsDir_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yml")

	if err := EnsureAppsDir(cfgPath); err != nil {
		t.Fatalf("EnsureAppsDir() error = %v", err)
	}

	appsDir := AppsDirPath(cfgPath)
	info, err := os.Stat(appsDir)
	if err != nil {
		t.Fatalf("apps.d directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("apps.d is not a directory")
	}
}

func TestEnsureAppsDir_ExistingDirectory(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yml")
	appsDir := AppsDirPath(cfgPath)

	if err := os.MkdirAll(appsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Should not error when directory already exists.
	if err := EnsureAppsDir(cfgPath); err != nil {
		t.Fatalf("EnsureAppsDir() on existing dir error = %v", err)
	}
}

func TestLoadAppEntry_Valid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "espansr.yml")

	content := `name: espansr
chord: "e"
command: "espansr gui"
environment: wsl
description: "Espanso template manager"
ready_cmd: "espansr status --json"
ready_timeout_ms: 3000
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	entry, err := LoadAppEntry(path)
	if err != nil {
		t.Fatalf("LoadAppEntry() error = %v", err)
	}

	if entry.Name != "espansr" {
		t.Errorf("Name = %q, want %q", entry.Name, "espansr")
	}
	if entry.Chord != "e" {
		t.Errorf("Chord = %q, want %q", entry.Chord, "e")
	}
	if entry.Command != "espansr gui" {
		t.Errorf("Command = %q, want %q", entry.Command, "espansr gui")
	}
	if entry.Environment != "wsl" {
		t.Errorf("Environment = %q, want %q", entry.Environment, "wsl")
	}
	if entry.Source != "apps.d/espansr.yml" {
		t.Errorf("Source = %q, want %q", entry.Source, "apps.d/espansr.yml")
	}
	if entry.ReadyCmd != "espansr status --json" {
		t.Errorf("ReadyCmd = %q, want %q", entry.ReadyCmd, "espansr status --json")
	}
	if entry.ReadyTimeoutMs != 3000 {
		t.Errorf("ReadyTimeoutMs = %d, want %d", entry.ReadyTimeoutMs, 3000)
	}
}

func TestLoadAppEntry_NormalizesEnvironment(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.yml")

	content := `name: app
chord: a
command: "echo hi"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	entry, err := LoadAppEntry(path)
	if err != nil {
		t.Fatalf("LoadAppEntry() error = %v", err)
	}

	if entry.Environment != "native" {
		t.Errorf("Environment = %q, want %q (should normalize empty to native)", entry.Environment, "native")
	}
}

func TestLoadAppEntry_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yml")

	if err := os.WriteFile(path, []byte("name: \"unterminated"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadAppEntry(path)
	if err == nil {
		t.Fatal("LoadAppEntry() should error on invalid YAML")
	}
	if !strings.Contains(err.Error(), "parsing") {
		t.Errorf("error = %q, want it to contain 'parsing'", err.Error())
	}
}

func TestLoadAppEntry_FileNotFound(t *testing.T) {
	_, err := LoadAppEntry("/nonexistent/path/app.yml")
	if err == nil {
		t.Fatal("LoadAppEntry() should error on missing file")
	}
}

func TestLoadAppsDir_MultipleFiles(t *testing.T) {
	dir := t.TempDir()
	appsDir := filepath.Join(dir, "apps.d")
	if err := os.MkdirAll(appsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Write two valid manifests.
	if err := os.WriteFile(filepath.Join(appsDir, "espansr.yml"), []byte("name: espansr\nchord: e\ncommand: espansr gui\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(appsDir, "templatr.yml"), []byte("name: templatr\nchord: t\ncommand: templatr gui\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	apps, errs := LoadAppsDir(appsDir)
	if len(errs) != 0 {
		t.Errorf("unexpected errors: %v", errs)
	}
	if len(apps) != 2 {
		t.Fatalf("loaded %d apps, want 2", len(apps))
	}

	// Should be sorted alphabetically.
	if apps[0].Name != "espansr" {
		t.Errorf("apps[0].Name = %q, want %q", apps[0].Name, "espansr")
	}
	if apps[1].Name != "templatr" {
		t.Errorf("apps[1].Name = %q, want %q", apps[1].Name, "templatr")
	}
}

func TestLoadAppsDir_InvalidFileSkipped(t *testing.T) {
	dir := t.TempDir()
	appsDir := filepath.Join(dir, "apps.d")
	if err := os.MkdirAll(appsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// One valid, one invalid.
	if err := os.WriteFile(filepath.Join(appsDir, "good.yml"), []byte("name: good\nchord: g\ncommand: echo good\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(appsDir, "bad.yml"), []byte("name: \"unterminated"), 0o644); err != nil {
		t.Fatal(err)
	}

	apps, errs := LoadAppsDir(appsDir)
	if len(errs) != 1 {
		t.Errorf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if len(apps) != 1 {
		t.Fatalf("expected 1 app, got %d", len(apps))
	}
	if apps[0].Name != "good" {
		t.Errorf("app.Name = %q, want %q", apps[0].Name, "good")
	}
}

func TestLoadAppsDir_EmptyDirectory(t *testing.T) {
	dir := t.TempDir()
	appsDir := filepath.Join(dir, "apps.d")
	if err := os.MkdirAll(appsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	apps, errs := LoadAppsDir(appsDir)
	if len(errs) != 0 {
		t.Errorf("unexpected errors: %v", errs)
	}
	if len(apps) != 0 {
		t.Errorf("expected 0 apps, got %d", len(apps))
	}
}

func TestLoadAppsDir_MissingDirectory(t *testing.T) {
	apps, errs := LoadAppsDir("/nonexistent/apps.d")
	if len(errs) != 0 {
		t.Errorf("missing dir should not produce errors, got %v", errs)
	}
	if len(apps) != 0 {
		t.Errorf("expected 0 apps, got %d", len(apps))
	}
}

func TestLoadAppsDir_IgnoresNonYAML(t *testing.T) {
	dir := t.TempDir()
	appsDir := filepath.Join(dir, "apps.d")
	if err := os.MkdirAll(appsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Write a .txt file that should be ignored.
	if err := os.WriteFile(filepath.Join(appsDir, "readme.txt"), []byte("not a manifest"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(appsDir, "good.yml"), []byte("name: good\nchord: g\ncommand: echo good\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	apps, errs := LoadAppsDir(appsDir)
	if len(errs) != 0 {
		t.Errorf("unexpected errors: %v", errs)
	}
	if len(apps) != 1 {
		t.Fatalf("expected 1 app, got %d", len(apps))
	}
}

func TestLoadAppsDir_IgnoresSubdirectories(t *testing.T) {
	dir := t.TempDir()
	appsDir := filepath.Join(dir, "apps.d")
	subDir := filepath.Join(appsDir, "nested")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// File inside subdirectory should be ignored (flat scan only).
	if err := os.WriteFile(filepath.Join(subDir, "app.yml"), []byte("name: nested\nchord: n\ncommand: echo nested\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	apps, errs := LoadAppsDir(appsDir)
	if len(errs) != 0 {
		t.Errorf("unexpected errors: %v", errs)
	}
	if len(apps) != 0 {
		t.Errorf("expected 0 apps (no recursive scan), got %d", len(apps))
	}
}

func TestMergeApps_NoConflicts(t *testing.T) {
	base := []AppEntry{
		{Name: "app1", Chord: "a", Command: "cmd1", Source: "config"},
	}
	dropins := []AppEntry{
		{Name: "app2", Chord: "b", Command: "cmd2", Source: "apps.d/app2.yml"},
	}

	merged, warnings := MergeApps(base, dropins, nil)
	if len(warnings) != 0 {
		t.Errorf("unexpected warnings: %v", warnings)
	}
	if len(merged) != 2 {
		t.Fatalf("expected 2 merged apps, got %d", len(merged))
	}
}

func TestMergeApps_DuplicateNameConfigWins(t *testing.T) {
	base := []AppEntry{
		{Name: "myapp", Chord: "a", Command: "cmd1", Source: "config"},
	}
	dropins := []AppEntry{
		{Name: "myapp", Chord: "b", Command: "cmd2", Source: "apps.d/myapp.yml"},
	}

	merged, warnings := MergeApps(base, dropins, nil)
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d: %v", len(warnings), warnings)
	}
	if !strings.Contains(warnings[0].Error(), "config.yml wins") {
		t.Errorf("warning = %q, want it to mention config.yml wins", warnings[0].Error())
	}
	// Only the config app should remain.
	if len(merged) != 1 {
		t.Fatalf("expected 1 merged app, got %d", len(merged))
	}
	if merged[0].Source != "config" {
		t.Errorf("merged[0].Source = %q, want %q", merged[0].Source, "config")
	}
}

func TestMergeApps_DuplicateChordConfigWins(t *testing.T) {
	base := []AppEntry{
		{Name: "app1", Chord: "a", Command: "cmd1", Source: "config"},
	}
	dropins := []AppEntry{
		{Name: "app2", Chord: "a", Command: "cmd2", Source: "apps.d/app2.yml"},
	}

	merged, warnings := MergeApps(base, dropins, nil)
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d: %v", len(warnings), warnings)
	}
	if !strings.Contains(warnings[0].Error(), "config.yml wins") {
		t.Errorf("warning = %q, want mention of config.yml wins", warnings[0].Error())
	}
	if len(merged) != 1 {
		t.Fatalf("expected 1 merged app, got %d", len(merged))
	}
}

func TestMergeApps_DuplicateNameBetweenDropins(t *testing.T) {
	base := []AppEntry{}
	dropins := []AppEntry{
		{Name: "myapp", Chord: "a", Command: "cmd1", Source: "apps.d/one.yml"},
		{Name: "myapp", Chord: "b", Command: "cmd2", Source: "apps.d/two.yml"},
	}

	merged, warnings := MergeApps(base, dropins, nil)
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d: %v", len(warnings), warnings)
	}
	if !strings.Contains(warnings[0].Error(), "both rejected") {
		t.Errorf("warning = %q, want 'both rejected'", warnings[0].Error())
	}
	// Both should be rejected.
	if len(merged) != 0 {
		t.Errorf("expected 0 merged apps, got %d", len(merged))
	}
}

func TestMergeApps_DuplicateChordBetweenDropins(t *testing.T) {
	base := []AppEntry{}
	dropins := []AppEntry{
		{Name: "app1", Chord: "a", Command: "cmd1", Source: "apps.d/one.yml"},
		{Name: "app2", Chord: "a", Command: "cmd2", Source: "apps.d/two.yml"},
	}

	merged, warnings := MergeApps(base, dropins, nil)
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d: %v", len(warnings), warnings)
	}
	if !strings.Contains(warnings[0].Error(), "both rejected") {
		t.Errorf("warning = %q, want 'both rejected'", warnings[0].Error())
	}
	if len(merged) != 0 {
		t.Errorf("expected 0 merged apps, got %d", len(merged))
	}
}

func TestMergeApps_CaseInsensitiveConflicts(t *testing.T) {
	base := []AppEntry{
		{Name: "MyApp", Chord: "A", Command: "cmd1", Source: "config"},
	}
	dropins := []AppEntry{
		{Name: "myapp", Chord: "b", Command: "cmd2", Source: "apps.d/dup.yml"},
	}

	merged, warnings := MergeApps(base, dropins, nil)
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning for case-insensitive name conflict, got %d", len(warnings))
	}
	if len(merged) != 1 {
		t.Errorf("expected 1 merged app, got %d", len(merged))
	}
}

func TestMergeApps_LogsWarnings(t *testing.T) {
	var buf strings.Builder
	logger := log.New(&buf, "", 0)

	base := []AppEntry{
		{Name: "app1", Chord: "a", Command: "cmd1", Source: "config"},
	}
	dropins := []AppEntry{
		{Name: "app1", Chord: "b", Command: "cmd2", Source: "apps.d/dup.yml"},
	}

	MergeApps(base, dropins, logger)
	if !strings.Contains(buf.String(), "warning:") {
		t.Errorf("expected warning in log output, got %q", buf.String())
	}
}

func TestLoadWithDropins_ConfigOnly(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yml")

	content := `leader_key: "ctrl+space"
apps:
  - name: app1
    chord: a
    command: "echo hi"
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadWithDropins(cfgPath, nil)
	if err != nil {
		t.Fatalf("LoadWithDropins() error = %v", err)
	}
	if len(cfg.Apps) != 1 {
		t.Fatalf("expected 1 app, got %d", len(cfg.Apps))
	}
	if cfg.Apps[0].Source != "config" {
		t.Errorf("Source = %q, want %q", cfg.Apps[0].Source, "config")
	}
}

func TestLoadWithDropins_MergesDropins(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yml")
	appsDir := filepath.Join(dir, "apps.d")

	if err := os.MkdirAll(appsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	content := `leader_key: "ctrl+space"
apps:
  - name: app1
    chord: a
    command: "echo app1"
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(appsDir, "dropin.yml"), []byte("name: app2\nchord: b\ncommand: echo app2\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadWithDropins(cfgPath, nil)
	if err != nil {
		t.Fatalf("LoadWithDropins() error = %v", err)
	}

	if len(cfg.Apps) != 2 {
		t.Fatalf("expected 2 apps, got %d", len(cfg.Apps))
	}

	// Config app first, then drop-in.
	if cfg.Apps[0].Source != "config" {
		t.Errorf("Apps[0].Source = %q, want %q", cfg.Apps[0].Source, "config")
	}
	if cfg.Apps[1].Source != "apps.d/dropin.yml" {
		t.Errorf("Apps[1].Source = %q, want %q", cfg.Apps[1].Source, "apps.d/dropin.yml")
	}
}

func TestLoadWithDropins_InvalidDropinSkipped(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yml")
	appsDir := filepath.Join(dir, "apps.d")

	if err := os.MkdirAll(appsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	content := `apps:
  - name: app1
    chord: a
    command: echo app1
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	// Invalid YAML drop-in.
	if err := os.WriteFile(filepath.Join(appsDir, "bad.yml"), []byte("name: \"unterminated"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Valid drop-in.
	if err := os.WriteFile(filepath.Join(appsDir, "good.yml"), []byte("name: good\nchord: g\ncommand: echo good\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadWithDropins(cfgPath, nil)
	if err != nil {
		t.Fatalf("LoadWithDropins() error = %v", err)
	}

	// Should have config app + valid drop-in only.
	if len(cfg.Apps) != 2 {
		t.Fatalf("expected 2 apps (config + valid drop-in), got %d", len(cfg.Apps))
	}
}

func TestLoadWithDropins_MissingAppsDir(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yml")

	content := `apps:
  - name: app1
    chord: a
    command: echo app1
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	// No apps.d/ directory — should not error.
	cfg, err := LoadWithDropins(cfgPath, nil)
	if err != nil {
		t.Fatalf("LoadWithDropins() error = %v", err)
	}
	if len(cfg.Apps) != 1 {
		t.Fatalf("expected 1 app, got %d", len(cfg.Apps))
	}
}

func TestLoadWithDropins_EmptyAppsDir(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yml")
	appsDir := filepath.Join(dir, "apps.d")

	if err := os.MkdirAll(appsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	content := `apps:
  - name: app1
    chord: a
    command: echo app1
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadWithDropins(cfgPath, nil)
	if err != nil {
		t.Fatalf("LoadWithDropins() error = %v", err)
	}
	if len(cfg.Apps) != 1 {
		t.Errorf("expected 1 app, got %d", len(cfg.Apps))
	}
}

func TestLoadWithDropins_DuplicateChordConfigWins(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yml")
	appsDir := filepath.Join(dir, "apps.d")

	if err := os.MkdirAll(appsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	content := `apps:
  - name: app1
    chord: a
    command: echo app1
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	// Drop-in with same chord should be skipped.
	if err := os.WriteFile(filepath.Join(appsDir, "conflict.yml"), []byte("name: conflict\nchord: a\ncommand: echo conflict\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadWithDropins(cfgPath, nil)
	if err != nil {
		t.Fatalf("LoadWithDropins() error = %v", err)
	}
	if len(cfg.Apps) != 1 {
		t.Fatalf("expected 1 app (config wins), got %d", len(cfg.Apps))
	}
	if cfg.Apps[0].Name != "app1" {
		t.Errorf("surviving app = %q, want %q", cfg.Apps[0].Name, "app1")
	}
}

func TestLoadWithDropins_DuplicateNameConfigWins(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yml")
	appsDir := filepath.Join(dir, "apps.d")

	if err := os.MkdirAll(appsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	content := `apps:
  - name: app1
    chord: a
    command: echo app1
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	// Drop-in with same name should be skipped.
	if err := os.WriteFile(filepath.Join(appsDir, "dup.yml"), []byte("name: app1\nchord: b\ncommand: echo other\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadWithDropins(cfgPath, nil)
	if err != nil {
		t.Fatalf("LoadWithDropins() error = %v", err)
	}
	if len(cfg.Apps) != 1 {
		t.Fatalf("expected 1 app (config wins), got %d", len(cfg.Apps))
	}
	if cfg.Apps[0].Command != "echo app1" {
		t.Errorf("surviving app command = %q, want config's command", cfg.Apps[0].Command)
	}
}

func TestLoad_SetsSourceToConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")

	content := `apps:
  - name: app1
    chord: a
    command: echo hi
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Apps[0].Source != "config" {
		t.Errorf("Source = %q, want %q", cfg.Apps[0].Source, "config")
	}
}

func TestSourceNotSerializedToYAML(t *testing.T) {
	// The Source field has yaml:"-" tag, so it should not appear in YAML output.
	entry := AppEntry{
		Name:    "test",
		Chord:   "t",
		Command: "echo test",
		Source:  "apps.d/test.yml",
	}

	data, err := yaml.Marshal(entry)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	if strings.Contains(string(data), "source") {
		t.Errorf("YAML output should not contain 'source', got: %s", string(data))
	}
}
