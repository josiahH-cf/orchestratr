package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

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
