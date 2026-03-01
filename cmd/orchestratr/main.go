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
	"strconv"
	"strings"
	"syscall"
	"text/tabwriter"

	"github.com/josiahH-cf/orchestratr/internal/api"
	"github.com/josiahH-cf/orchestratr/internal/daemon"
	"github.com/josiahH-cf/orchestratr/internal/hotkey"
	"github.com/josiahH-cf/orchestratr/internal/registry"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		fmt.Fprintln(stdout, "orchestratr — system-wide app launcher")
		fmt.Fprintln(stdout, "Usage: orchestratr [start|stop|status|reload|list|trigger|version]")
		return nil
	}

	switch args[0] {
	case "version":
		fmt.Fprintln(stdout, "orchestratr v0.0.0-dev")
		return nil

	case "list":
		return runList(stdout, stderr)

	case "start":
		return runStart(stdout, stderr)

	case "stop":
		return runStop(stdout, stderr)

	case "status":
		return runStatus(stdout, stderr)

	case "trigger":
		return runTrigger(stdout, stderr)

	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

// runStart launches the daemon in the foreground.
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

	apiPort := 9876 // default
	cfg, err := registry.LoadAndValidate(cfgPath)
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

	logger := log.New(stderr, "orchestratr: ", log.LstdFlags)
	if logFile != nil {
		logger = log.New(io.MultiWriter(stderr, logFile), "orchestratr: ", log.LstdFlags)
	}

	// Build reload function for POST /reload.
	reloadFn := func() (*registry.Config, error) {
		newCfg, loadErr := registry.LoadAndValidate(cfgPath)
		if loadErr != nil {
			return nil, loadErr
		}
		if reg != nil {
			reg.Swap(*newCfg)
		}
		return newCfg, nil
	}

	// Start API server.
	apiSrv := api.NewServer(apiPort, "v0.0.0-dev", reg, reloadFn)
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

	// Build and start the hotkey engine.
	var hotkeyEngine *hotkey.Engine
	if cfg != nil {
		chords := buildChords(cfg.Apps)
		listener := hotkey.NewPlatformListener()
		engine, engineErr := hotkey.NewEngine(hotkey.EngineConfig{
			LeaderKey:      cfg.LeaderKey,
			ChordTimeoutMs: cfg.ChordTimeoutMs,
			Chords:         chords,
			OnAction: func(action string) {
				logger.Printf("chord dispatched: %s", action)
				// TODO: integrate with launcher to actually launch apps
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

	fmt.Fprintf(stdout, "orchestratr daemon starting on port %d\n", actualPort)
	logger.Printf("PID %d, API port %d", os.Getpid(), actualPort)

	return d.Run(context.Background())
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

	fmt.Fprintf(stdout, "orchestratr is running (PID %d)\n", pid)
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
func runList(stdout, stderr io.Writer) error {
	path := registry.DefaultConfigPath()

	// Allow override via environment variable for testing.
	if envPath := os.Getenv("ORCHESTRATR_CONFIG"); envPath != "" {
		path = envPath
	}

	cfg, err := registry.LoadAndValidate(path)
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

	w := tabwriter.NewWriter(stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tCHORD\tCOMMAND\tENV\tDESCRIPTION")
	for _, app := range cfg.Apps {
		env := app.Environment
		if env == "" {
			env = "native"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			app.Name, app.Chord, app.Command, env, app.Description)
	}
	return w.Flush()
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

// runTrigger sends a trigger request to the running daemon's API,
// simulating a leader key press. This is the Wayland manual fallback:
// users configure their compositor to run "orchestratr trigger" as a
// custom keybinding.
func runTrigger(stdout, stderr io.Writer) error {
	port, err := readPort()
	if err != nil {
		return fmt.Errorf("daemon is not running: %w", err)
	}

	url := fmt.Sprintf("http://127.0.0.1:%d/trigger", port)
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return fmt.Errorf("sending trigger: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Fprintln(stdout, "leader key triggered")
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
