package api

import (
	"sort"
	"sync"
	"time"
)

// AppState tracks the lifecycle state of a single registered app.
type AppState struct {
	Name       string     `json:"name"`
	Launched   bool       `json:"launched"`
	Ready      bool       `json:"ready"`
	LaunchedAt *time.Time `json:"launched_at,omitempty"`
	ReadyAt    *time.Time `json:"ready_at,omitempty"`
	Error      string     `json:"error,omitempty"`
	ErrorAt    *time.Time `json:"error_at,omitempty"`
}

// StateTracker tracks the lifecycle state of all registered apps.
type StateTracker struct {
	mu   sync.RWMutex
	apps map[string]*AppState
}

// NewStateTracker creates an empty StateTracker.
func NewStateTracker() *StateTracker {
	return &StateTracker{
		apps: make(map[string]*AppState),
	}
}

// SetLaunched marks an app as launched. Creates the entry if it
// does not exist.
func (st *StateTracker) SetLaunched(name string) {
	st.mu.Lock()
	defer st.mu.Unlock()

	now := time.Now()
	state, ok := st.apps[name]
	if !ok {
		state = &AppState{Name: name}
		st.apps[name] = state
	}
	state.Launched = true
	state.LaunchedAt = &now
	// Reset ready state on re-launch.
	state.Ready = false
	state.ReadyAt = nil
	// Clear any previous error on re-launch.
	state.Error = ""
	state.ErrorAt = nil
}

// SetReady marks an app as ready. Creates the entry if it does not
// exist.
func (st *StateTracker) SetReady(name string) {
	st.mu.Lock()
	defer st.mu.Unlock()

	now := time.Now()
	state, ok := st.apps[name]
	if !ok {
		state = &AppState{Name: name, Launched: true, LaunchedAt: &now}
		st.apps[name] = state
	}
	state.Ready = true
	state.ReadyAt = &now
}

// Get returns the state for an app. Returns nil if not tracked.
func (st *StateTracker) Get(name string) *AppState {
	st.mu.RLock()
	defer st.mu.RUnlock()

	s, ok := st.apps[name]
	if !ok {
		return nil
	}
	// Return a copy.
	cp := *s
	return &cp
}

// SetStopped marks an app as not launched and not ready, clearing
// timestamps. It is a no-op for unknown apps.
func (st *StateTracker) SetStopped(name string) {
	st.mu.Lock()
	defer st.mu.Unlock()

	state, ok := st.apps[name]
	if !ok {
		return
	}
	state.Launched = false
	state.Ready = false
	state.LaunchedAt = nil
	state.ReadyAt = nil
	state.Error = ""
	state.ErrorAt = nil
}

// SetError records an error message on an app's state. Creates the
// entry if it does not exist.
func (st *StateTracker) SetError(name, msg string) {
	st.mu.Lock()
	defer st.mu.Unlock()

	now := time.Now()
	state, ok := st.apps[name]
	if !ok {
		state = &AppState{Name: name}
		st.apps[name] = state
	}
	state.Error = msg
	state.ErrorAt = &now
}

// ClearError removes the error state from an app. It is a no-op for
// unknown apps.
func (st *StateTracker) ClearError(name string) {
	st.mu.Lock()
	defer st.mu.Unlock()

	state, ok := st.apps[name]
	if !ok {
		return
	}
	state.Error = ""
	state.ErrorAt = nil
}

// Sync removes state entries for apps not in the provided name list.
func (st *StateTracker) Sync(appNames []string) {
	allowed := make(map[string]bool, len(appNames))
	for _, name := range appNames {
		allowed[name] = true
	}

	st.mu.Lock()
	defer st.mu.Unlock()
	for name := range st.apps {
		if !allowed[name] {
			delete(st.apps, name)
		}
	}
}

// All returns a copy of all tracked app states, sorted by name for
// deterministic output.
func (st *StateTracker) All() []AppState {
	st.mu.RLock()
	defer st.mu.RUnlock()

	result := make([]AppState, 0, len(st.apps))
	for _, s := range st.apps {
		result = append(result, *s)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}
