package main

import (
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/josiahH-cf/orchestratr/internal/autostart"
	"github.com/josiahH-cf/orchestratr/internal/hotkey"
	"github.com/josiahH-cf/orchestratr/internal/platform"
	"github.com/josiahH-cf/orchestratr/internal/registry"
)

// autostartManager returns the platform-appropriate autostart Manager,
// respecting the ORCHESTRATR_AUTOSTART_DIR environment variable for
// testing.
func autostartManager() autostart.Manager {
	dir := os.Getenv("ORCHESTRATR_AUTOSTART_DIR")
	if dir != "" {
		switch runtime.GOOS {
		case "linux":
			return &autostart.LinuxManager{ConfigDir: dir}
		case "darwin":
			return &autostart.DarwinManager{LaunchAgentsDir: dir}
		case "windows":
			return &autostart.WindowsManager{RegistryDir: dir}
		default:
			return autostart.NewManager()
		}
	}
	return autostart.NewManager()
}

// isWSL2 checks for WSL2, respecting ORCHESTRATR_PROC_VERSION override for testing.
func isWSL2() bool {
	if path := os.Getenv("ORCHESTRATR_PROC_VERSION"); path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return false
		}
		return containsWSL2Markers(string(data))
	}
	return platform.IsWSL2()
}

// containsWSL2Markers checks if a string contains WSL2/Microsoft markers.
func containsWSL2Markers(s string) bool {
	lower := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		lower[i] = c
	}
	ls := string(lower)
	for _, marker := range []string{"microsoft", "wsl2"} {
		for i := 0; i <= len(ls)-len(marker); i++ {
			if ls[i:i+len(marker)] == marker {
				return true
			}
		}
	}
	return false
}

// runInstall configures autostart, verifies hotkey registration,
// detects WSL2, and checks macOS accessibility.
func runInstall(stdout, stderr io.Writer) error {
	// Detect WSL2 and warn.
	if isWSL2() {
		fmt.Fprintln(stderr, platform.WSL2Warning())
		fmt.Fprintln(stderr)
	}

	// Ensure config exists (create defaults if missing).
	cfgPath := registry.DefaultConfigPath()
	if envPath := os.Getenv("ORCHESTRATR_CONFIG"); envPath != "" {
		cfgPath = envPath
	}

	if _, err := registry.EnsureDefaults(cfgPath); err != nil {
		fmt.Fprintf(stderr, "warning: could not create default config: %v\n", err)
	}

	// Load and validate config.
	cfg, err := registry.LoadAndValidate(cfgPath)
	if err != nil {
		fmt.Fprintf(stderr, "warning: config at %s has issues: %v\n", cfgPath, err)
		// Continue install even with config issues.
	}

	// Check macOS accessibility.
	granted, accErr := platform.CheckAccessibility()
	if accErr != nil {
		fmt.Fprintf(stderr, "warning: could not check accessibility: %v\n", accErr)
	} else if !granted {
		fmt.Fprintln(stderr, platform.AccessibilityPrompt())
		fmt.Fprintln(stderr)
	}

	// Verify hotkey registration if config is loaded.
	if cfg != nil && cfg.LeaderKey != "" {
		verifyHotkeyRegistration(stdout, stderr, cfg.LeaderKey)
	}

	// Get binary path for autostart.
	binaryPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("detecting binary path: %w", err)
	}

	// Configure autostart.
	mgr := autostartManager()
	if err := mgr.Install(binaryPath); err != nil {
		return fmt.Errorf("configuring autostart: %w", err)
	}

	fmt.Fprintf(stdout, "orchestratr installed successfully\n")
	fmt.Fprintf(stdout, "  autostart: %s\n", mgr.Description())
	fmt.Fprintf(stdout, "  config: %s\n", cfgPath)

	if mgr.IsInstalled() {
		fmt.Fprintf(stdout, "  autostart configured — orchestratr will start at login\n")
	}

	return nil
}

// verifyHotkeyRegistration attempts to register the leader key and
// reports any issues.
func verifyHotkeyRegistration(stdout, stderr io.Writer, leaderKey string) {
	key, err := hotkey.ParseKey(leaderKey)
	if err != nil {
		fmt.Fprintf(stderr, "warning: leader key %q is invalid: %v\n", leaderKey, err)
		return
	}

	listener := hotkey.NewPlatformListener()
	warning, regErr := listener.Register(key)
	if regErr != nil {
		fmt.Fprintf(stderr, "warning: could not register leader key %q: %v\n", leaderKey, regErr)
		fmt.Fprintf(stderr, "  hotkeys may not work — consider changing leader_key in config\n")
	}
	if warning != "" {
		fmt.Fprintf(stderr, "warning: %s\n", warning)
	}
	_ = listener.Stop()
}

// runUninstall removes autostart configuration.
func runUninstall(stdout, stderr io.Writer) error {
	mgr := autostartManager()

	if !mgr.IsInstalled() {
		fmt.Fprintln(stdout, "orchestratr autostart is not installed")
		return nil
	}

	if err := mgr.Uninstall(); err != nil {
		return fmt.Errorf("removing autostart: %w", err)
	}

	fmt.Fprintf(stdout, "orchestratr autostart removed (%s)\n", mgr.Description())
	return nil
}
