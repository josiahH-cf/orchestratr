//go:build (linux || windows) && !notray

package tray

import (
	"sync"
	"testing"
	"time"
)

// TestSystrayProvider_Implements_Provider verifies compile-time
// interface compliance.
func TestSystrayProvider_Implements_Provider(t *testing.T) {
	var _ Provider = (*SystrayProvider)(nil)
}

// newTestProvider builds a SystrayProvider with a fake runFn so tests
// do not require a real desktop environment. The fake runFn calls
// onReady immediately and then blocks until Setup's done channel is
// closed by Quit().
func newTestProvider() *SystrayProvider {
	p := &SystrayProvider{}
	p.runFn = func(onReady, onExit func()) {
		onReady()
		// Setup() writes p.done before launching this goroutine, so
		// reading p.done here (after onReady) is always safe and
		// refers to the correct channel that Quit() will close.
		p.mu.Lock()
		ch := p.done
		p.mu.Unlock()
		if ch != nil {
			<-ch
		}
		onExit()
	}
	// Inject a no-op quitter so tests don't close the package-level
	// systray.quitChan, which would corrupt state for other tests.
	p.quitFn = func() {}
	return p
}

func TestSystrayProvider_Setup_WithFakeRun(t *testing.T) {
	p := newTestProvider()

	if err := p.Setup(); err != nil {
		t.Fatalf("Setup() error = %v", err)
	}

	// done is now tracked by the provider; Quit should not panic.
	p.Quit()
}

func TestSystrayProvider_SetState_BeforeSetup(t *testing.T) {
	p := &SystrayProvider{}
	// Must not panic even when systray is not initialised.
	if err := p.SetState("running"); err != nil {
		t.Errorf("SetState before setup error = %v", err)
	}
}

func TestSystrayProvider_SetState_TracksValue(t *testing.T) {
	p := &SystrayProvider{}
	_ = p.SetState("paused")

	p.mu.Lock()
	got := p.state
	p.mu.Unlock()

	if got != "paused" {
		t.Errorf("state = %q, want %q", got, "paused")
	}
}

func TestSystrayProvider_Callbacks_RegisteredAndFired(t *testing.T) {
	p := &SystrayProvider{}

	var mu sync.Mutex
	calls := map[string]int{}
	inc := func(k string) func() {
		return func() {
			mu.Lock()
			calls[k]++
			mu.Unlock()
		}
	}

	p.OnPause(inc("pause"))
	p.OnResume(inc("resume"))
	p.OnQuit(inc("quit"))
	p.OnConfigure(inc("configure"))

	// Fire each callback directly (simulates a menu click reaching
	// the dispatch goroutine).
	p.mu.Lock()
	fnPause := p.onPauseFn
	fnResume := p.onResumeFn
	fnQuit := p.onQuitFn
	fnConfigure := p.onConfigureFn
	p.mu.Unlock()

	for name, fn := range map[string]func(){
		"pause":     fnPause,
		"resume":    fnResume,
		"quit":      fnQuit,
		"configure": fnConfigure,
	} {
		if fn == nil {
			t.Errorf("callback %q is nil after registration", name)
			continue
		}
		fn()
	}

	mu.Lock()
	defer mu.Unlock()
	for _, key := range []string{"pause", "resume", "quit", "configure"} {
		if calls[key] != 1 {
			t.Errorf("callback %q called %d times, want 1", key, calls[key])
		}
	}
}

func TestSystrayProvider_Quit_Idempotent(t *testing.T) {
	p := newTestProvider()
	_ = p.Setup()

	// Calling Quit twice must not panic.
	p.Quit()
	p.Quit()
}

func TestSystrayProvider_Quit_BeforeSetup(t *testing.T) {
	p := &SystrayProvider{}
	// Quit before Setup must not panic.
	p.Quit()
}

func TestSystrayProvider_Setup_Timeout(t *testing.T) {
	// A runFn that never calls onReady triggers the 3-second timeout.
	// We use a 50ms timeout override to keep the test fast.
	p := &SystrayProvider{}
	p.runFn = func(onReady, onExit func()) {
		// Never calls onReady — simulates a hung systray backend.
		time.Sleep(10 * time.Second)
	}

	// Temporarily reduce the timeout for test speed using the
	// package-level sentinel. Since we can't override the constant,
	// we test that Setup eventually returns an error. We rely on the
	// real 3s timeout — acceptable because this test is skipped in
	// short mode.
	if testing.Short() {
		t.Skip("skipping timeout test in short mode")
	}

	done := make(chan error, 1)
	go func() { done <- p.Setup() }()
	select {
	case err := <-done:
		if err == nil {
			t.Error("Setup() = nil, want timeout error")
		}
	case <-time.After(5 * time.Second):
		t.Error("Setup() did not return within 5s")
	}
}

func TestNewPlatformProvider_ReturnsProvider(t *testing.T) {
	p := NewPlatformProvider()
	if p == nil {
		t.Fatal("NewPlatformProvider() returned nil")
	}
}

func TestNewPlatformProvider_HeadlessOnNoDisplay(t *testing.T) {
	// Override the display check by unsetting display env vars and
	// re-running the detection logic directly.
	t.Setenv("DBUS_SESSION_BUS_ADDRESS", "")
	t.Setenv("DISPLAY", "")
	t.Setenv("WAYLAND_DISPLAY", "")

	if isDisplayAvailable() {
		t.Skip("display still detected despite clearing env vars (system socket exists)")
	}

	p := NewPlatformProvider()
	if _, ok := p.(*HeadlessProvider); !ok {
		t.Errorf("NewPlatformProvider() = %T, want *HeadlessProvider when no display", p)
	}
}

func TestGenerateIcon_ValidPNG(t *testing.T) {
	data := generateIcon()
	if len(data) == 0 {
		t.Fatal("generateIcon() returned empty bytes")
	}
	// PNG magic bytes: \x89PNG
	if len(data) < 4 || data[0] != 0x89 || data[1] != 'P' || data[2] != 'N' || data[3] != 'G' {
		t.Errorf("generateIcon() does not start with PNG magic bytes")
	}
}
