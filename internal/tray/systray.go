//go:build (linux || windows) && !notray

package tray

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"fyne.io/systray"
)

// SystrayProvider implements Provider using fyne.io/systray.
// On Linux it communicates with the compositor via the D-Bus
// StatusNotifierItem protocol (pure Go — no CGo required).
// On Windows it uses the Win32 Shell_NotifyIcon API.
type SystrayProvider struct {
	mu        sync.Mutex
	state     string
	setupDone bool // true after onReady has fired
	ready     chan struct{}
	done      chan struct{}

	// callbacks registered by the daemon before Setup is called.
	onPauseFn     func()
	onResumeFn    func()
	onQuitFn      func()
	onConfigureFn func()

	// injectable for tests.
	runFn  func(onReady, onExit func()) // defaults to systray.Run
	quitFn func()                       // defaults to systray.Quit
}

// Setup initialises the tray icon and context menu. It starts
// systray.Run in a dedicated goroutine and blocks until the tray
// signals it is ready (or the 3-second timeout expires).
func (p *SystrayProvider) Setup() error {
	p.mu.Lock()
	p.ready = make(chan struct{})
	p.done = make(chan struct{})
	runFn := p.runFn
	if runFn == nil {
		runFn = systray.Run
	}
	if p.quitFn == nil {
		p.quitFn = systray.Quit
	}
	p.mu.Unlock()

	go runFn(p.onReady, p.onExit)

	select {
	case <-p.ready:
		return nil
	case <-time.After(3 * time.Second):
		return errTrayTimeout
	}
}

// errTrayTimeout is returned when the systray does not become ready.
var errTrayTimeout = fmt.Errorf("tray did not become ready within 3s")

// onReady is the systray.Run callback. It configures the icon and
// menu, then signals that Setup can return.
func (p *SystrayProvider) onReady() {
	systray.SetTooltip("orchestratr: starting")
	systray.SetIcon(generateIcon())

	miStatus := systray.AddMenuItem("orchestratr", "Status")
	miStatus.Disable()
	systray.AddSeparator()

	miPause := systray.AddMenuItem("Pause", "Pause hotkey capture")
	miResume := systray.AddMenuItem("Resume", "Resume hotkey capture")
	miResume.Disable()
	systray.AddSeparator()

	miConfigure := systray.AddMenuItem("Configure…", "Open configuration")
	systray.AddSeparator()

	miQuit := systray.AddMenuItem("Quit", "Stop the orchestratr daemon")

	// Signal Setup to return.
	p.mu.Lock()
	p.setupDone = true
	close(p.ready)
	p.mu.Unlock()

	// Dispatch menu clicks in a goroutine so systray's event loop
	// is never blocked.
	go func() {
		for {
			select {
			case <-miPause.ClickedCh:
				miPause.Disable()
				miResume.Enable()
				p.mu.Lock()
				fn := p.onPauseFn
				p.mu.Unlock()
				if fn != nil {
					fn()
				}

			case <-miResume.ClickedCh:
				miResume.Disable()
				miPause.Enable()
				p.mu.Lock()
				fn := p.onResumeFn
				p.mu.Unlock()
				if fn != nil {
					fn()
				}

			case <-miConfigure.ClickedCh:
				p.mu.Lock()
				fn := p.onConfigureFn
				p.mu.Unlock()
				if fn != nil {
					fn()
				}

			case <-miQuit.ClickedCh:
				p.mu.Lock()
				fn := p.onQuitFn
				p.mu.Unlock()
				if fn != nil {
					fn()
				}

			case <-p.done:
				return
			}
		}
	}()

	_ = miStatus // suppress unused warning (it's always disabled/display-only)
}

// onExit is the systray.Run exit callback.
func (p *SystrayProvider) onExit() {}

// SetState updates the tray tooltip to reflect the daemon state.
func (p *SystrayProvider) SetState(state string) error {
	p.mu.Lock()
	p.state = state
	active := p.setupDone
	p.mu.Unlock()

	if active {
		systray.SetTooltip("orchestratr: " + state)
	}
	return nil
}

// Quit removes the tray icon and releases resources. Safe to call
// multiple times.
func (p *SystrayProvider) Quit() {
	p.mu.Lock()
	done := p.done
	p.setupDone = false
	p.mu.Unlock()

	if done == nil {
		return
	}

	// Signal the menu dispatch goroutine to exit.
	select {
	case <-done:
	default:
		close(done)
	}

	p.mu.Lock()
	qFn := p.quitFn
	p.mu.Unlock()
	if qFn != nil {
		qFn()
	}
}

// OnPause registers the callback invoked when the user selects Pause.
func (p *SystrayProvider) OnPause(fn func()) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.onPauseFn = fn
}

// OnResume registers the callback invoked when the user selects Resume.
func (p *SystrayProvider) OnResume(fn func()) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.onResumeFn = fn
}

// OnQuit registers the callback invoked when the user selects Quit.
func (p *SystrayProvider) OnQuit(fn func()) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.onQuitFn = fn
}

// OnConfigure registers the callback invoked when the user selects Configure.
func (p *SystrayProvider) OnConfigure(fn func()) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.onConfigureFn = fn
}

// NotifyError displays an error notification. On Linux it uses
// notify-send (best-effort). On Windows it temporarily updates the
// tray tooltip with the error, then restores it after 5 seconds.
func (p *SystrayProvider) NotifyError(title, message string) {
	switch runtime.GOOS {
	case "linux":
		if _, err := exec.LookPath("notify-send"); err != nil {
			return // best-effort: notify-send not installed
		}
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = exec.CommandContext(ctx,
			"notify-send",
			"--urgency=critical",
			"--app-name=orchestratr",
			title,
			message,
		).Run()

	case "windows":
		p.mu.Lock()
		active := p.setupDone
		p.mu.Unlock()
		if !active {
			return
		}
		// Temporarily flash the error in the tooltip, then restore.
		go func() {
			systray.SetTooltip("orchestratr ERROR: " + title + " — " + message)
			time.Sleep(5 * time.Second)
			p.mu.Lock()
			if !p.setupDone {
				p.mu.Unlock()
				return // tray was quit during the sleep
			}
			prev := "orchestratr: " + p.state
			p.mu.Unlock()
			systray.SetTooltip(prev)
		}()
	}
}

// generateIcon creates a minimal 16×16 dark-blue PNG to use as the
// tray icon. It is generated at runtime so no embedded asset file
// is required.
func generateIcon() []byte {
	img := image.NewNRGBA(image.Rect(0, 0, 16, 16))
	iconColor := color.NRGBA{R: 30, G: 90, B: 180, A: 255}
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			img.SetNRGBA(x, y, iconColor)
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}
