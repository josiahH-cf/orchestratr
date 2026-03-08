package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/josiahH-cf/orchestratr/internal/api"
	"github.com/josiahH-cf/orchestratr/internal/daemon"
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
	if !strings.Contains(stdout.String(), Version) {
		t.Errorf("output = %q, want version string %q", stdout.String(), Version)
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

func TestRun_StartForegroundFlag(t *testing.T) {
	// Verify that "start --foreground" is parsed without error up to the
	// lock-acquisition stage (which rejects because we write our own PID).
	// --force bypasses the WSL2 guard so this test works on WSL2 hosts.
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "orchestratr.pid")
	t.Setenv("ORCHESTRATR_LOCK_PATH", lockPath)

	// Write our own PID to simulate a running daemon so start exits fast.
	if err := os.WriteFile(lockPath, []byte(strconv.Itoa(os.Getpid())), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	err := run([]string{"start", "--foreground", "--force"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected lock conflict error")
	}
	if !strings.Contains(err.Error(), "another instance") {
		t.Errorf("error = %q, want 'another instance'", err.Error())
	}
}

func TestRun_StatusShowsPort(t *testing.T) {
	// When the daemon is running and a port file exists, status should
	// show both PID and port.
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "orchestratr.pid")
	portPath := filepath.Join(dir, "port")

	t.Setenv("ORCHESTRATR_LOCK_PATH", lockPath)
	t.Setenv("ORCHESTRATR_PORT_PATH", portPath)

	// Write our own PID and a port file.
	if err := os.WriteFile(lockPath, []byte(strconv.Itoa(os.Getpid())), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(portPath, []byte("9876"), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	err := run([]string{"status"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run(status) error = %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "running") {
		t.Errorf("output = %q, want 'running'", out)
	}
	if !strings.Contains(out, "port 9876") {
		t.Errorf("output = %q, want 'port 9876'", out)
	}
}

func TestRun_StatusWithoutPortFile(t *testing.T) {
	// When the daemon is running but no port file exists, status should
	// still show PID without error.
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "orchestratr.pid")

	t.Setenv("ORCHESTRATR_LOCK_PATH", lockPath)
	t.Setenv("ORCHESTRATR_PORT_PATH", filepath.Join(dir, "no-such-port"))

	if err := os.WriteFile(lockPath, []byte(strconv.Itoa(os.Getpid())), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	err := run([]string{"status"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run(status) error = %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "running") {
		t.Errorf("output = %q, want 'running'", out)
	}
	// Should NOT contain "port" when file is missing.
	if strings.Contains(out, "port") {
		t.Errorf("output = %q, should not contain 'port' without port file", out)
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

func TestRun_StatusStalePIDCleansArtifacts(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "orchestratr.pid")
	portPath := filepath.Join(dir, "port")

	t.Setenv("ORCHESTRATR_LOCK_PATH", lockPath)
	t.Setenv("ORCHESTRATR_PORT_PATH", portPath)

	if err := os.WriteFile(lockPath, []byte("999999"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(portPath, []byte("9876"), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	err := run([]string{"status"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run(status) error = %v", err)
	}

	if !strings.Contains(stdout.String(), "stale PID") {
		t.Errorf("output = %q, want stale PID message", stdout.String())
	}

	if _, statErr := os.Stat(lockPath); !os.IsNotExist(statErr) {
		t.Errorf("lock file should be removed, stat error = %v", statErr)
	}
	if _, statErr := os.Stat(portPath); !os.IsNotExist(statErr) {
		t.Errorf("port file should be removed, stat error = %v", statErr)
	}
}

func TestRun_StatusInvalidPIDCleansArtifacts(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "orchestratr.pid")
	portPath := filepath.Join(dir, "port")

	t.Setenv("ORCHESTRATR_LOCK_PATH", lockPath)
	t.Setenv("ORCHESTRATR_PORT_PATH", portPath)

	if err := os.WriteFile(lockPath, []byte("not-a-number"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(portPath, []byte("9876"), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	err := run([]string{"status"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run(status) error = %v", err)
	}

	if !strings.Contains(stdout.String(), "not running") {
		t.Errorf("output = %q, want not running message", stdout.String())
	}

	if _, statErr := os.Stat(lockPath); !os.IsNotExist(statErr) {
		t.Errorf("lock file should be removed, stat error = %v", statErr)
	}
	if _, statErr := os.Stat(portPath); !os.IsNotExist(statErr) {
		t.Errorf("port file should be removed, stat error = %v", statErr)
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
	err := run([]string{"start", "--foreground", "--force"}, &stdout, &stderr)
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

func TestRun_ReloadNotRunning(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("ORCHESTRATR_PORT_PATH", filepath.Join(dir, "no-such-port"))

	var stdout, stderr bytes.Buffer
	err := run([]string{"reload"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("run(reload) should error when daemon is not running")
	}
	if !strings.Contains(err.Error(), "not running") {
		t.Errorf("error = %q, want 'not running'", err.Error())
	}
}

func TestRun_ReloadSuccess(t *testing.T) {
	// Create a config file for the reload to load.
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yml")
	content := `leader_key: ctrl+space
apps:
  - name: testapp
    chord: t
    command: echo hello
    environment: native
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, _ := registry.LoadAndValidate(cfgPath)
	reg := registry.NewRegistry(*cfg)

	reloadFn := func() (*registry.Config, error) {
		c, err := registry.LoadAndValidate(cfgPath)
		if err != nil {
			return nil, err
		}
		reg.Swap(*c)
		return c, nil
	}

	// Start an API server.
	srv := api.NewServer(0, "v0.0.0-dev", reg, reloadFn)
	go func() { _ = srv.Start() }()
	if !srv.WaitReady(2) {
		t.Fatal("API server did not become ready")
	}
	defer srv.Stop()

	// Write port file.
	portFile := filepath.Join(dir, "port")
	if err := daemon.WritePortFile(portFile, srv.Port()); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ORCHESTRATR_PORT_PATH", portFile)

	var stdout, stderr bytes.Buffer
	err := run([]string{"reload"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run(reload) error = %v", err)
	}
	if !strings.Contains(stdout.String(), "config reloaded") {
		t.Errorf("output = %q, want 'config reloaded'", stdout.String())
	}
	if !strings.Contains(stdout.String(), "1 apps") {
		t.Errorf("output = %q, want '1 apps'", stdout.String())
	}
}

func TestRun_ReloadValidationError(t *testing.T) {
	dir := t.TempDir()

	// Start with valid config.
	cfgPath := filepath.Join(dir, "config.yml")
	validYAML := `leader_key: ctrl+space
apps:
  - name: app1
    chord: a
    command: cmd1
    environment: native
`
	if err := os.WriteFile(cfgPath, []byte(validYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, _ := registry.LoadAndValidate(cfgPath)
	reg := registry.NewRegistry(*cfg)

	reloadFn := func() (*registry.Config, error) {
		c, err := registry.LoadAndValidate(cfgPath)
		if err != nil {
			return nil, err
		}
		reg.Swap(*c)
		return c, nil
	}

	srv := api.NewServer(0, "v0.0.0-dev", reg, reloadFn)
	go func() { _ = srv.Start() }()
	if !srv.WaitReady(2) {
		t.Fatal("API server did not become ready")
	}
	defer srv.Stop()

	portFile := filepath.Join(dir, "port")
	if err := daemon.WritePortFile(portFile, srv.Port()); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ORCHESTRATR_PORT_PATH", portFile)

	// Now write an invalid config (duplicate chords).
	badYAML := `leader_key: ctrl+space
apps:
  - name: app1
    chord: a
    command: cmd1
  - name: app2
    chord: a
    command: cmd2
`
	if err := os.WriteFile(cfgPath, []byte(badYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	err := run([]string{"reload"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("run(reload) should error on validation failure")
	}
	if !strings.Contains(err.Error(), "reload failed") {
		t.Errorf("error = %q, want 'reload failed'", err.Error())
	}
}

func TestVerboseEvent_Format(t *testing.T) {
	var buf bytes.Buffer
	emitVerboseEvent(&buf, "leader_key_pressed", nil)

	out := buf.String()
	if !strings.Contains(out, "EVENT") {
		t.Errorf("output = %q, want EVENT prefix", out)
	}
	if !strings.Contains(out, "leader_key_pressed") {
		t.Errorf("output = %q, want event type", out)
	}
}

func TestVerboseEvent_WithFields(t *testing.T) {
	var buf bytes.Buffer
	emitVerboseEvent(&buf, "chord_received", map[string]string{
		"chord": "e",
	})

	out := buf.String()
	if !strings.Contains(out, "chord=e") {
		t.Errorf("output = %q, want chord=e", out)
	}
}

func TestVerboseEvent_AppLaunching(t *testing.T) {
	var buf bytes.Buffer
	emitVerboseEvent(&buf, "app_launching", map[string]string{
		"name":    "espansr",
		"command": "espansr gui",
		"env":     "wsl",
	})

	out := buf.String()
	if !strings.Contains(out, "app_launching") {
		t.Errorf("output = %q, want app_launching", out)
	}
	if !strings.Contains(out, "name=espansr") {
		t.Errorf("output = %q, want name=espansr", out)
	}
}

func TestVerboseEvent_Timestamp(t *testing.T) {
	var buf bytes.Buffer
	emitVerboseEvent(&buf, "test_event", nil)

	out := buf.String()
	// Should start with a timestamp in brackets.
	if len(out) < 2 || out[0] != '[' {
		t.Errorf("output = %q, want timestamp in brackets", out)
	}
	// Verify year is present (basic sanity check).
	year := fmt.Sprintf("%d", time.Now().Year())
	if !strings.Contains(out, year) {
		t.Errorf("output = %q, want current year %s", out, year)
	}
}

func TestRun_StartVerboseFlag(t *testing.T) {
	// Verify that "start --foreground --verbose" is parsed correctly.
	// It will fail at lock acquisition (same as existing foreground test),
	// but we verify the flag is accepted without error.
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "orchestratr.pid")
	t.Setenv("ORCHESTRATR_LOCK_PATH", lockPath)

	// Write our own PID to simulate a running daemon so start exits fast.
	if err := os.WriteFile(lockPath, []byte(strconv.Itoa(os.Getpid())), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	err := run([]string{"start", "--foreground", "--verbose", "--force"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected lock conflict error")
	}
	// The error should be about "another instance", not about an unknown flag.
	if !strings.Contains(err.Error(), "another instance") {
		t.Errorf("error = %q, want 'another instance' (not flag parse error)", err.Error())
	}
}

func TestRun_StartWSL2Guard(t *testing.T) {
	// This test verifies WSL2 guard behavior. Results depend on the
	// test environment:
	// - On WSL2: guard fires, returns error mentioning WSL2
	// - On non-WSL2: guard does not fire, reaches lock stage
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "orchestratr.pid")
	t.Setenv("ORCHESTRATR_LOCK_PATH", lockPath)

	if err := os.WriteFile(lockPath, []byte(strconv.Itoa(os.Getpid())), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	err := run([]string{"start", "--foreground"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected an error")
	}

	isWSL2 := strings.Contains(strings.ToLower(readProcVersion(t)), "microsoft")
	if isWSL2 {
		if !strings.Contains(err.Error(), "WSL2") {
			t.Errorf("on WSL2: error = %q, want WSL2 refusal", err.Error())
		}
		if !strings.Contains(stderr.String(), "WSL2") {
			t.Errorf("on WSL2: stderr should contain WSL2 warning")
		}
	} else {
		if strings.Contains(err.Error(), "WSL2") {
			t.Errorf("on non-WSL2: error = %q, should not mention WSL2", err.Error())
		}
	}
}

func readProcVersion(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile("/proc/version")
	if err != nil {
		return "" // not Linux or can't read — assume non-WSL2
	}
	return string(data)
}

func TestRun_StartForceBypassesWSL2Guard(t *testing.T) {
	// With --force, start should proceed past the WSL2 guard even if
	// IsWSL2() were true. On a non-WSL2 system this just confirms
	// --force does not cause a parse error. The real WSL2 test
	// requires a WSL2 environment.
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "orchestratr.pid")
	t.Setenv("ORCHESTRATR_LOCK_PATH", lockPath)

	if err := os.WriteFile(lockPath, []byte(strconv.Itoa(os.Getpid())), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	err := run([]string{"start", "--foreground", "--force"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected lock conflict error")
	}
	// Should be a lock error, not a flag error.
	if !strings.Contains(err.Error(), "another instance") {
		t.Errorf("error = %q, want 'another instance'", err.Error())
	}
}
