package registry

import (
	"fmt"
	"strings"
	"sync"
)

// Registry provides thread-safe access to the app configuration.
// It wraps a Config and protects concurrent reads/writes with a
// read-write mutex.
type Registry struct {
	mu  sync.RWMutex
	cfg Config
}

// NewRegistry creates a Registry from the given Config. The config
// should already be validated before calling this function.
func NewRegistry(cfg Config) *Registry {
	return &Registry{cfg: cfg}
}

// Config returns a copy of the current configuration.
func (r *Registry) Config() Config {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy so the caller cannot mutate the registry.
	cp := r.cfg
	cp.Apps = make([]AppEntry, len(r.cfg.Apps))
	copy(cp.Apps, r.cfg.Apps)
	return cp
}

// Apps returns a copy of the current app entries.
func (r *Registry) Apps() []AppEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	apps := make([]AppEntry, len(r.cfg.Apps))
	copy(apps, r.cfg.Apps)
	return apps
}

// FindByChord returns the app entry matching the given chord key
// (case-insensitive). Returns false if no match is found.
func (r *Registry) FindByChord(chord string) (AppEntry, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, app := range r.cfg.Apps {
		if strings.EqualFold(app.Chord, chord) {
			return app, true
		}
	}
	return AppEntry{}, false
}

// FindByName returns the app entry matching the given name
// (case-insensitive). Returns false if no match is found.
func (r *Registry) FindByName(name string) (AppEntry, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, app := range r.cfg.Apps {
		if strings.EqualFold(app.Name, name) {
			return app, true
		}
	}
	return AppEntry{}, false
}

// Swap atomically replaces the registry's configuration. The new
// config should be validated before calling Swap.
func (r *Registry) Swap(cfg Config) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cfg = cfg
}

// Len returns the number of registered apps.
func (r *Registry) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.cfg.Apps)
}

// String returns a human-readable summary of the registry.
func (r *Registry) String() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return fmt.Sprintf("Registry(%d apps)", len(r.cfg.Apps))
}
