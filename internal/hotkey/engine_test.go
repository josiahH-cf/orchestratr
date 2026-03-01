package hotkey

import (
	"log"
	"os"
	"sync"
	"testing"
	"time"
)

// --- testListener: a mock Listener for unit tests ---

type testListener struct {
	mu         sync.Mutex
	leader     Key
	registered bool
	stopped    bool
	events     chan<- KeyEvent
	stopCh     chan struct{}
	info       PlatformInfo
}

func newTestListener() *testListener {
	return &testListener{
		stopCh: make(chan struct{}),
		info:   PlatformInfo{OS: "test", Method: "mock"},
	}
}

func (l *testListener) Info() PlatformInfo { return l.info }

func (l *testListener) Register(leader Key) (string, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.leader = leader
	l.registered = true
	return CheckConflicts(leader), nil
}

func (l *testListener) Start(events chan<- KeyEvent) error {
	l.mu.Lock()
	l.events = events
	l.mu.Unlock()
	<-l.stopCh
	return nil
}

func (l *testListener) Stop() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.stopped {
		l.stopped = true
		close(l.stopCh)
	}
	return nil
}

// inject sends a key event into the listener from the test.
func (l *testListener) inject(k Key, pressed bool) {
	l.mu.Lock()
	ch := l.events
	l.mu.Unlock()
	if ch != nil {
		ch <- KeyEvent{Key: k, Pressed: pressed}
	}
}

// --- Tests ---

func testLogger() *log.Logger {
	return log.New(os.Stderr, "test-hotkey: ", log.LstdFlags)
}

func TestNewEngine_Valid(t *testing.T) {
	listener := newTestListener()
	e, err := NewEngine(EngineConfig{
		LeaderKey:      "ctrl+space",
		ChordTimeoutMs: 1000,
		Chords:         []Chord{{Key: Key{Code: "e"}, Action: "espansr"}},
		OnAction:       func(string) {},
		Logger:         testLogger(),
	}, listener)
	if err != nil {
		t.Fatalf("NewEngine error: %v", err)
	}
	if !e.Leader().Equal(Key{Modifiers: ModCtrl, Code: "space"}) {
		t.Errorf("leader = %v, want ctrl+space", e.Leader())
	}
	if e.State() != StateIdle {
		t.Errorf("initial state = %v, want idle", e.State())
	}
}

func TestNewEngine_Errors(t *testing.T) {
	listener := newTestListener()

	tests := []struct {
		name string
		cfg  EngineConfig
		l    Listener
		want string
	}{
		{
			name: "nil OnAction",
			cfg:  EngineConfig{LeaderKey: "ctrl+space", OnAction: nil},
			l:    listener,
			want: "OnAction",
		},
		{
			name: "nil listener",
			cfg:  EngineConfig{LeaderKey: "ctrl+space", OnAction: func(string) {}},
			l:    nil,
			want: "listener",
		},
		{
			name: "invalid leader key",
			cfg:  EngineConfig{LeaderKey: "", OnAction: func(string) {}},
			l:    listener,
			want: "empty",
		},
		{
			name: "duplicate chord",
			cfg: EngineConfig{
				LeaderKey: "ctrl+space",
				OnAction:  func(string) {},
				Chords: []Chord{
					{Key: Key{Code: "e"}, Action: "a"},
					{Key: Key{Code: "e"}, Action: "b"},
				},
			},
			l:    listener,
			want: "duplicate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewEngine(tt.cfg, tt.l)
			if err == nil {
				t.Fatal("expected error")
			}
			if !containsLower(err.Error(), tt.want) {
				t.Errorf("error = %q, want substring %q", err.Error(), tt.want)
			}
		})
	}
}

func TestEngine_ChordMatch(t *testing.T) {
	listener := newTestListener()
	leader := Key{Modifiers: ModCtrl, Code: "space"}
	chordE := Key{Code: "e"}

	var dispatched string
	var mu sync.Mutex
	dispatches := make(chan string, 1)

	e, err := NewEngine(EngineConfig{
		LeaderKey:      "ctrl+space",
		ChordTimeoutMs: 2000,
		Chords:         []Chord{{Key: chordE, Action: "espansr"}},
		OnAction: func(action string) {
			mu.Lock()
			dispatched = action
			mu.Unlock()
			dispatches <- action
		},
		Logger: testLogger(),
	}, listener)
	if err != nil {
		t.Fatalf("NewEngine error: %v", err)
	}

	if err := e.Start(); err != nil {
		t.Fatalf("Start error: %v", err)
	}
	defer e.Stop()

	// Give the loop a moment to start.
	time.Sleep(20 * time.Millisecond)

	// Press leader key.
	listener.inject(leader, true)
	time.Sleep(20 * time.Millisecond)

	if e.State() != StateChordWait {
		t.Fatalf("state = %v, want chord_wait", e.State())
	}

	// Press chord key.
	listener.inject(chordE, true)

	select {
	case action := <-dispatches:
		if action != "espansr" {
			t.Errorf("dispatched action = %q, want %q", action, "espansr")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for dispatch")
	}

	// Engine should return to idle.
	time.Sleep(20 * time.Millisecond)
	if e.State() != StateIdle {
		t.Errorf("state after chord = %v, want idle", e.State())
	}

	mu.Lock()
	got := dispatched
	mu.Unlock()
	if got != "espansr" {
		t.Errorf("dispatched = %q, want %q", got, "espansr")
	}
}

func TestEngine_UnrecognizedChord(t *testing.T) {
	listener := newTestListener()
	leader := Key{Modifiers: ModCtrl, Code: "space"}

	dispatched := false
	e, err := NewEngine(EngineConfig{
		LeaderKey:      "ctrl+space",
		ChordTimeoutMs: 2000,
		Chords:         []Chord{{Key: Key{Code: "e"}, Action: "espansr"}},
		OnAction: func(string) {
			dispatched = true
		},
		Logger: testLogger(),
	}, listener)
	if err != nil {
		t.Fatalf("NewEngine error: %v", err)
	}

	if err := e.Start(); err != nil {
		t.Fatalf("Start error: %v", err)
	}
	defer e.Stop()

	time.Sleep(20 * time.Millisecond)

	// Press leader key.
	listener.inject(leader, true)
	time.Sleep(20 * time.Millisecond)

	// Press unrecognized chord.
	listener.inject(Key{Code: "z"}, true)
	time.Sleep(50 * time.Millisecond)

	// Should return to idle without dispatch.
	if e.State() != StateIdle {
		t.Errorf("state = %v, want idle", e.State())
	}
	if dispatched {
		t.Error("expected no dispatch for unrecognized chord")
	}
}

func TestEngine_Timeout(t *testing.T) {
	listener := newTestListener()
	leader := Key{Modifiers: ModCtrl, Code: "space"}

	dispatched := false
	e, err := NewEngine(EngineConfig{
		LeaderKey:      "ctrl+space",
		ChordTimeoutMs: 100, // short timeout for test
		Chords:         []Chord{{Key: Key{Code: "e"}, Action: "espansr"}},
		OnAction: func(string) {
			dispatched = true
		},
		Logger: testLogger(),
	}, listener)
	if err != nil {
		t.Fatalf("NewEngine error: %v", err)
	}

	if err := e.Start(); err != nil {
		t.Fatalf("Start error: %v", err)
	}
	defer e.Stop()

	time.Sleep(20 * time.Millisecond)

	// Press leader key.
	listener.inject(leader, true)
	time.Sleep(20 * time.Millisecond)
	if e.State() != StateChordWait {
		t.Fatalf("state = %v, want chord_wait", e.State())
	}

	// Wait for timeout to expire.
	time.Sleep(200 * time.Millisecond)

	if e.State() != StateIdle {
		t.Errorf("state after timeout = %v, want idle", e.State())
	}
	if dispatched {
		t.Error("expected no dispatch after timeout")
	}
}

func TestEngine_KeyUpIgnored(t *testing.T) {
	listener := newTestListener()
	leader := Key{Modifiers: ModCtrl, Code: "space"}

	dispatched := false
	e, err := NewEngine(EngineConfig{
		LeaderKey:      "ctrl+space",
		ChordTimeoutMs: 1000,
		Chords:         []Chord{{Key: Key{Code: "e"}, Action: "espansr"}},
		OnAction: func(string) {
			dispatched = true
		},
		Logger: testLogger(),
	}, listener)
	if err != nil {
		t.Fatalf("NewEngine error: %v", err)
	}

	if err := e.Start(); err != nil {
		t.Fatalf("Start error: %v", err)
	}
	defer e.Stop()

	time.Sleep(20 * time.Millisecond)

	// Send key-up for leader — should be ignored.
	listener.inject(leader, false)
	time.Sleep(50 * time.Millisecond)

	if e.State() != StateIdle {
		t.Errorf("state = %v, want idle (key-up should be ignored)", e.State())
	}
	if dispatched {
		t.Error("no dispatch expected from key-up")
	}
}

func TestEngine_StopClean(t *testing.T) {
	listener := newTestListener()

	e, err := NewEngine(EngineConfig{
		LeaderKey:      "ctrl+space",
		ChordTimeoutMs: 1000,
		OnAction:       func(string) {},
		Logger:         testLogger(),
	}, listener)
	if err != nil {
		t.Fatalf("NewEngine error: %v", err)
	}

	if err := e.Start(); err != nil {
		t.Fatalf("Start error: %v", err)
	}

	// Double stop should not panic.
	if err := e.Stop(); err != nil {
		t.Errorf("first Stop() error: %v", err)
	}
	if err := e.Stop(); err != nil {
		t.Errorf("second Stop() error: %v", err)
	}

	if e.State() != StateIdle {
		t.Errorf("state after stop = %v, want idle", e.State())
	}
}

func TestEngine_DefaultTimeout(t *testing.T) {
	listener := newTestListener()
	e, err := NewEngine(EngineConfig{
		LeaderKey:      "ctrl+space",
		ChordTimeoutMs: 0, // should default to 2000
		OnAction:       func(string) {},
		Logger:         testLogger(),
	}, listener)
	if err != nil {
		t.Fatalf("NewEngine error: %v", err)
	}
	if e.timeout != 2000*time.Millisecond {
		t.Errorf("timeout = %v, want 2s", e.timeout)
	}
}

func TestEngine_MultipleChords(t *testing.T) {
	listener := newTestListener()
	leader := Key{Modifiers: ModCtrl, Code: "space"}

	dispatches := make(chan string, 4)
	e, err := NewEngine(EngineConfig{
		LeaderKey:      "ctrl+space",
		ChordTimeoutMs: 2000,
		Chords: []Chord{
			{Key: Key{Code: "e"}, Action: "espansr"},
			{Key: Key{Code: "f"}, Action: "firefox"},
			{Key: Key{Code: "t"}, Action: "terminal"},
		},
		OnAction: func(action string) {
			dispatches <- action
		},
		Logger: testLogger(),
	}, listener)
	if err != nil {
		t.Fatalf("NewEngine error: %v", err)
	}

	if err := e.Start(); err != nil {
		t.Fatalf("Start error: %v", err)
	}
	defer e.Stop()

	time.Sleep(20 * time.Millisecond)

	// First chord: leader → f
	listener.inject(leader, true)
	time.Sleep(20 * time.Millisecond)
	listener.inject(Key{Code: "f"}, true)

	select {
	case action := <-dispatches:
		if action != "firefox" {
			t.Errorf("first dispatch = %q, want %q", action, "firefox")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out")
	}

	time.Sleep(20 * time.Millisecond)

	// Second chord: leader → t
	listener.inject(leader, true)
	time.Sleep(20 * time.Millisecond)
	listener.inject(Key{Code: "t"}, true)

	select {
	case action := <-dispatches:
		if action != "terminal" {
			t.Errorf("second dispatch = %q, want %q", action, "terminal")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out")
	}
}

func TestEngineState_String(t *testing.T) {
	if StateIdle.String() != "idle" {
		t.Errorf("StateIdle.String() = %q", StateIdle.String())
	}
	if StateChordWait.String() != "chord_wait" {
		t.Errorf("StateChordWait.String() = %q", StateChordWait.String())
	}
	if EngineState(99).String() != "unknown(99)" {
		t.Errorf("unknown state string = %q", EngineState(99).String())
	}
}
func TestEngine_Trigger(t *testing.T) {
	listener := newTestListener()
	chordE := Key{Code: "e"}

	dispatches := make(chan string, 1)
	e, err := NewEngine(EngineConfig{
		LeaderKey:      "ctrl+space",
		ChordTimeoutMs: 2000,
		Chords:         []Chord{{Key: chordE, Action: "espansr"}},
		OnAction: func(action string) {
			dispatches <- action
		},
		Logger: testLogger(),
	}, listener)
	if err != nil {
		t.Fatalf("NewEngine error: %v", err)
	}

	if err := e.Start(); err != nil {
		t.Fatalf("Start error: %v", err)
	}
	defer e.Stop()

	time.Sleep(20 * time.Millisecond)

	// Trigger simulates leader key press via API.
	if err := e.Trigger(); err != nil {
		t.Fatalf("Trigger error: %v", err)
	}
	time.Sleep(20 * time.Millisecond)

	if e.State() != StateChordWait {
		t.Fatalf("state = %v, want chord_wait after Trigger", e.State())
	}

	// Now send chord via listener — should dispatch.
	listener.inject(chordE, true)

	select {
	case action := <-dispatches:
		if action != "espansr" {
			t.Errorf("dispatched = %q, want %q", action, "espansr")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for dispatch after Trigger")
	}
}

func TestEngine_TriggerAfterStop(t *testing.T) {
	listener := newTestListener()
	e, err := NewEngine(EngineConfig{
		LeaderKey:      "ctrl+space",
		ChordTimeoutMs: 1000,
		OnAction:       func(string) {},
		Logger:         testLogger(),
	}, listener)
	if err != nil {
		t.Fatalf("NewEngine error: %v", err)
	}

	if err := e.Start(); err != nil {
		t.Fatalf("Start error: %v", err)
	}
	e.Stop()

	// Trigger should return an error after stop.
	if err := e.Trigger(); err == nil {
		t.Error("Trigger after Stop should return error")
	}
}

func TestEngine_SwapChords(t *testing.T) {
	listener := newTestListener()
	leader := Key{Modifiers: ModCtrl, Code: "space"}

	dispatches := make(chan string, 1)
	e, err := NewEngine(EngineConfig{
		LeaderKey:      "ctrl+space",
		ChordTimeoutMs: 2000,
		Chords:         []Chord{{Key: Key{Code: "e"}, Action: "espansr"}},
		OnAction: func(action string) {
			dispatches <- action
		},
		Logger: testLogger(),
	}, listener)
	if err != nil {
		t.Fatalf("NewEngine error: %v", err)
	}

	if err := e.Start(); err != nil {
		t.Fatalf("Start error: %v", err)
	}
	defer e.Stop()

	time.Sleep(20 * time.Millisecond)

	// Swap to a new chord set.
	if err := e.SwapChords([]Chord{
		{Key: Key{Code: "f"}, Action: "firefox"},
	}); err != nil {
		t.Fatalf("SwapChords error: %v", err)
	}

	// Old chord should no longer work.
	listener.inject(leader, true)
	time.Sleep(20 * time.Millisecond)
	listener.inject(Key{Code: "e"}, true)
	time.Sleep(50 * time.Millisecond)

	select {
	case action := <-dispatches:
		t.Fatalf("old chord should not dispatch, got %q", action)
	default:
		// good — no dispatch
	}

	// New chord should work.
	listener.inject(leader, true)
	time.Sleep(20 * time.Millisecond)
	listener.inject(Key{Code: "f"}, true)

	select {
	case action := <-dispatches:
		if action != "firefox" {
			t.Errorf("dispatched = %q, want %q", action, "firefox")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for dispatch")
	}
}

func TestEngine_Pause(t *testing.T) {
	listener := newTestListener()
	leader := Key{Modifiers: ModCtrl, Code: "space"}
	chordE := Key{Code: "e"}

	dispatches := make(chan string, 1)
	e, err := NewEngine(EngineConfig{
		LeaderKey:      "ctrl+space",
		ChordTimeoutMs: 2000,
		Chords:         []Chord{{Key: chordE, Action: "espansr"}},
		OnAction: func(action string) {
			dispatches <- action
		},
		Logger: testLogger(),
	}, listener)
	if err != nil {
		t.Fatalf("NewEngine error: %v", err)
	}

	if err := e.Start(); err != nil {
		t.Fatalf("Start error: %v", err)
	}
	defer e.Stop()

	time.Sleep(20 * time.Millisecond)

	// Pause the engine.
	e.Pause()
	if !e.Paused() {
		t.Fatal("expected Paused() = true")
	}

	// Leader key should be ignored while paused.
	listener.inject(leader, true)
	time.Sleep(50 * time.Millisecond)

	if e.State() != StateIdle {
		t.Errorf("state = %v, want idle (paused should ignore leader)", e.State())
	}

	// Trigger should fail while paused.
	if err := e.Trigger(); err == nil {
		t.Error("Trigger should fail while paused")
	}

	// Resume the engine.
	e.Resume()
	if e.Paused() {
		t.Fatal("expected Paused() = false after Resume")
	}

	// Hotkeys should work again.
	listener.inject(leader, true)
	time.Sleep(20 * time.Millisecond)
	listener.inject(chordE, true)

	select {
	case action := <-dispatches:
		if action != "espansr" {
			t.Errorf("dispatched = %q, want %q", action, "espansr")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for dispatch after Resume")
	}
}

func TestEngine_PauseIdempotent(t *testing.T) {
	listener := newTestListener()
	e, err := NewEngine(EngineConfig{
		LeaderKey:      "ctrl+space",
		ChordTimeoutMs: 1000,
		OnAction:       func(string) {},
		Logger:         testLogger(),
	}, listener)
	if err != nil {
		t.Fatalf("NewEngine error: %v", err)
	}

	// Double Pause should not panic.
	e.Pause()
	e.Pause()
	if !e.Paused() {
		t.Error("expected Paused() = true")
	}

	// Double Resume should not panic.
	e.Resume()
	e.Resume()
	if e.Paused() {
		t.Error("expected Paused() = false")
	}
}

func TestEngine_SwapChords_DuplicateRejected(t *testing.T) {
	listener := newTestListener()
	e, err := NewEngine(EngineConfig{
		LeaderKey:      "ctrl+space",
		ChordTimeoutMs: 1000,
		Chords:         []Chord{{Key: Key{Code: "e"}, Action: "espansr"}},
		OnAction:       func(string) {},
		Logger:         testLogger(),
	}, listener)
	if err != nil {
		t.Fatalf("NewEngine error: %v", err)
	}

	err = e.SwapChords([]Chord{
		{Key: Key{Code: "a"}, Action: "app1"},
		{Key: Key{Code: "a"}, Action: "app2"},
	})
	if err == nil {
		t.Fatal("SwapChords should reject duplicates")
	}
	if !containsLower(err.Error(), "duplicate") {
		t.Errorf("error = %q, want 'duplicate'", err.Error())
	}

	// Original chords should be unchanged.
	e.mu.RLock()
	_, hasE := e.chords["e"]
	e.mu.RUnlock()
	if !hasE {
		t.Error("original chord 'e' should be preserved after failed swap")
	}
}
