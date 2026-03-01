// Package registry provides the app registry: YAML config parsing,
// validation, hot-reload, and thread-safe query access.
package registry

// Config represents the top-level orchestratr configuration file.
type Config struct {
	LeaderKey      string     `yaml:"leader_key"`
	ChordTimeoutMs int        `yaml:"chord_timeout_ms"`
	APIPort        int        `yaml:"api_port"`
	LogLevel       string     `yaml:"log_level"`
	Apps           []AppEntry `yaml:"apps"`
}

// AppEntry describes a single application registered in the config.
type AppEntry struct {
	Name           string `yaml:"name"`
	Chord          string `yaml:"chord"`
	Command        string `yaml:"command"`
	Environment    string `yaml:"environment"`
	Description    string `yaml:"description,omitempty"`
	ReadyCmd       string `yaml:"ready_cmd,omitempty"`
	ReadyTimeoutMs int    `yaml:"ready_timeout_ms,omitempty"`
}

// DefaultConfig returns a Config with sensible default values and an
// empty app list.
func DefaultConfig() Config {
	return Config{
		LeaderKey:      "ctrl+space",
		ChordTimeoutMs: 2000,
		APIPort:        9876,
		LogLevel:       "info",
		Apps:           []AppEntry{},
	}
}
