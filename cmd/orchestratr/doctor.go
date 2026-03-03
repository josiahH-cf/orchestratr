package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/josiahH-cf/orchestratr/internal/registry"
)

// CheckStatus represents the outcome of a single doctor check.
type CheckStatus string

const (
	StatusPass CheckStatus = "PASS"
	StatusFail CheckStatus = "FAIL"
	StatusWarn CheckStatus = "WARN"
	StatusSkip CheckStatus = "SKIP"
)

// CheckResult holds the outcome of a single diagnostic check.
type CheckResult struct {
	Name    string      `json:"name"`
	Status  CheckStatus `json:"status"`
	Message string      `json:"message"`
}

// DoctorReport is the complete diagnostic report.
type DoctorReport struct {
	Checks  []CheckResult `json:"checks"`
	Passed  int           `json:"passed"`
	Warned  int           `json:"warned"`
	Failed  int           `json:"failed"`
	Skipped int           `json:"skipped"`
}

// lookPath is a package-level variable so tests can override it.
var lookPath = exec.LookPath

// runDoctor executes all diagnostic checks and prints the results.
// When jsonOutput is true the report is emitted as machine-readable JSON.
func runDoctor(jsonOutput bool, stdout, stderr io.Writer) error {
	cfgPath := registry.DefaultConfigPath()
	if envPath := os.Getenv("ORCHESTRATR_CONFIG"); envPath != "" {
		cfgPath = envPath
	}

	var checks []CheckResult

	// 1. Daemon check.
	checks = append(checks, checkDaemon())

	// 2–3. Config parse + validation.
	cfg, parseResult := checkConfigParseable(cfgPath)
	checks = append(checks, parseResult)

	if cfg != nil {
		checks = append(checks, checkConfigValid(cfg))
	} else {
		checks = append(checks, CheckResult{
			Name:    "Config valid",
			Status:  StatusSkip,
			Message: "skipped (config could not be parsed)",
		})
	}

	// 4. apps.d/ directory.
	checks = append(checks, checkAppsDirExists(cfgPath))

	// 5. apps.d/ files.
	checks = append(checks, checkAppsDirFiles(cfgPath)...)

	// Collect all apps (config + drop-ins) for per-app checks.
	var allApps []registry.AppEntry
	if cfg != nil {
		allApps = append(allApps, cfg.Apps...)
	}
	// Also include drop-in apps that loaded successfully.
	appsDir := registry.AppsDirPath(cfgPath)
	dropins, _ := registry.LoadAppsDir(appsDir)
	if cfg != nil && len(dropins) > 0 {
		merged, _ := registry.MergeApps(cfg.Apps, dropins, nil)
		allApps = merged
	} else if cfg == nil {
		allApps = dropins
	}

	// 6. Per-app command checks.
	for _, app := range allApps {
		checks = append(checks, checkAppCommand(app))
	}

	// 7. WSL availability.
	checks = append(checks, checkWSLAvailable(allApps))

	// 8. Per-app ready_cmd checks.
	for _, app := range allApps {
		if app.ReadyCmd == "" {
			continue
		}
		checks = append(checks, checkReadyCmd(app))
	}

	// Build report.
	report := DoctorReport{Checks: checks}
	for _, c := range checks {
		switch c.Status {
		case StatusPass:
			report.Passed++
		case StatusWarn:
			report.Warned++
		case StatusFail:
			report.Failed++
		case StatusSkip:
			report.Skipped++
		}
	}

	if jsonOutput {
		return printDoctorJSON(report, stdout)
	}
	return printDoctorText(report, stdout)
}

// printDoctorJSON marshals the report as indented JSON.
func printDoctorJSON(report DoctorReport, w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

// printDoctorText formats the report in the human-readable dotted style.
func printDoctorText(report DoctorReport, w io.Writer) error {
	fmt.Fprintln(w, "orchestratr doctor")
	fmt.Fprintln(w)

	for _, c := range report.Checks {
		// Pad the name with dots to a fixed width.
		label := c.Name + " "
		const width = 22
		if len(label) < width {
			label += strings.Repeat(".", width-len(label))
		}
		fmt.Fprintf(w, "  %s %s", label, c.Status)
		if c.Message != "" {
			fmt.Fprintf(w, "  (%s)", c.Message)
		}
		fmt.Fprintln(w)
	}

	fmt.Fprintln(w)
	fmt.Fprintf(w, "  %d passed, %d warning, %d failed\n",
		report.Passed, report.Warned, report.Failed)

	return nil
}

// checkDaemon verifies the daemon is running by reading the PID file
// and sending signal 0 to the process.
func checkDaemon() CheckResult {
	pid, err := readPIDFile()
	if err != nil {
		return CheckResult{
			Name:    "Daemon",
			Status:  StatusFail,
			Message: "not running (no PID file)",
		}
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return CheckResult{
			Name:    "Daemon",
			Status:  StatusFail,
			Message: fmt.Sprintf("not running (PID %d not found)", pid),
		}
	}

	if err := proc.Signal(syscall.Signal(0)); err != nil {
		return CheckResult{
			Name:    "Daemon",
			Status:  StatusFail,
			Message: fmt.Sprintf("not running (stale PID %d)", pid),
		}
	}

	// Process is alive. Try to read port for richer output.
	msg := fmt.Sprintf("PID %d", pid)
	if port, portErr := readPort(); portErr == nil {
		// Optionally verify the health endpoint responds.
		client := &http.Client{Timeout: 2 * time.Second}
		if resp, httpErr := client.Get(fmt.Sprintf("http://127.0.0.1:%d/health", port)); httpErr == nil {
			resp.Body.Close()
			msg = fmt.Sprintf("PID %d, port %d", pid, port)
		} else {
			msg = fmt.Sprintf("PID %d, port %d (API not responding)", pid, port)
		}
	}

	return CheckResult{
		Name:    "Daemon",
		Status:  StatusPass,
		Message: msg,
	}
}

// checkConfigParseable attempts to load and parse the config file.
// Returns the parsed config (nil on failure) and the check result.
func checkConfigParseable(cfgPath string) (*registry.Config, CheckResult) {
	cfg, err := registry.Load(cfgPath)
	if err != nil {
		return nil, CheckResult{
			Name:    "Config parse",
			Status:  StatusFail,
			Message: err.Error(),
		}
	}
	return cfg, CheckResult{
		Name:    "Config parse",
		Status:  StatusPass,
		Message: fmt.Sprintf("%d app(s) in config", len(cfg.Apps)),
	}
}

// checkConfigValid runs validation on a parsed config.
func checkConfigValid(cfg *registry.Config) CheckResult {
	errs := registry.ValidateConfig(cfg)
	if len(errs) == 0 {
		return CheckResult{
			Name:    "Config valid",
			Status:  StatusPass,
			Message: "all entries valid",
		}
	}

	msgs := make([]string, len(errs))
	for i, e := range errs {
		msgs[i] = e.Error()
	}
	return CheckResult{
		Name:    "Config valid",
		Status:  StatusFail,
		Message: strings.Join(msgs, "; "),
	}
}

// checkAppsDirExists checks whether the apps.d/ directory exists.
func checkAppsDirExists(cfgPath string) CheckResult {
	dir := registry.AppsDirPath(cfgPath)
	info, err := os.Stat(dir)
	if err != nil {
		return CheckResult{
			Name:    "apps.d/ dir",
			Status:  StatusWarn,
			Message: "directory not found",
		}
	}
	if !info.IsDir() {
		return CheckResult{
			Name:    "apps.d/ dir",
			Status:  StatusFail,
			Message: "path exists but is not a directory",
		}
	}
	return CheckResult{
		Name:    "apps.d/ dir",
		Status:  StatusPass,
		Message: "exists",
	}
}

// checkAppsDirFiles scans the apps.d/ directory and reports per-file results.
func checkAppsDirFiles(cfgPath string) []CheckResult {
	dir := registry.AppsDirPath(cfgPath)
	if _, err := os.Stat(dir); err != nil {
		return nil // directory check already reported the issue
	}

	apps, errs := registry.LoadAppsDir(dir)
	var results []CheckResult

	for _, app := range apps {
		source := app.Source
		if strings.HasPrefix(source, "apps.d/") {
			source = filepath.Base(source)
		}
		results = append(results, CheckResult{
			Name:    fmt.Sprintf("apps.d/%s", source),
			Status:  StatusPass,
			Message: fmt.Sprintf("loaded %s", app.Name),
		})
	}

	for _, e := range errs {
		results = append(results, CheckResult{
			Name:    "apps.d/ file",
			Status:  StatusFail,
			Message: e.Error(),
		})
	}

	if len(apps) == 0 && len(errs) == 0 {
		results = append(results, CheckResult{
			Name:    "apps.d/ files",
			Status:  StatusPass,
			Message: "no drop-in files",
		})
	}

	return results
}

// checkAppCommand checks whether the first token of an app's command
// is resolvable via PATH.
func checkAppCommand(app registry.AppEntry) CheckResult {
	name := fmt.Sprintf("%s command", app.Name)

	cmdBin := firstToken(app.Command)
	if cmdBin == "" {
		return CheckResult{
			Name:    name,
			Status:  StatusWarn,
			Message: "empty command",
		}
	}

	// For WSL apps, the host runs wsl.exe; the inner command isn't
	// resolvable on the host PATH. Check wsl.exe instead.
	if strings.HasPrefix(app.Environment, "wsl") {
		wslBin := "wsl.exe"
		if _, err := lookPath(wslBin); err != nil {
			return CheckResult{
				Name:    name,
				Status:  StatusWarn,
				Message: fmt.Sprintf("wsl.exe not found (needed to run %q)", app.Command),
			}
		}
		return CheckResult{
			Name:    name,
			Status:  StatusPass,
			Message: fmt.Sprintf("wsl.exe → %s", app.Command),
		}
	}

	if _, err := lookPath(cmdBin); err != nil {
		return CheckResult{
			Name:    name,
			Status:  StatusWarn,
			Message: fmt.Sprintf("%s not found in PATH", cmdBin),
		}
	}
	return CheckResult{
		Name:    name,
		Status:  StatusPass,
		Message: cmdBin,
	}
}

// checkWSLAvailable checks whether wsl.exe is available when at least
// one app uses a WSL environment.
func checkWSLAvailable(apps []registry.AppEntry) CheckResult {
	hasWSL := false
	for _, app := range apps {
		if strings.HasPrefix(app.Environment, "wsl") {
			hasWSL = true
			break
		}
	}

	if !hasWSL {
		return CheckResult{
			Name:    "WSL available",
			Status:  StatusSkip,
			Message: "no WSL apps configured",
		}
	}

	if _, err := lookPath("wsl.exe"); err != nil {
		return CheckResult{
			Name:    "WSL available",
			Status:  StatusFail,
			Message: "wsl.exe not found in PATH",
		}
	}

	return CheckResult{
		Name:    "WSL available",
		Status:  StatusPass,
		Message: "wsl.exe found",
	}
}

// checkReadyCmd validates that a ready_cmd's first token is resolvable.
func checkReadyCmd(app registry.AppEntry) CheckResult {
	name := fmt.Sprintf("%s ready_cmd", app.Name)

	cmdBin := firstToken(app.ReadyCmd)
	if cmdBin == "" {
		return CheckResult{
			Name:    name,
			Status:  StatusWarn,
			Message: "empty ready_cmd",
		}
	}

	// For WSL apps, the ready_cmd also runs via wsl.exe.
	if strings.HasPrefix(app.Environment, "wsl") {
		return CheckResult{
			Name:    name,
			Status:  StatusPass,
			Message: fmt.Sprintf("wsl.exe → %s", app.ReadyCmd),
		}
	}

	if _, err := lookPath(cmdBin); err != nil {
		return CheckResult{
			Name:    name,
			Status:  StatusWarn,
			Message: fmt.Sprintf("%s not found in PATH", cmdBin),
		}
	}
	return CheckResult{
		Name:    name,
		Status:  StatusPass,
		Message: cmdBin,
	}
}

// firstToken splits a command string and returns the first word
// (the binary name).
func firstToken(cmd string) string {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}
