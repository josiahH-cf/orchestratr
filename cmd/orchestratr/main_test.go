package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/josiahH-cf/orchestratr/internal/registry"
)

func TestRun_NoArgs(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run(nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "orchestratr") {
		t.Errorf("output = %q, want it to contain 'orchestratr'", stdout.String())
	}
}

func TestRun_Version(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run([]string{"version"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "v0.0.0-dev") {
		t.Errorf("output = %q, want version string", stdout.String())
	}
}

func TestRun_UnknownCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run([]string{"bogus"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("run() should return error for unknown command")
	}
	if !strings.Contains(err.Error(), "unknown command") {
		t.Errorf("error = %q, want 'unknown command'", err.Error())
	}
}

func TestRun_ListWithApps(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")

	content := `
leader_key: "ctrl+space"
apps:
  - name: espansr
    chord: e
    command: "espansr gui"
    environment: wsl
    description: "Espanso template manager"
  - name: browser
    chord: b
    command: "firefox"
    environment: native
    description: "Web browser"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("ORCHESTRATR_CONFIG", path)

	var stdout, stderr bytes.Buffer
	err := run([]string{"list"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run(list) error = %v", err)
	}

	out := stdout.String()

	// Check header.
	if !strings.Contains(out, "NAME") {
		t.Error("output should contain NAME header")
	}
	if !strings.Contains(out, "CHORD") {
		t.Error("output should contain CHORD header")
	}
	if !strings.Contains(out, "COMMAND") {
		t.Error("output should contain COMMAND header")
	}

	// Check app entries.
	if !strings.Contains(out, "espansr") {
		t.Error("output should contain espansr")
	}
	if !strings.Contains(out, "browser") {
		t.Error("output should contain browser")
	}
	if !strings.Contains(out, "wsl") {
		t.Error("output should contain wsl environment")
	}
}

func TestRun_ListEmptyRegistry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")

	content := `
leader_key: "ctrl+space"
apps: []
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("ORCHESTRATR_CONFIG", path)

	var stdout, stderr bytes.Buffer
	err := run([]string{"list"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run(list) error = %v", err)
	}

	if !strings.Contains(stdout.String(), "No apps registered") {
		t.Errorf("output = %q, want 'No apps registered' message", stdout.String())
	}
}

func TestRun_ListMissingConfig(t *testing.T) {
	t.Setenv("ORCHESTRATR_CONFIG", "/nonexistent/path/config.yml")

	var stdout, stderr bytes.Buffer
	err := run([]string{"list"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("run(list) should return error for missing config")
	}
	if !strings.Contains(err.Error(), "config not found") {
		t.Errorf("error = %q, want 'config not found'", err.Error())
	}
}

func TestRun_ListInvalidConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")

	content := `
apps:
  - name: app1
    chord: a
    command: cmd1
  - name: app2
    chord: a
    command: cmd2
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("ORCHESTRATR_CONFIG", path)

	var stdout, stderr bytes.Buffer
	err := run([]string{"list"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("run(list) should return error for invalid config")
	}
	if !strings.Contains(err.Error(), "duplicate chord") {
		t.Errorf("error = %q, want 'duplicate chord'", err.Error())
	}
}

func TestRun_ListDefaultsEnvironment(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")

	// App without explicit environment should show "native".
	content := `
apps:
  - name: app1
    chord: a
    command: "echo hi"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("ORCHESTRATR_CONFIG", path)

	var stdout, stderr bytes.Buffer
	err := run([]string{"list"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run(list) error = %v", err)
	}

	if !strings.Contains(stdout.String(), "native") {
		t.Errorf("output = %q, want default 'native' environment", stdout.String())
	}
}

func TestRun_StatusNotRunning(t *testing.T) {
	// Point lock path to a temp dir so no PID file exists.
	dir := t.TempDir()
	t.Setenv("ORCHESTRATR_LOCK_PATH", filepath.Join(dir, "orchestratr.pid"))

	var stdout, stderr bytes.Buffer
	err := run([]string{"status"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run(status) error = %v", err)
	}
	if !strings.Contains(stdout.String(), "not running") {
		t.Errorf("output = %q, want 'not running'", stdout.String())
	}
}

func TestRun_StopNotRunning(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("ORCHESTRATR_LOCK_PATH", filepath.Join(dir, "orchestratr.pid"))

	var stdout, stderr bytes.Buffer
	err := run([]string{"stop"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("run(stop) should error when daemon is not running")
	}
	if !strings.Contains(err.Error(), "not running") {
		t.Errorf("error = %q, want 'not running'", err.Error())
	}
}

func TestRun_StartRejectsDuplicate(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "orchestratr.pid")
	t.Setenv("ORCHESTRATR_LOCK_PATH", lockPath)

	// Write our own PID to simulate a running daemon.
	if err := os.WriteFile(lockPath, []byte(strconv.Itoa(os.Getpid())), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	err := run([]string{"start"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("run(start) should error when another instance is running")
	}
	if !strings.Contains(err.Error(), "another instance") {
		t.Errorf("error = %q, want 'another instance'", err.Error())
	}
}
func TestBuildChords_ValidApps(t *testing.T) {
	apps := []registry.AppEntry{
		{Name: "espansr", Chord: "e", Command: "espansr gui"},
		{Name: "firefox", Chord: "f", Command: "firefox"},
		{Name: "terminal", Chord: "ctrl+t", Command: "kitty"},
	}
	chords := buildChords(apps)
	if len(chords) != 3 {
		t.Fatalf("got %d chords, want 3", len(chords))
	}
	if chords[0].Action != "espansr" {
		t.Errorf("chords[0].Action = %q, want %q", chords[0].Action, "espansr")
	}
	if chords[0].Key.Code != "e" {
		t.Errorf("chords[0].Key.Code = %q, want %q", chords[0].Key.Code, "e")
	}
	if chords[2].Action != "terminal" {
		t.Errorf("chords[2].Action = %q, want %q", chords[2].Action, "terminal")
	}
}

func TestBuildChords_SkipsEmptyChord(t *testing.T) {
	apps := []registry.AppEntry{
		{Name: "espansr", Chord: "e", Command: "espansr gui"},
		{Name: "nokey", Chord: "", Command: "nokey"},
	}
	chords := buildChords(apps)
	if len(chords) != 1 {
		t.Fatalf("got %d chords, want 1 (empty chord skipped)", len(chords))
	}
}

func TestBuildChords_SkipsInvalidChord(t *testing.T) {
	apps := []registry.AppEntry{
		{Name: "good", Chord: "e", Command: "echo"},
		{Name: "bad", Chord: "ctrl+shift", Command: "nope"}, // no base key
	}
	chords := buildChords(apps)
	if len(chords) != 1 {
		t.Fatalf("got %d chords, want 1 (invalid chord skipped)", len(chords))
	}
	if chords[0].Action != "good" {
		t.Errorf("remaining chord = %q, want %q", chords[0].Action, "good")
	}
}

func TestBuildChords_EmptyList(t *testing.T) {
	chords := buildChords(nil)
	if len(chords) != 0 {
		t.Errorf("got %d chords, want 0", len(chords))
	}
}

func TestRun_TriggerNotRunning(t *testing.T) {
	// Point port file to a non-existent file so trigger fails.
	dir := t.TempDir()
	t.Setenv("ORCHESTRATR_PORT_PATH", filepath.Join(dir, "no-such-port"))

	var stdout, stderr bytes.Buffer
	err := run([]string{"trigger"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("run(trigger) should error when daemon is not running")
	}
	if !strings.Contains(err.Error(), "not running") {
		t.Errorf("error = %q, want 'not running'", err.Error())
	}
}

func TestRun_UsageIncludesTrigger(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run(nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "trigger") {
		t.Errorf("usage output should mention 'trigger': %q", stdout.String())
	}
}