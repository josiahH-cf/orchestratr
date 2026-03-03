package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/josiahH-cf/orchestratr/internal/api"
	"github.com/josiahH-cf/orchestratr/internal/daemon"
	"github.com/josiahH-cf/orchestratr/internal/gui"
	"github.com/josiahH-cf/orchestratr/internal/hotkey"
	"github.com/josiahH-cf/orchestratr/internal/launcher"
	"github.com/josiahH-cf/orchestratr/internal/registry"
	"github.com/josiahH-cf/orchestratr/internal/tray"
)

// Version is set at build time via -ldflags "-X main.Version=..."
var Version = "v0.0.0-dev"

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		fmt.Fprintln(stdout, "orchestratr — system-wide app launcher")
		fmt.Fprintln(stdout, "Usage: orchestratr [start|stop|status|reload|list|launch|trigger|configure|doctor|install|uninstall|version]")
		return nil
	}

	switch args[0] {
	case "version":
		fmt.Fprintf(stdout, "orchestratr %s\n", Version)
		return nil

	case "list":
		return runList(stdout, stderr)

	case "start":
		if !hasFlag(args[1:], "--foreground") {
			return daemonize(stdout, stderr)
		}
		return runStart(stdout, stderr)

	case "stop":
		return runStop(stdout, stderr)

	case "status":
		return runStatus(stdout, stderr)

	case "trigger":
		var chord string
		if len(args) > 1 {
			chord = args[1]
		}
		return runTrigger(chord, stdout, stderr)

	case "launch":
		if len(args) < 2 {
			return fmt.Errorf("usage: orchestratr launch <app-name>")
		}
		return runLaunch(args[1], stdout, stderr)

	case "reload":
		return runReload(stdout, stderr)

	case "configure":
		return runConfigure(stdout, stderr)

	case "install":
		return runInstall(stdout, stderr)

	case "uninstall":
		return runUninstall(stdout, stderr)

	case "doctor":
		jsonFlag := hasFlag(args[1:], "--json")
		return runDoctor(jsonFlag, stdout, stderr)

	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

// hasFlag checks whether a flag is present in the argument slice.
func hasFlag(args []string, flag string) bool {
	for _, a := range args {
		if a == flag {
			return true
		}
	}
	return false
}

// daemonize re-executes the binary with "start --foreground" in the
// background, prints the child PID, and returns immediately.
func daemonize(stdout, stderr io.Writer) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolving executable path: %w", err)
	}

	cmd := execCommand(exe, "start", "--foreground")
	cmd.Env = os.Environ()

	// Detach from the parent's stdio. The child logs to its own
	// log file; we don't need its stdout/stderr.
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	if startErr := cmd.Start(); startErr != nil {
		return fmt.Errorf("starting daemon: %w", startErr)
	}

	// Release the child so it survives the parent exiting.
	if releaseErr := cmd.Process.Release(); releaseErr != nil {
		return fmt.Errorf("releasing daemon process: %w", releaseErr)
	}

	fmt.Fprintf(stdout, "orchestratr daemon started (PID %d)\n", cmd.Process.Pid)
	return nil
}

// execCommand is a variable so tests can override it.
var execCommand = defaultExecCommand

func defaultExecCommand(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}

// runStart launches the daemon in the foreground. It is always called
// with --foreground present; daemonize() handles the background case.
// When the process is the daemonized child, cmd.Stderr is nil so
// writing to stderr is harmless (goes to /dev/null) (OI-8).
func runStart(stdout, stderr io.Writer) error {
	lockPath := lockPathFromEnv()
	lock, err := daemon.AcquireLock(lockPath)
	if err != nil {
		return fmt.Errorf("cannot start: %w", err)
	}
	defer lock.Release()

	// Load config for API port.
	cfgPath := registry.DefaultConfigPath()
	if envPath := os.Getenv("ORCHESTRATR_CONFIG"); envPath != "" {
		cfgPath = envPath
	}

	// Ensure default config and apps.d/ directory exist on first run.
	if _, ensureErr := registry.EnsureDefaults(cfgPath); ensureErr != nil {
		fmt.Fprintf(stderr, "warning: could not create default config: %v\n", ensureErr)
	}
	if ensureErr := registry.EnsureAppsDir(cfgPath); ensureErr != nil {
		fmt.Fprintf(stderr, "warning: could not create apps.d directory: %v\n", ensureErr)
	}

	apiPort := 9876 // default
	cfg, err := registry.LoadWithDropins(cfgPath, nil)
	if err == nil && cfg.APIPort > 0 {
		apiPort = cfg.APIPort
	}

	// Build registry from config.
	var reg *registry.Registry
	if cfg != nil {
		reg = registry.NewRegistry(*cfg)
	}

	// Set up log file.
	logPath := daemon.DefaultLogPath()
	logFile, err := daemon.SetupLogFile(logPath)
	if err != nil {
		fmt.Fprintf(stderr, "warning: could not open log file: %v (logging to stderr)\n", err)
	} else {
		defer logFile.Close()
	}

	// Log to both stderr and the log file. When running as the
	// daemonized child, stderr is /dev/null so no output leaks to the
	// parent's terminal (OI-8).
	logger := log.New(stderr, "orchestratr: ", log.LstdFlags)
	if logFile != nil {
		logger = log.New(io.MultiWriter(stderr, logFile), "orchestratr: ", log.LstdFlags)
	}

	// Declare hotkeyEngine and apiSrv early so the reload closure
	// can capture them. They are assigned real values later.
	var hotkeyEngine *hotkey.Engine
	var apiSrv *api.Server

	// Build reload function for POST /reload and file watcher.
	reloadFn := func() (*registry.Config, error) {
		newCfg, loadErr := registry.LoadWithDropins(cfgPath, logger)
		if loadErr != nil {
			return nil, loadErr
		}
		if reg != nil {
			reg.Swap(*newCfg)
		}

		// Update hotkey chord map.
		if hotkeyEngine != nil {
			newChords := buildChords(newCfg.Apps)
			if swapErr := hotkeyEngine.SwapChords(newChords); swapErr != nil {
				logger.Printf("warning: hotkey chord swap failed: %v", swapErr)
			}
		}

		// Sync state tracker: remove entries for apps no longer in config.
		if apiSrv != nil {
			names := make([]string, len(newCfg.Apps))
			for i, a := range newCfg.Apps {
				names[i] = a.Name
			}
			apiSrv.State().Sync(names)
		}

		return newCfg, nil
	}

	// Warn on port 0.
	if cfg != nil && cfg.APIPort == 0 {
		logger.Println("warning: api_port is 0; a random port will be assigned each start")
	}

	// Start API server.
	apiSrv = api.NewServer(apiPort, Version, reg, reloadFn)
	go func() {
		if srvErr := apiSrv.Start(); srvErr != nil && srvErr != http.ErrServerClosed {
			logger.Printf("API server error: %v", srvErr)
		}
	}()
	defer apiSrv.Stop()

	// Wait for the API server to be listening before writing the
	// port file; otherwise clients may read a port that isn't ready.
	if !apiSrv.WaitReady(5) {
		logger.Println("warning: API server did not become ready within 5s")
	}

	// Write port discovery file.
	portFilePath := daemon.DefaultPortFilePath()
	actualPort := apiSrv.Port()
	if writeErr := daemon.WritePortFile(portFilePath, actualPort); writeErr != nil {
		logger.Printf("warning: could not write port file: %v", writeErr)
	} else {
		defer daemon.RemovePortFile(portFilePath)
	}

	d := daemon.New(daemon.Config{
		LogLevel: "info",
		APIPort:  actualPort,
	})
	d.SetLogger(logger)

	// Set up system tray.
	var trayProvider tray.Provider = &tray.HeadlessProvider{}
	if setupErr := trayProvider.Setup(); setupErr != nil {
		logger.Printf("warning: tray setup failed: %v", setupErr)
	}

	// Build the app launcher.
	exec := launcher.NewPlatformExecutor(
		launcher.WithLogger(logger),
		launcher.WithExitCallback(func(name string, exitErr error) {
			apiSrv.State().SetStopped(name)
			if exitErr != nil {
				logger.Printf("app %q stopped with error: %v", name, exitErr)
			} else {
				logger.Printf("app %q stopped", name)
			}
		}),
	)
	defer exec.StopAll()

	// Create daemon-scoped context for cancellation of background
	// goroutines (readiness polling, etc.).
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Build and start the hotkey engine.
	if cfg != nil {
		chords := buildChords(cfg.Apps)
		listener := hotkey.NewPlatformListener()
		engine, engineErr := hotkey.NewEngine(hotkey.EngineConfig{
			LeaderKey:      cfg.LeaderKey,
			ChordTimeoutMs: cfg.ChordTimeoutMs,
			Chords:         chords,
			OnAction: func(action string) {
				logger.Printf("chord dispatched: %s", action)
				launchApp(ctx, exec, reg, apiSrv, logger, action, trayProvider)
			},
			Logger: logger,
		}, listener)
		if engineErr != nil {
			logger.Printf("warning: hotkey engine not available: %v", engineErr)
		} else {
			if startErr := engine.Start(); startErr != nil {
				logger.Printf("warning: hotkey engine failed to start: %v", startErr)
			} else {
				hotkeyEngine = engine
				defer hotkeyEngine.Stop()

				// Wire the trigger API endpoint to the engine.
				apiSrv.SetTriggerFunc(func() error {
					return hotkeyEngine.Trigger()
				})
			}
		}
	}

	// Wire the launch API endpoint to the launcher (works even without
	// the hotkey engine, enabling WSL/Wayland/CLI app launching).
	apiSrv.SetLaunchFunc(func(name string) (int, error) {
		if reg == nil {
			return 0, fmt.Errorf("registry not loaded")
		}
		app, found := reg.FindByName(name)
		if !found {
			return 0, fmt.Errorf("app %q not found", name)
		}
		if exec.IsRunning(name) {
			pid, _ := exec.PID(name)
			if focusErr := launcher.FocusWindow(pid); focusErr != nil {
				logger.Printf("focus %q (PID %d) failed: %v", name, pid, focusErr)
			} else {
				logger.Printf("focused %q (PID %d)", name, pid)
			}
			return pid, nil
		}
		result, launchErr := exec.Launch(app)
		if launchErr != nil {
			if trayProvider != nil {
				trayProvider.NotifyError("Launch Failed", fmt.Sprintf("%s: %v", name, launchErr))
			}
			return 0, launchErr
		}
		logger.Printf("launched %q (PID %d)", name, result.PID)
		apiSrv.State().SetLaunched(name)
		go launcher.PollReadiness(ctx, app, apiSrv.State(), exec, logger)
		return result.PID, nil
	})

	// Wire tray callbacks now that daemon and engine are initialized.

	// Track the GUI server instance so it can be stopped on shutdown
	// or when a new configure request arrives (OI-10).
	var guiSrv *gui.Server
	var guiMu sync.Mutex

	trayProvider.OnPause(func() {
		if err := d.Pause(); err != nil {
			logger.Printf("pause failed: %v", err)
			return
		}
		if hotkeyEngine != nil {
			hotkeyEngine.Pause()
		}
		_ = trayProvider.SetState("paused")
	})
	trayProvider.OnResume(func() {
		if err := d.Resume(); err != nil {
			logger.Printf("resume failed: %v", err)
			return
		}
		if hotkeyEngine != nil {
			hotkeyEngine.Resume()
		}
		_ = trayProvider.SetState("running")
	})
	trayProvider.OnQuit(func() {
		logger.Println("quit requested via tray")
		cancel()
	})
	trayProvider.OnConfigure(func() {
		logger.Println("configure requested via tray")

		guiMu.Lock()
		// Stop the previous GUI server if still running.
		if guiSrv != nil {
			guiSrv.Stop()
		}
		guiSrv = gui.NewServer(cfgPath, actualPort, logger)
		srv := guiSrv
		guiMu.Unlock()

		if guiErr := srv.Start(); guiErr != nil {
			logger.Printf("gui server start failed: %v", guiErr)
			return
		}
		logger.Printf("management GUI listening on %s", srv.URL())
		if openErr := srv.OpenBrowser(); openErr != nil {
			logger.Printf("could not open browser: %v — visit %s", openErr, srv.URL())
		}
	})
	// Stop any running GUI server on daemon shutdown.
	defer func() {
		guiMu.Lock()
		if guiSrv != nil {
			guiSrv.Stop()
		}
		guiMu.Unlock()
	}()
	_ = trayProvider.SetState("running")

	// Start file watcher for config hot-reload (watches both config.yml and apps.d/).
	watcherReload := func(path string) error {
		_, err := reloadFn()
		return err
	}
	appsDir := registry.AppsDirPath(cfgPath)
	w := registry.NewWatcher(cfgPath, watcherReload,
		registry.WithLogger(logger),
		registry.WithAppsDir(appsDir),
	)
	if watchErr := w.Start(context.Background()); watchErr != nil {
		logger.Printf("warning: file watcher not available: %v", watchErr)
	} else {
		defer w.Stop()
	}

	fmt.Fprintf(stdout, "orchestratr daemon starting on port %d\n", actualPort)
	logger.Printf("PID %d, API port %d", os.Getpid(), actualPort)

	err = d.Run(ctx)
	trayProvider.Quit()
	return err
}

// runStop sends SIGTERM to the running daemon.
func runStop(stdout, stderr io.Writer) error {
	pid, err := readPIDFile()
	if err != nil {
		return fmt.Errorf("daemon is not running: %w", err)
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("finding process %d: %w", pid, err)
	}

	if err := proc.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("sending SIGTERM to PID %d: %w", pid, err)
	}

	fmt.Fprintf(stdout, "sent stop signal to orchestratr (PID %d)\n", pid)
	return nil
}

// runStatus reports whether the daemon is running.
func runStatus(stdout, stderr io.Writer) error {
	pid, err := readPIDFile()
	if err != nil {
		fmt.Fprintln(stdout, "orchestratr is not running")
		return nil
	}

	// Check if the process is alive.
	proc, err := os.FindProcess(pid)
	if err != nil {
		fmt.Fprintln(stdout, "orchestratr is not running")
		return nil
	}

	if err := proc.Signal(syscall.Signal(0)); err != nil {
		fmt.Fprintf(stdout, "orchestratr is not running (stale PID %d)\n", pid)
		return nil
	}

	// Try to read the port file for richer output (OI-7).
	if port, portErr := readPort(); portErr == nil {
		fmt.Fprintf(stdout, "orchestratr is running (PID %d, port %d)\n", pid, port)
	} else {
		fmt.Fprintf(stdout, "orchestratr is running (PID %d)\n", pid)
	}
	return nil
}

// readPIDFile reads the daemon PID from the lock file.
func readPIDFile() (int, error) {
	data, err := os.ReadFile(lockPathFromEnv())
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(data)))
}

// lockPathFromEnv returns the lock file path, allowing override via
// ORCHESTRATR_LOCK_PATH for testing.
func lockPathFromEnv() string {
	if p := os.Getenv("ORCHESTRATR_LOCK_PATH"); p != "" {
		return p
	}
	return daemon.DefaultLockPath()
}

// runList loads the config and prints the app registry as a table.
// When the daemon is running, it enriches the table with live
// running/stopped status from the API.
func runList(stdout, stderr io.Writer) error {
	path := registry.DefaultConfigPath()

	// Allow override via environment variable for testing.
	if envPath := os.Getenv("ORCHESTRATR_CONFIG"); envPath != "" {
		path = envPath
	}

	cfg, err := registry.LoadWithDropins(path, nil)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("config not found at %s\nRun orchestratr once to generate a default config, or set ORCHESTRATR_CONFIG", path)
		}
		return fmt.Errorf("loading config: %w", err)
	}

	if len(cfg.Apps) == 0 {
		fmt.Fprintln(stdout, "No apps registered. Edit your config file to add apps.")
		fmt.Fprintf(stdout, "Config: %s\n", path)
		return nil
	}

	// Try to fetch live state from the running daemon.
	states := fetchAppStates()

	w := tabwriter.NewWriter(stdout, 0, 4, 2, ' ', 0)
	if states != nil {
		fmt.Fprintln(w, "NAME\tCHORD\tSOURCE\tSTATUS\tCOMMAND\tENV\tDESCRIPTION")
		for _, app := range cfg.Apps {
			status := "stopped"
			if s, ok := states[app.Name]; ok && s.Launched {
				status = "launched"
				if s.Ready {
					status = "ready"
				}
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				app.Name, app.Chord, app.Source, status, app.Command, app.Environment, app.Description)
		}
	} else {
		fmt.Fprintln(w, "NAME\tCHORD\tSOURCE\tCOMMAND\tENV\tDESCRIPTION")
		for _, app := range cfg.Apps {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				app.Name, app.Chord, app.Source, app.Command, app.Environment, app.Description)
		}
	}
	return w.Flush()
}

// fetchAppStates attempts to contact the running daemon and retrieve
// live state for all apps. Returns nil if the daemon is not reachable.
func fetchAppStates() map[string]*api.AppState {
	port, err := readPort()
	if err != nil {
		return nil
	}

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/health", port))
	if err != nil {
		return nil
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil
	}

	// Fetch per-app state. The /apps endpoint returns registry entries;
	// we need /apps/{name}/state for each. First get the app list.
	appsResp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/apps", port))
	if err != nil {
		return nil
	}
	defer appsResp.Body.Close()

	var apps []struct {
		Name string `json:"name"`
	}
	if decErr := json.NewDecoder(appsResp.Body).Decode(&apps); decErr != nil {
		return nil
	}

	states := make(map[string]*api.AppState, len(apps))
	for _, a := range apps {
		stateResp, stateErr := client.Get(fmt.Sprintf("http://127.0.0.1:%d/apps/%s/state", port, a.Name))
		if stateErr != nil {
			continue
		}
		var s api.AppState
		decErr := json.NewDecoder(stateResp.Body).Decode(&s)
		stateResp.Body.Close()
		if decErr != nil {
			continue
		}
		states[a.Name] = &s
	}
	return states
}

// buildChords converts registry app entries into hotkey.Chord values.
// Entries with empty chords are skipped. Invalid chord strings are
// logged as warnings and skipped.
func buildChords(apps []registry.AppEntry) []hotkey.Chord {
	var chords []hotkey.Chord
	for _, app := range apps {
		if app.Chord == "" {
			continue
		}
		k, err := hotkey.ParseKey(app.Chord)
		if err != nil {
			// Log but don't fail — one bad chord shouldn't prevent startup.
			log.Printf("warning: skipping app %q: invalid chord %q: %v", app.Name, app.Chord, err)
			continue
		}
		chords = append(chords, hotkey.Chord{Key: k, Action: app.Name})
	}
	return chords
}

// launchApp looks up an app by name in the registry, checks if it is
// already running (attempting bring-to-front if so), and launches it
// via the executor. It updates the API state tracker on success and
// sends a tray notification on failure.
func launchApp(ctx context.Context, exec launcher.Executor, reg *registry.Registry, apiSrv *api.Server, logger *log.Logger, name string, trayProv tray.Provider) {
	if reg == nil {
		logger.Printf("launch %q: registry not loaded", name)
		return
	}

	app, found := reg.FindByName(name)
	if !found {
		logger.Printf("launch %q: app not found in registry", name)
		return
	}

	if exec.IsRunning(name) {
		pid, _ := exec.PID(name)
		if err := launcher.FocusWindow(pid); err != nil {
			logger.Printf("focus %q (PID %d) failed: %v", name, pid, err)
		} else {
			logger.Printf("focused %q (PID %d)", name, pid)
		}
		return
	}

	result, err := exec.Launch(app)
	if err != nil {
		logger.Printf("launch %q failed: %v", name, err)
		// Record detailed error on app state for diagnostic visibility.
		errMsg := fmt.Sprintf("command=%q env=%s: %v", app.Command, app.Environment, err)
		apiSrv.State().SetError(name, errMsg)
		if trayProv != nil {
			trayProv.NotifyError("Launch Failed", fmt.Sprintf("%s: %v", name, err))
		}
		return
	}

	logger.Printf("launched %q (PID %d)", name, result.PID)
	apiSrv.State().SetLaunched(name)
	go launcher.PollReadiness(ctx, app, apiSrv.State(), exec, logger)
}

// runTrigger sends a trigger request to the running daemon's API.
// When chord is non-empty, the daemon looks up the chord and launches
// the matching app directly without the hotkey engine (OI-9, OI-12).
// When chord is empty, it simulates a leader key press (original
// behavior — Wayland manual fallback).
func runTrigger(chord string, stdout, stderr io.Writer) error {
	port, err := readPort()
	if err != nil {
		return fmt.Errorf("daemon is not running: %w", err)
	}

	url := fmt.Sprintf("http://127.0.0.1:%d/trigger", port)

	var body io.Reader
	if chord != "" {
		body = strings.NewReader(fmt.Sprintf(`{"chord":%q}`, chord))
	}

	resp, err := http.Post(url, "application/json", body)
	if err != nil {
		return fmt.Errorf("sending trigger: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		if chord != "" {
			var result struct {
				App string `json:"app"`
				PID int    `json:"pid"`
			}
			if decErr := json.NewDecoder(resp.Body).Decode(&result); decErr == nil && result.App != "" {
				fmt.Fprintf(stdout, "launched %s (PID %d)\n", result.App, result.PID)
			} else {
				fmt.Fprintln(stdout, "chord triggered")
			}
		} else {
			fmt.Fprintln(stdout, "leader key triggered")
		}
		return nil
	}

	var errResp struct {
		Error string `json:"error"`
	}
	if decErr := json.NewDecoder(resp.Body).Decode(&errResp); decErr == nil && errResp.Error != "" {
		return fmt.Errorf("trigger failed: %s", errResp.Error)
	}
	return fmt.Errorf("trigger failed: HTTP %d", resp.StatusCode)
}

// runLaunch sends a launch request for a specific app to the running
// daemon's API. This enables launching apps by name without the hotkey
// engine (useful on WSL, Wayland, or for scripting).
func runLaunch(name string, stdout, stderr io.Writer) error {
	port, err := readPort()
	if err != nil {
		return fmt.Errorf("daemon is not running: %w", err)
	}

	url := fmt.Sprintf("http://127.0.0.1:%d/apps/%s/launch", port, name)
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return fmt.Errorf("launching %q: %w", name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var result struct {
			PID int `json:"pid"`
		}
		if decErr := json.NewDecoder(resp.Body).Decode(&result); decErr == nil && result.PID > 0 {
			fmt.Fprintf(stdout, "launched %s (PID %d)\n", name, result.PID)
		} else {
			fmt.Fprintf(stdout, "launched %s\n", name)
		}
		return nil
	}

	var errResp struct {
		Error string `json:"error"`
	}
	if decErr := json.NewDecoder(resp.Body).Decode(&errResp); decErr == nil && errResp.Error != "" {
		return fmt.Errorf("launch %q failed: %s", name, errResp.Error)
	}
	return fmt.Errorf("launch %q failed: HTTP %d", name, resp.StatusCode)
}

// runReload sends a reload request to the running daemon's API.
func runReload(stdout, stderr io.Writer) error {
	port, err := readPort()
	if err != nil {
		return fmt.Errorf("daemon is not running: %w", err)
	}

	url := fmt.Sprintf("http://127.0.0.1:%d/reload", port)
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return fmt.Errorf("sending reload: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var result struct {
			Status string `json:"status"`
			Apps   []any  `json:"apps"`
		}
		if decErr := json.NewDecoder(resp.Body).Decode(&result); decErr == nil {
			fmt.Fprintf(stdout, "config reloaded (%d apps)\n", len(result.Apps))
		} else {
			fmt.Fprintln(stdout, "config reloaded")
		}
		return nil
	}

	var errResp struct {
		Error string `json:"error"`
	}
	if decErr := json.NewDecoder(resp.Body).Decode(&errResp); decErr == nil && errResp.Error != "" {
		return fmt.Errorf("reload failed: %s", errResp.Error)
	}
	return fmt.Errorf("reload failed: HTTP %d", resp.StatusCode)
}

// runConfigure opens the web-based management GUI. It reads the
// config path and daemon port, starts a temporary HTTP server, and
// opens the user's browser.
func runConfigure(stdout, stderr io.Writer) error {
	cfgPath := registry.DefaultConfigPath()
	if envPath := os.Getenv("ORCHESTRATR_CONFIG"); envPath != "" {
		cfgPath = envPath
	}

	// Try to detect the daemon port for live reload.
	daemonPort := 0
	if port, err := readPort(); err == nil {
		daemonPort = port
	}

	logger := log.New(stderr, "gui: ", log.LstdFlags)
	srv := gui.NewServer(cfgPath, daemonPort, logger)
	if err := srv.Start(); err != nil {
		return fmt.Errorf("starting gui: %w", err)
	}
	defer srv.Stop()

	fmt.Fprintf(stdout, "management GUI: %s\n", srv.URL())

	if err := srv.OpenBrowser(); err != nil {
		fmt.Fprintf(stderr, "could not open browser: %v\n", err)
		fmt.Fprintf(stdout, "open %s in your browser\n", srv.URL())
	}

	// Block until interrupted.
	fmt.Fprintln(stdout, "press Ctrl+C to close the GUI server")
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	fmt.Fprintln(stdout, "\nshutting down GUI server")
	return nil
}

// readPort reads the daemon port from the port discovery file.
func readPort() (int, error) {
	portPath := daemon.DefaultPortFilePath()
	if p := os.Getenv("ORCHESTRATR_PORT_PATH"); p != "" {
		portPath = p
	}
	data, err := os.ReadFile(portPath)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(data)))
}
