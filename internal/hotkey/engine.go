package hotkey

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// EngineState represents the current state of the hotkey engine.
type EngineState int

const (
	// StateIdle means the engine is waiting for the leader key.
	StateIdle EngineState = iota
	// StateChordWait means the leader was pressed and the engine is
	// capturing chord keystrokes within the timeout window.
	StateChordWait
)

// String returns a human-readable state name.
func (s EngineState) String() string {
	switch s {
	case StateIdle:
		return "idle"
	case StateChordWait:
		return "chord_wait"
	default:
		return fmt.Sprintf("unknown(%d)", int(s))
	}
}

// Chord maps a key sequence to a named action.
type Chord struct {
	Key    Key
	Action string // e.g. app name to launch
}

// ActionFunc is called when a chord is matched. The argument is the
// action string from the matched Chord.
type ActionFunc func(action string)

// EngineConfig holds configuration for the hotkey Engine.
type EngineConfig struct {
	// LeaderKey is the global activation key (e.g. "ctrl+space").
	LeaderKey string
	// ChordTimeoutMs is how long (in ms) to wait for chord input
	// after the leader key is pressed. Default: 2000.
	ChordTimeoutMs int
	// Chords is the set of registered chord mappings.
	Chords []Chord
	// OnAction is called when a chord matches. Must not be nil.
	OnAction ActionFunc
	// Logger receives diagnostic messages. If nil, a default is used.
	Logger *log.Logger
}

// Engine is the hotkey state machine that coordinates a Listener with
// chord matching and action dispatch.
type Engine struct {
	leader   Key
	timeout  time.Duration
	chords   map[string]string // key.String() → action
	onAction ActionFunc
	logger   *log.Logger

	listener Listener
	events   chan KeyEvent
	stopCh   chan struct{}
	done     chan struct{}

	mu     sync.RWMutex
	state  EngineState
	paused bool
}

// NewEngine creates a hotkey Engine from the given config and Listener.
// It parses the leader key and builds the chord lookup table.
// Returns an error if the config is invalid.
func NewEngine(cfg EngineConfig, listener Listener) (*Engine, error) {
	if cfg.OnAction == nil {
		return nil, fmt.Errorf("EngineConfig.OnAction must not be nil")
	}
	if listener == nil {
		return nil, fmt.Errorf("listener must not be nil")
	}

	leader, err := ParseKey(cfg.LeaderKey)
	if err != nil {
		return nil, fmt.Errorf("invalid leader key %q: %w", cfg.LeaderKey, err)
	}

	timeout := time.Duration(cfg.ChordTimeoutMs) * time.Millisecond
	if timeout <= 0 {
		timeout = 2000 * time.Millisecond
	}

	logger := cfg.Logger
	if logger == nil {
		logger = log.New(os.Stderr, "hotkey: ", log.LstdFlags)
	}

	chordMap := make(map[string]string, len(cfg.Chords))
	for _, c := range cfg.Chords {
		canonical := c.Key.String()
		if _, dup := chordMap[canonical]; dup {
			return nil, fmt.Errorf("duplicate chord %q", canonical)
		}
		chordMap[canonical] = c.Action
	}

	return &Engine{
		leader:   leader,
		timeout:  timeout,
		chords:   chordMap,
		onAction: cfg.OnAction,
		logger:   logger,
		listener: listener,
		events:   make(chan KeyEvent, 16),
		stopCh:   make(chan struct{}),
		done:     make(chan struct{}),
		state:    StateIdle,
	}, nil
}

// SwapChords atomically replaces the engine's chord lookup table.
// Returns an error if the new chord set contains duplicates.
func (e *Engine) SwapChords(chords []Chord) error {
	newMap := make(map[string]string, len(chords))
	for _, c := range chords {
		canonical := c.Key.String()
		if _, dup := newMap[canonical]; dup {
			return fmt.Errorf("duplicate chord %q", canonical)
		}
		newMap[canonical] = c.Action
	}

	e.mu.Lock()
	e.chords = newMap
	e.mu.Unlock()
	return nil
}

// State returns the engine's current state.
func (e *Engine) State() EngineState {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.state
}

// Leader returns the parsed leader key.
func (e *Engine) Leader() Key {
	return e.leader
}

// Start registers the leader key with the listener and begins the
// event processing loop. It returns after the listener is registered
// and the loop goroutine is running. Call Stop to shut down.
func (e *Engine) Start() error {
	warning, err := e.listener.Register(e.leader)
	if err != nil {
		return fmt.Errorf("registering leader key: %w", err)
	}
	if warning != "" {
		e.logger.Printf("warning: %s", warning)
	}

	info := e.listener.Info()
	e.logger.Printf("hotkey engine started: platform=%s, leader=%s", info, e.leader)

	// Start the listener in a goroutine.
	go func() {
		if listenErr := e.listener.Start(e.events); listenErr != nil {
			e.logger.Printf("listener error: %v", listenErr)
		}
	}()

	// Start the event processing loop.
	go e.loop()

	return nil
}

// Stop shuts down the engine, releasing the hotkey registration.
func (e *Engine) Stop() error {
	select {
	case <-e.stopCh:
		// Already stopped.
		return nil
	default:
		close(e.stopCh)
	}

	err := e.listener.Stop()

	// Wait for the loop to finish.
	<-e.done

	e.mu.Lock()
	e.state = StateIdle
	e.mu.Unlock()

	e.logger.Println("hotkey engine stopped")
	return err
}

// Pause disables hotkey processing without stopping the engine.
// Key events are silently discarded while paused.
func (e *Engine) Pause() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if !e.paused {
		e.paused = true
		e.logger.Println("hotkey engine paused")
	}
}

// Resume re-enables hotkey processing after a Pause.
func (e *Engine) Resume() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.paused {
		e.paused = false
		e.logger.Println("hotkey engine resumed")
	}
}

// Paused reports whether the engine is currently paused.
func (e *Engine) Paused() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.paused
}

// Trigger simulates a leader key press via the API. This puts the
// engine into ChordWait state as if the physical leader key were
// pressed. It is used by the Wayland manual fallback (orchestratr
// trigger) and can also be used for testing.
func (e *Engine) Trigger() error {
	select {
	case <-e.stopCh:
		return fmt.Errorf("engine is stopped")
	default:
	}

	e.mu.RLock()
	paused := e.paused
	e.mu.RUnlock()
	if paused {
		return fmt.Errorf("engine is paused")
	}

	e.events <- KeyEvent{Key: e.leader, Pressed: true}
	return nil
}

// loop is the main event processing goroutine.
func (e *Engine) loop() {
	defer close(e.done)

	var chordTimer *time.Timer
	var chordTimeout <-chan time.Time

	for {
		select {
		case <-e.stopCh:
			if chordTimer != nil {
				chordTimer.Stop()
			}
			e.listener.UngrabKeyboard()
			return

		case evt, ok := <-e.events:
			if !ok {
				return
			}
			if !evt.Pressed {
				continue // only process key-down events
			}

			e.mu.RLock()
			st := e.state
			paused := e.paused
			e.mu.RUnlock()

			if paused {
				continue // silently discard events while paused
			}

			switch st {
			case StateIdle:
				if evt.Key.Equal(e.leader) {
					e.mu.Lock()
					e.state = StateChordWait
					e.mu.Unlock()

					// Grab the entire keyboard so chord keystrokes
					// are captured exclusively and not leaked to the
					// focused application.
					if grabErr := e.listener.GrabKeyboard(); grabErr != nil {
						e.logger.Printf("warning: keyboard grab failed: %v", grabErr)
					}

					chordTimer = time.NewTimer(e.timeout)
					chordTimeout = chordTimer.C
					e.logger.Println("leader key activated, waiting for chord")
				}

			case StateChordWait:
				if chordTimer != nil {
					chordTimer.Stop()
					chordTimer = nil
					chordTimeout = nil
				}

				// Release the keyboard grab before dispatching so the
				// launched app can receive normal input immediately.
				e.listener.UngrabKeyboard()

				canonical := evt.Key.String()
				e.mu.RLock()
				action, ok := e.chords[canonical]
				e.mu.RUnlock()
				if ok {
					e.logger.Printf("chord matched: %s → %s", canonical, action)
					e.onAction(action)
				} else {
					e.logger.Printf("unrecognized chord: %s (ignoring)", canonical)
				}

				e.mu.Lock()
				e.state = StateIdle
				e.mu.Unlock()
			}

		case <-chordTimeout:
			e.logger.Println("chord timeout expired")
			e.listener.UngrabKeyboard()
			chordTimer = nil
			chordTimeout = nil

			e.mu.Lock()
			e.state = StateIdle
			e.mu.Unlock()
		}
	}
}
