package registry

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Load reads and parses a YAML config file at the given path.
// It returns the parsed Config and any error encountered during
// reading or parsing.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	// Normalize empty environment to "native" and set source to
	// "config" for all apps loaded from the main config file.
	for i := range cfg.Apps {
		if cfg.Apps[i].Environment == "" {
			cfg.Apps[i].Environment = "native"
		}
		cfg.Apps[i].Source = "config"
	}

	return &cfg, nil
}

// LoadAndValidate reads, parses, and validates a config file. It
// returns the Config and a combined error if validation fails.
func LoadAndValidate(path string) (*Config, error) {
	cfg, err := Load(path)
	if err != nil {
		return nil, err
	}

	if errs := ValidateConfig(cfg); len(errs) > 0 {
		msg := fmt.Sprintf("config validation failed with %d error(s):", len(errs))
		for _, e := range errs {
			msg += "\n  - " + e.Error()
		}
		return nil, fmt.Errorf("%s", msg)
	}

	return cfg, nil
}

// DefaultConfigPath returns the platform-appropriate path for the
// orchestratr config file.
//
//   - Linux/macOS: ~/.config/orchestratr/config.yml
//   - Windows: %APPDATA%/orchestratr/config.yml
func DefaultConfigPath() string {
	var base string

	switch runtime.GOOS {
	case "windows":
		base = os.Getenv("APPDATA")
		if base == "" {
			base = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
		}
	default:
		base = os.Getenv("XDG_CONFIG_HOME")
		if base == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				home = "."
			}
			base = filepath.Join(home, ".config")
		}
	}

	return filepath.Join(base, "orchestratr", "config.yml")
}

// EnsureDefaults creates the config directory and a default config
// file if they do not already exist. Returns the path to the config
// file and any error.
func EnsureDefaults(path string) (string, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return path, fmt.Errorf("creating config directory: %w", err)
	}

	if _, err := os.Stat(path); err == nil {
		// File already exists; nothing to do.
		return path, nil
	} else if !os.IsNotExist(err) {
		return path, fmt.Errorf("checking config file: %w", err)
	}

	cfg := DefaultConfig()
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return path, fmt.Errorf("marshaling default config: %w", err)
	}

	header := "# orchestratr configuration\n# See https://github.com/josiahH-cf/orchestratr for documentation.\n\n"
	if err := os.WriteFile(path, []byte(header+string(data)), 0o644); err != nil {
		return path, fmt.Errorf("writing default config: %w", err)
	}

	return path, nil
}

// AppsDirPath returns the platform-appropriate path for the apps.d
// drop-in directory. It is always a sibling of the config file directory.
func AppsDirPath(configPath string) string {
	return filepath.Join(filepath.Dir(configPath), "apps.d")
}

// DefaultAppsDirPath returns the platform-appropriate path for the
// apps.d drop-in directory using the default config location.
func DefaultAppsDirPath() string {
	return AppsDirPath(DefaultConfigPath())
}

// EnsureAppsDir creates the apps.d directory alongside the config
// file if it does not already exist. On Unix the directory is created
// with 0700 permissions.
func EnsureAppsDir(configPath string) error {
	dir := AppsDirPath(configPath)
	perm := os.FileMode(0o700)
	if runtime.GOOS == "windows" {
		perm = 0o755
	}
	if err := os.MkdirAll(dir, perm); err != nil {
		return fmt.Errorf("creating apps.d directory: %w", err)
	}
	return nil
}

// LoadAppEntry reads a single YAML file and parses it as an AppEntry.
func LoadAppEntry(path string) (*AppEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	var entry AppEntry
	if err := yaml.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	// Normalize environment.
	if entry.Environment == "" {
		entry.Environment = "native"
	}

	// Set source to the relative apps.d path.
	entry.Source = "apps.d/" + filepath.Base(path)
	return &entry, nil
}

// LoadAppsDir scans the given directory for *.yml files and parses
// each as a single AppEntry. Invalid files are collected as errors
// but do not prevent other files from loading.
func LoadAppsDir(dir string) ([]AppEntry, []error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, []error{fmt.Errorf("reading apps.d directory: %w", err)}
	}

	// Sort for deterministic order.
	var ymlFiles []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(strings.ToLower(name), ".yml") || strings.HasSuffix(strings.ToLower(name), ".yaml") {
			ymlFiles = append(ymlFiles, name)
		}
	}
	sort.Strings(ymlFiles)

	var apps []AppEntry
	var errs []error

	for _, name := range ymlFiles {
		path := filepath.Join(dir, name)
		entry, loadErr := LoadAppEntry(path)
		if loadErr != nil {
			errs = append(errs, loadErr)
			continue
		}
		apps = append(apps, *entry)
	}

	return apps, errs
}

// MergeApps merges base apps (from config.yml) with drop-in apps
// (from apps.d/). It implements the conflict resolution rules:
//   - Duplicate name between config and drop-in: config wins, warning logged
//   - Duplicate chord between config and drop-in: config wins, warning logged
//   - Duplicate name between two drop-ins: both rejected, error logged
//   - Duplicate chord between two drop-ins: both rejected, error logged
//
// Returns the merged app list and any warnings/errors encountered.
func MergeApps(base []AppEntry, dropins []AppEntry, logger *log.Logger) ([]AppEntry, []error) {
	var warnings []error

	// Build lookup sets from config.yml apps.
	baseNames := make(map[string]string)  // lowercase name -> source
	baseChords := make(map[string]string) // lowercase chord -> app name
	for _, app := range base {
		baseNames[strings.ToLower(app.Name)] = app.Source
		if app.Chord != "" {
			baseChords[strings.ToLower(app.Chord)] = app.Name
		}
	}

	// First pass: check drop-in conflicts with config.yml.
	var candidates []AppEntry
	for _, app := range dropins {
		nameLower := strings.ToLower(app.Name)
		chordLower := strings.ToLower(app.Chord)

		if _, exists := baseNames[nameLower]; exists {
			w := fmt.Errorf("drop-in %s: app name %q conflicts with config.yml entry; config.yml wins", app.Source, app.Name)
			warnings = append(warnings, w)
			if logger != nil {
				logger.Printf("warning: %v", w)
			}
			continue
		}

		if existingApp, exists := baseChords[chordLower]; exists {
			w := fmt.Errorf("drop-in %s: chord %q conflicts with config.yml app %q; config.yml wins", app.Source, app.Chord, existingApp)
			warnings = append(warnings, w)
			if logger != nil {
				logger.Printf("warning: %v", w)
			}
			continue
		}

		candidates = append(candidates, app)
	}

	// Second pass: check conflicts among drop-in candidates themselves.
	dropinNames := make(map[string]int)  // lowercase name -> index in candidates
	dropinChords := make(map[string]int) // lowercase chord -> index in candidates
	rejected := make(map[int]bool)       // indices to reject

	for i, app := range candidates {
		nameLower := strings.ToLower(app.Name)
		chordLower := strings.ToLower(app.Chord)

		if prevIdx, exists := dropinNames[nameLower]; exists {
			w := fmt.Errorf("drop-in conflict: app name %q in both %s and %s; both rejected",
				app.Name, candidates[prevIdx].Source, app.Source)
			warnings = append(warnings, w)
			if logger != nil {
				logger.Printf("error: %v", w)
			}
			rejected[prevIdx] = true
			rejected[i] = true
			continue
		}

		if prevIdx, exists := dropinChords[chordLower]; exists {
			w := fmt.Errorf("drop-in conflict: chord %q in both %s and %s; both rejected",
				app.Chord, candidates[prevIdx].Source, app.Source)
			warnings = append(warnings, w)
			if logger != nil {
				logger.Printf("error: %v", w)
			}
			rejected[prevIdx] = true
			rejected[i] = true
			continue
		}

		dropinNames[nameLower] = i
		dropinChords[chordLower] = i
	}

	// Build final list: base apps + surviving drop-in candidates.
	merged := make([]AppEntry, len(base))
	copy(merged, base)

	for i, app := range candidates {
		if !rejected[i] {
			merged = append(merged, app)
		}
	}

	return merged, warnings
}

// LoadWithDropins reads the config.yml file and merges in any drop-in
// apps from the apps.d/ directory. It validates the merged result.
// Parse errors in individual drop-in files are logged but do not
// prevent other files from loading.
func LoadWithDropins(configPath string, logger *log.Logger) (*Config, error) {
	cfg, err := Load(configPath)
	if err != nil {
		return nil, err
	}

	appsDir := AppsDirPath(configPath)
	dropins, loadErrs := LoadAppsDir(appsDir)
	for _, e := range loadErrs {
		if logger != nil {
			logger.Printf("warning: %v", e)
		}
	}

	if len(dropins) > 0 {
		merged, mergeWarnings := MergeApps(cfg.Apps, dropins, logger)
		_ = mergeWarnings // already logged by MergeApps
		cfg.Apps = merged
	}

	// Validate the merged config.
	if errs := ValidateConfig(cfg); len(errs) > 0 {
		msg := fmt.Sprintf("config validation failed with %d error(s):", len(errs))
		for _, e := range errs {
			msg += "\n  - " + e.Error()
		}
		return nil, fmt.Errorf("%s", msg)
	}

	return cfg, nil
}
