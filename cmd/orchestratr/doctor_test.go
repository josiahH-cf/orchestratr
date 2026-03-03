package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// setupDoctorEnv creates a temp dir and points env vars at it so
// doctor checks use isolated paths. Returns the temp dir path.
func setupDoctorEnv(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("ORCHESTRATR_CONFIG", filepath.Join(dir, "config.yml"))
	t.Setenv("ORCHESTRATR_LOCK_PATH", filepath.Join(dir, "orchestratr.pid"))
	t.Setenv("ORCHESTRATR_PORT_PATH", filepath.Join(dir, "port"))
	return dir
}

func TestDoctor_DaemonNotRunning(t *testing.T) {
	dir := setupDoctorEnv(t)

	// Write a valid config so other checks pass.
	cfg := `leader_key: ctrl+space
apps:
  - name: testapp
    chord: t
    command: echo
    environment: native
`
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "apps.d"), 0o755); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	err := runDoctor(false, &stdout, &stderr)

	// Should not return a hard error — doctor reports results.
	if err != nil {
		t.Fatalf("runDoctor() error = %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "FAIL") {
		t.Errorf("output should contain FAIL for daemon check: %q", out)
	}
	if !strings.Contains(strings.ToLower(out), "daemon") {
		t.Errorf("output should mention daemon: %q", out)
	}
}

func TestDoctor_ValidConfig(t *testing.T) {
	dir := setupDoctorEnv(t)

	cfg := `leader_key: ctrl+space
apps:
  - name: testapp
    chord: t
    command: echo
    environment: native
`
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "apps.d"), 0o755); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	_ = runDoctor(false, &stdout, &stderr)

	out := stdout.String()
	if !strings.Contains(out, "Config parse") {
		t.Errorf("output should contain Config parse check: %q", out)
	}
	// Config parse and validation should pass.
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Config parse") && !strings.Contains(line, "PASS") {
			t.Errorf("Config parse should PASS: %q", line)
		}
		if strings.Contains(line, "Config valid") && !strings.Contains(line, "PASS") {
			t.Errorf("Config valid should PASS: %q", line)
		}
	}
}

func TestDoctor_InvalidYAML(t *testing.T) {
	dir := setupDoctorEnv(t)

	// Write garbage YAML.
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte("{{{{invalid"), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	_ = runDoctor(false, &stdout, &stderr)

	out := stdout.String()
	found := false
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "Config parse") && strings.Contains(line, "FAIL") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Config parse should FAIL for invalid YAML: %q", out)
	}
}

func TestDoctor_ValidationErrors(t *testing.T) {
	dir := setupDoctorEnv(t)

	// Duplicate chords — valid YAML but fails validation.
	cfg := `leader_key: ctrl+space
apps:
  - name: app1
    chord: a
    command: cmd1
  - name: app2
    chord: a
    command: cmd2
`
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	_ = runDoctor(false, &stdout, &stderr)

	out := stdout.String()
	found := false
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "Config valid") && strings.Contains(line, "FAIL") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Config valid should FAIL for duplicate chords: %q", out)
	}
}

func TestDoctor_MissingAppsDir(t *testing.T) {
	dir := setupDoctorEnv(t)

	cfg := `leader_key: ctrl+space
apps:
  - name: testapp
    chord: t
    command: echo
    environment: native
`
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	// Deliberately do NOT create apps.d/.

	var stdout, stderr bytes.Buffer
	_ = runDoctor(false, &stdout, &stderr)

	out := stdout.String()
	found := false
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "apps.d/") && strings.Contains(line, "WARN") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("apps.d/ check should WARN when directory missing: %q", out)
	}
}

func TestDoctor_AppsDirWithFiles(t *testing.T) {
	dir := setupDoctorEnv(t)

	cfg := `leader_key: ctrl+space
apps: []
`
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	appsDir := filepath.Join(dir, "apps.d")
	if err := os.MkdirAll(appsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	dropin := `name: myapp
chord: m
command: echo hello
environment: native
`
	if err := os.WriteFile(filepath.Join(appsDir, "myapp.yml"), []byte(dropin), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	_ = runDoctor(false, &stdout, &stderr)

	out := stdout.String()
	if !strings.Contains(out, "apps.d/") && !strings.Contains(out, "PASS") {
		t.Errorf("apps.d/ check should PASS with valid files: %q", out)
	}
}

func TestDoctor_AppsDirBadFile(t *testing.T) {
	dir := setupDoctorEnv(t)

	cfg := `leader_key: ctrl+space
apps: []
`
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	appsDir := filepath.Join(dir, "apps.d")
	if err := os.MkdirAll(appsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Write invalid YAML as a drop-in.
	if err := os.WriteFile(filepath.Join(appsDir, "bad.yml"), []byte("{{nope"), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	_ = runDoctor(false, &stdout, &stderr)

	out := stdout.String()
	found := false
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "bad.yml") && strings.Contains(line, "FAIL") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("bad drop-in file should produce FAIL: %q", out)
	}
}

func TestDoctor_CommandNotFound(t *testing.T) {
	dir := setupDoctorEnv(t)

	cfg := `leader_key: ctrl+space
apps:
  - name: ghostapp
    chord: g
    command: "this-command-does-not-exist-xyz"
    environment: native
`
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "apps.d"), 0o755); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	_ = runDoctor(false, &stdout, &stderr)

	out := stdout.String()
	found := false
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "ghostapp") && strings.Contains(line, "command") && strings.Contains(line, "WARN") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("missing command should produce WARN: %q", out)
	}
}

func TestDoctor_CommandFound(t *testing.T) {
	dir := setupDoctorEnv(t)

	// "echo" should exist on all platforms.
	cfg := `leader_key: ctrl+space
apps:
  - name: echoapp
    chord: e
    command: "echo hello"
    environment: native
`
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "apps.d"), 0o755); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	_ = runDoctor(false, &stdout, &stderr)

	out := stdout.String()
	found := false
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "echoapp") && strings.Contains(line, "command") && strings.Contains(line, "PASS") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("echo command should produce PASS: %q", out)
	}
}

func TestDoctor_WSLSkipped(t *testing.T) {
	dir := setupDoctorEnv(t)

	// No WSL apps — WSL check should be skipped.
	cfg := `leader_key: ctrl+space
apps:
  - name: testapp
    chord: t
    command: echo
    environment: native
`
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "apps.d"), 0o755); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	_ = runDoctor(false, &stdout, &stderr)

	out := stdout.String()
	found := false
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "WSL") && strings.Contains(line, "SKIP") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("WSL check should be SKIP when no WSL apps: %q", out)
	}
}

func TestDoctor_ReadyCmdCheck(t *testing.T) {
	dir := setupDoctorEnv(t)

	cfg := `leader_key: ctrl+space
apps:
  - name: withready
    chord: w
    command: echo
    environment: native
    ready_cmd: "echo ok"
  - name: noready
    chord: n
    command: echo
    environment: native
`
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "apps.d"), 0o755); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	_ = runDoctor(false, &stdout, &stderr)

	out := stdout.String()
	// "withready" has a ready_cmd and "echo" is resolvable → PASS.
	foundPass := false
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "withready") && strings.Contains(line, "ready_cmd") && strings.Contains(line, "PASS") {
			foundPass = true
			break
		}
	}
	if !foundPass {
		t.Errorf("withready ready_cmd should PASS: %q", out)
	}

	// "noready" has no ready_cmd → should be skipped (not present in output as a ready_cmd line).
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "noready") && strings.Contains(line, "ready_cmd") {
			t.Errorf("noready should not have a ready_cmd check line: %q", line)
		}
	}
}

func TestDoctor_JSONOutput(t *testing.T) {
	dir := setupDoctorEnv(t)

	cfg := `leader_key: ctrl+space
apps:
  - name: testapp
    chord: t
    command: echo
    environment: native
`
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "apps.d"), 0o755); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	err := runDoctor(true, &stdout, &stderr)
	if err != nil {
		t.Fatalf("runDoctor(json) error = %v", err)
	}

	var report DoctorReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("JSON unmarshal error: %v\noutput: %s", err, stdout.String())
	}

	if len(report.Checks) == 0 {
		t.Error("expected at least one check in JSON output")
	}

	// Verify structure: each check has name, status, message.
	for _, c := range report.Checks {
		if c.Name == "" {
			t.Error("check name should not be empty")
		}
		if c.Status == "" {
			t.Errorf("check %q status should not be empty", c.Name)
		}
	}

	// Summary counts should be consistent.
	total := report.Passed + report.Warned + report.Failed + report.Skipped
	if total != len(report.Checks) {
		t.Errorf("summary counts (%d) != check count (%d)", total, len(report.Checks))
	}
}

func TestDoctor_JSONStructuredFields(t *testing.T) {
	dir := setupDoctorEnv(t)

	cfg := `leader_key: ctrl+space
apps:
  - name: testapp
    chord: t
    command: echo
    environment: native
`
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "apps.d"), 0o755); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	_ = runDoctor(true, &stdout, &stderr)

	// Should be valid JSON with expected top-level keys.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(stdout.Bytes(), &raw); err != nil {
		t.Fatalf("not valid JSON: %v", err)
	}
	for _, key := range []string{"checks", "passed", "warned", "failed", "skipped"} {
		if _, ok := raw[key]; !ok {
			t.Errorf("JSON output missing key %q", key)
		}
	}
}

func TestDoctor_DaemonRunning(t *testing.T) {
	dir := setupDoctorEnv(t)

	// Simulate a running daemon by writing our own PID.
	lockPath := filepath.Join(dir, "orchestratr.pid")
	if err := os.WriteFile(lockPath, []byte(strconv.Itoa(os.Getpid())), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := `leader_key: ctrl+space
apps:
  - name: testapp
    chord: t
    command: echo
    environment: native
`
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "apps.d"), 0o755); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	_ = runDoctor(false, &stdout, &stderr)

	out := stdout.String()
	found := false
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "Daemon") && strings.Contains(line, "PASS") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Daemon check should PASS when PID alive: %q", out)
	}
}

func TestDoctor_SummaryLine(t *testing.T) {
	dir := setupDoctorEnv(t)

	cfg := `leader_key: ctrl+space
apps:
  - name: testapp
    chord: t
    command: echo
    environment: native
`
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "apps.d"), 0o755); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	_ = runDoctor(false, &stdout, &stderr)

	out := stdout.String()
	if !strings.Contains(out, "passed") || !strings.Contains(out, "failed") {
		t.Errorf("output should contain summary with passed/failed counts: %q", out)
	}
}

func TestDoctor_MissingConfig(t *testing.T) {
	dir := setupDoctorEnv(t)
	// Config file does not exist.
	_ = dir

	var stdout, stderr bytes.Buffer
	_ = runDoctor(false, &stdout, &stderr)

	out := stdout.String()
	found := false
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "Config parse") && strings.Contains(line, "FAIL") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Config parse should FAIL when config missing: %q", out)
	}
}

func TestDoctor_ViaRunCommand(t *testing.T) {
	dir := setupDoctorEnv(t)

	cfg := `leader_key: ctrl+space
apps:
  - name: testapp
    chord: t
    command: echo
    environment: native
`
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "apps.d"), 0o755); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	err := run([]string{"doctor"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run(doctor) error = %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "Daemon") {
		t.Errorf("output should contain Daemon check: %q", out)
	}
}

func TestDoctor_ViaRunCommandJSON(t *testing.T) {
	dir := setupDoctorEnv(t)

	cfg := `leader_key: ctrl+space
apps:
  - name: testapp
    chord: t
    command: echo
    environment: native
`
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "apps.d"), 0o755); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	err := run([]string{"doctor", "--json"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run(doctor --json) error = %v", err)
	}

	var report DoctorReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("JSON unmarshal error: %v", err)
	}
	if len(report.Checks) == 0 {
		t.Error("expected checks in JSON output")
	}
}
