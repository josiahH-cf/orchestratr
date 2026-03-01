package hotkey

import (
	"errors"
	"os"
	"runtime"
	"testing"
)

func TestNewStubListener_Info(t *testing.T) {
	l := NewStubListener()
	info := l.Info()
	if info.OS != runtime.GOOS {
		t.Errorf("Info().OS = %q, want %q", info.OS, runtime.GOOS)
	}
	if info.Method != "stub" {
		t.Errorf("Info().Method = %q, want %q", info.Method, "stub")
	}
}

func TestNewStubListener_Register(t *testing.T) {
	l := NewStubListener()
	warning, err := l.Register(Key{Modifiers: ModCtrl, Code: "space"})
	if err != nil {
		t.Fatalf("Register error: %v", err)
	}
	// ctrl+space is a known conflict, so warning should be non-empty.
	if warning == "" {
		t.Error("expected conflict warning for ctrl+space")
	}

	// A non-conflicting key should produce no warning.
	l2 := NewStubListener()
	warning2, err := l2.Register(Key{Modifiers: ModCtrl | ModAlt, Code: "o"})
	if err != nil {
		t.Fatalf("Register error: %v", err)
	}
	if warning2 != "" {
		t.Errorf("unexpected warning: %q", warning2)
	}
}

func TestNewStubListener_StartReturnsNotImplemented(t *testing.T) {
	l := NewStubListener()
	events := make(chan KeyEvent, 1)
	err := l.Start(events)
	if !errors.Is(err, ErrNotImplemented) {
		t.Errorf("Start() error = %v, want ErrNotImplemented", err)
	}
}

func TestNewStubListener_StopIdempotent(t *testing.T) {
	l := NewStubListener()
	if err := l.Stop(); err != nil {
		t.Errorf("first Stop() error: %v", err)
	}
	if err := l.Stop(); err != nil {
		t.Errorf("second Stop() error: %v", err)
	}
}

func TestNewPlatformListener_Reports(t *testing.T) {
	l := NewPlatformListener()
	info := l.Info()

	// On any platform, the listener should report a valid OS.
	validOS := map[string]bool{"linux": true, "darwin": true, "windows": true}
	if !validOS[info.OS] {
		t.Errorf("Info().OS = %q, expected linux/darwin/windows", info.OS)
	}

	// Method should be non-empty.
	if info.Method == "" {
		t.Error("Info().Method is empty")
	}

	t.Logf("platform listener: %s", info)
}

func TestNewPlatformListener_X11Detected(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("X11 detection only tested on Linux")
	}
	if os.Getenv("DISPLAY") == "" {
		t.Skip("DISPLAY not set; X11 listener cannot be tested")
	}

	l := NewPlatformListener()
	info := l.Info()

	// If the X11 connection succeeds, we expect x11_xgrabkey. If it
	// fails (e.g., no actual X server), the fallback is acceptable.
	switch info.Method {
	case "x11_xgrabkey":
		t.Logf("X11 listener active")
	case "wayland_stub", "stub":
		t.Logf("X11 connection unavailable, fell back to %s", info.Method)
	default:
		t.Errorf("unexpected method %q on Linux with DISPLAY set", info.Method)
	}
}

func TestPlatformInfo_String(t *testing.T) {
	p := PlatformInfo{OS: "linux", Method: "x11_xgrabkey"}
	got := p.String()
	want := "linux/x11_xgrabkey"
	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}
