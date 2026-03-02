// Package registry provides the app registry: YAML config parsing,
// validation, hot-reload, and thread-safe query access.
package registry

// Config represents the top-level orchestratr configuration file.
type Config struct {
	LeaderKey      string     `yaml:"leader_key" json:"leader_key"`
	ChordTimeoutMs int        `yaml:"chord_timeout_ms" json:"chord_timeout_ms"`
	APIPort        int        `yaml:"api_port" json:"api_port"`
	LogLevel       string     `yaml:"log_level" json:"log_level"`
	Apps           []AppEntry `yaml:"apps" json:"apps"`
}

// AppEntry describes a single application registered in the config.
type AppEntry struct {
	Name           string `yaml:"name" json:"name"`
	Chord          string `yaml:"chord" json:"chord"`
	Command        string `yaml:"command" json:"command"`
	Environment    string `yaml:"environment" json:"environment"`
	Description    string `yaml:"description,omitempty" json:"description,omitempty"`
	ReadyCmd       string `yaml:"ready_cmd,omitempty" json:"ready_cmd,omitempty"`
	ReadyTimeoutMs int    `yaml:"ready_timeout_ms,omitempty" json:"ready_timeout_ms,omitempty"`
	Detached       bool   `yaml:"detached,omitempty" json:"detached,omitempty"`
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
