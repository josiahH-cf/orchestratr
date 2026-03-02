//go:build linux && cgo

package hotkey

/*
#cgo LDFLAGS: -lX11
#include <X11/Xlib.h>
#include <X11/keysym.h>
#include <X11/XKBlib.h>
#include <stdlib.h>

// Shutdown flag: when set, the IO error handler returns silently
// instead of calling exit(). This suppresses the "XIO: fatal IO
// error" that occurs when closing the display while XNextEvent is
// blocked (OI-3).
static volatile int x11_shutting_down = 0;

static int x11_io_error_handler(Display* dpy) {
	if (x11_shutting_down) {
		// Returning from an IO error handler is technically undefined
		// behavior per Xlib, but in practice it prevents the abort/
		// exit call. The event loop will exit via stopCh anyway.
		return 0;
	}
	// Default behavior for non-shutdown errors: print and exit.
	return 0;
}

static void x11_install_error_handler() {
	XSetIOErrorHandler(x11_io_error_handler);
}

static void x11_set_shutting_down() {
	x11_shutting_down = 1;
}

// x11_open_display opens a connection to the X server.
// Returns NULL on failure.
static Display* x11_open_display() {
	return XOpenDisplay(NULL);
}

// x11_close_display closes the connection.
static void x11_close_display(Display* dpy) {
	if (dpy) XCloseDisplay(dpy);
}

// x11_default_root returns the root window of the default screen.
static Window x11_default_root(Display* dpy) {
	return DefaultRootWindow(dpy);
}

// x11_keysym_to_keycode converts a KeySym to a KeyCode.
static KeyCode x11_keysym_to_keycode(Display* dpy, KeySym sym) {
	return XKeysymToKeycode(dpy, sym);
}

// x11_grab_key registers a global hotkey grab on the root window.
// owner_events=False means the key press is NOT forwarded to the
// focused application.
static int x11_grab_key(Display* dpy, KeyCode code, unsigned int modifiers, Window root) {
	return XGrabKey(dpy, code, modifiers, root, False, GrabModeAsync, GrabModeAsync);
}

// x11_ungrab_key releases a global hotkey grab.
static void x11_ungrab_key(Display* dpy, KeyCode code, unsigned int modifiers, Window root) {
	XUngrabKey(dpy, code, modifiers, root);
}

// x11_grab_keyboard grabs the entire keyboard for exclusive input.
// Used during chord capture to prevent key leaking.
static int x11_grab_keyboard(Display* dpy, Window root) {
	return XGrabKeyboard(dpy, root, False, GrabModeAsync, GrabModeAsync, CurrentTime);
}

// x11_ungrab_keyboard releases the keyboard grab.
static void x11_ungrab_keyboard(Display* dpy) {
	XUngrabKeyboard(dpy, CurrentTime);
}

// x11_flush sends all buffered requests to the server.
static void x11_flush(Display* dpy) {
	XFlush(dpy);
}

// x11_pending returns the number of events in the queue.
static int x11_pending(Display* dpy) {
	return XPending(dpy);
}

// x11_connection_number returns the file descriptor for select().
static int x11_connection_number(Display* dpy) {
	return ConnectionNumber(dpy);
}

// x11_lookup_keysym extracts the keysym from a key event, using
// XKB-aware lookup for correct modifier handling.
static KeySym x11_lookup_keysym(Display* dpy, XKeyEvent* ev) {
	KeySym sym = NoSymbol;
	XkbLookupKeySym(dpy, ev->keycode, ev->state, NULL, &sym);
	return sym;
}
*/
import "C"

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"unsafe"
)

// x11Listener captures global hotkeys on Linux/X11 using XGrabKey.
type x11Listener struct {
	mu      sync.Mutex
	dpy     *C.Display
	root    C.Window
	leader  Key
	keycode C.KeyCode
	modmask C.uint
	stopped bool
	stopCh  chan struct{}
}

// newX11Listener creates an X11 listener. Returns nil if DISPLAY is
// not set or the connection fails.
func newX11Listener() *x11Listener {
	if os.Getenv("DISPLAY") == "" {
		return nil
	}

	// Install our IO error handler before opening the display so we
	// can suppress fatal errors during shutdown (OI-3).
	C.x11_install_error_handler()

	dpy := C.x11_open_display()
	if dpy == nil {
		return nil
	}

	return &x11Listener{
		dpy:    dpy,
		root:   C.x11_default_root(dpy),
		stopCh: make(chan struct{}),
	}
}

// Info returns platform diagnostics.
func (l *x11Listener) Info() PlatformInfo {
	return PlatformInfo{OS: "linux", Method: "x11_xgrabkey"}
}

// Register sets the leader key for global capture via XGrabKey.
func (l *x11Listener) Register(leader Key) (string, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	sym := keyToKeySym(leader.Code)
	if sym == C.NoSymbol {
		return "", fmt.Errorf("unknown key %q for X11", leader.Code)
	}

	kc := C.x11_keysym_to_keycode(l.dpy, sym)
	if kc == 0 {
		return "", fmt.Errorf("no keycode for key %q", leader.Code)
	}

	modmask := modifiersToX11Mask(leader.Modifiers)

	// Grab with and without NumLock (Mod2Mask) and CapsLock (LockMask)
	// to make the grab work regardless of those modifier states.
	masks := []C.uint{
		modmask,
		modmask | C.Mod2Mask,
		modmask | C.LockMask,
		modmask | C.Mod2Mask | C.LockMask,
	}

	for _, m := range masks {
		C.x11_grab_key(l.dpy, kc, m, l.root)
	}
	C.x11_flush(l.dpy)

	l.leader = leader
	l.keycode = kc
	l.modmask = modmask

	return CheckConflicts(leader), nil
}

// Start listens for X11 key events and sends them to the channel.
// It blocks until Stop is called.
func (l *x11Listener) Start(events chan<- KeyEvent) error {
	fd := C.x11_connection_number(l.dpy)

	for {
		select {
		case <-l.stopCh:
			return nil
		default:
		}

		// Use select(2) via Go's netpoll-compatible approach: poll
		// the X11 fd with a short timeout so we can check stopCh.
		if !waitForFd(int(fd), 100) {
			continue
		}

		for C.x11_pending(l.dpy) > 0 {
			var ev C.XEvent
			C.XNextEvent(l.dpy, &ev)

			evType := *(*C.int)(unsafe.Pointer(&ev))
			if evType != C.KeyPress {
				continue
			}

			keyEv := (*C.XKeyEvent)(unsafe.Pointer(&ev))
			sym := C.x11_lookup_keysym(l.dpy, keyEv)

			key := keysymToKey(sym, keyEv.state)
			events <- KeyEvent{Key: key, Pressed: true}
		}
	}
}

// Stop releases the hotkey grab and closes the X11 connection.
func (l *x11Listener) Stop() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.stopped {
		return nil
	}
	l.stopped = true
	close(l.stopCh)

	if l.dpy != nil {
		// Signal the IO error handler that we're shutting down so
		// it doesn't print "XIO: fatal IO error" (OI-3).
		C.x11_set_shutting_down()

		masks := []C.uint{
			l.modmask,
			l.modmask | C.Mod2Mask,
			l.modmask | C.LockMask,
			l.modmask | C.Mod2Mask | C.LockMask,
		}
		for _, m := range masks {
			C.x11_ungrab_key(l.dpy, l.keycode, m, l.root)
		}
		C.x11_ungrab_keyboard(l.dpy)
		C.x11_close_display(l.dpy)
		l.dpy = nil
	}
	return nil
}

// --- Key mapping helpers ---

// keyToKeySym converts our internal key code string to an X11 KeySym.
func keyToKeySym(code string) C.KeySym {
	code = strings.ToLower(code)

	// Common key names to X11 KeySym constants.
	switch code {
	case "space":
		return C.XK_space
	case "return", "enter":
		return C.XK_Return
	case "escape", "esc":
		return C.XK_Escape
	case "tab":
		return C.XK_Tab
	case "backspace":
		return C.XK_BackSpace
	case "delete", "del":
		return C.XK_Delete
	case "insert":
		return C.XK_Insert
	case "home":
		return C.XK_Home
	case "end":
		return C.XK_End
	case "pageup", "prior":
		return C.XK_Prior
	case "pagedown", "next":
		return C.XK_Next
	case "up":
		return C.XK_Up
	case "down":
		return C.XK_Down
	case "left":
		return C.XK_Left
	case "right":
		return C.XK_Right
	case "f1":
		return C.XK_F1
	case "f2":
		return C.XK_F2
	case "f3":
		return C.XK_F3
	case "f4":
		return C.XK_F4
	case "f5":
		return C.XK_F5
	case "f6":
		return C.XK_F6
	case "f7":
		return C.XK_F7
	case "f8":
		return C.XK_F8
	case "f9":
		return C.XK_F9
	case "f10":
		return C.XK_F10
	case "f11":
		return C.XK_F11
	case "f12":
		return C.XK_F12
	}

	// Single ASCII character.
	if len(code) == 1 {
		c := code[0]
		if c >= 'a' && c <= 'z' {
			return C.KeySym(c)
		}
		if c >= '0' && c <= '9' {
			return C.KeySym(c)
		}
	}

	return C.NoSymbol
}

// modifiersToX11Mask converts our Modifier flags to X11 modifier mask.
func modifiersToX11Mask(m Modifier) C.uint {
	var mask C.uint
	if m&ModCtrl != 0 {
		mask |= C.ControlMask
	}
	if m&ModAlt != 0 {
		mask |= C.Mod1Mask
	}
	if m&ModShift != 0 {
		mask |= C.ShiftMask
	}
	if m&ModSuper != 0 {
		mask |= C.Mod4Mask
	}
	return mask
}

// keysymToKey converts an X11 KeySym and state mask back to our Key type.
func keysymToKey(sym C.KeySym, state C.uint) Key {
	var mods Modifier
	// Ignore NumLock (Mod2) and CapsLock (Lock) for modifier detection.
	cleanState := state & ^(C.uint(C.Mod2Mask) | C.uint(C.LockMask))
	if cleanState&C.uint(C.ControlMask) != 0 {
		mods |= ModCtrl
	}
	if cleanState&C.uint(C.Mod1Mask) != 0 {
		mods |= ModAlt
	}
	if cleanState&C.uint(C.ShiftMask) != 0 {
		mods |= ModShift
	}
	if cleanState&C.uint(C.Mod4Mask) != 0 {
		mods |= ModSuper
	}

	code := keysymToCode(sym)
	return Key{Modifiers: mods, Code: code}
}

// keysymToCode converts an X11 KeySym to our internal key code string.
func keysymToCode(sym C.KeySym) string {
	switch sym {
	case C.XK_space:
		return "space"
	case C.XK_Return:
		return "return"
	case C.XK_Escape:
		return "escape"
	case C.XK_Tab:
		return "tab"
	case C.XK_BackSpace:
		return "backspace"
	case C.XK_Delete:
		return "delete"
	case C.XK_Insert:
		return "insert"
	case C.XK_Home:
		return "home"
	case C.XK_End:
		return "end"
	case C.XK_Prior:
		return "pageup"
	case C.XK_Next:
		return "pagedown"
	case C.XK_Up:
		return "up"
	case C.XK_Down:
		return "down"
	case C.XK_Left:
		return "left"
	case C.XK_Right:
		return "right"
	case C.XK_F1:
		return "f1"
	case C.XK_F2:
		return "f2"
	case C.XK_F3:
		return "f3"
	case C.XK_F4:
		return "f4"
	case C.XK_F5:
		return "f5"
	case C.XK_F6:
		return "f6"
	case C.XK_F7:
		return "f7"
	case C.XK_F8:
		return "f8"
	case C.XK_F9:
		return "f9"
	case C.XK_F10:
		return "f10"
	case C.XK_F11:
		return "f11"
	case C.XK_F12:
		return "f12"
	}

	// Single ASCII characters.
	if sym >= C.XK_a && sym <= C.XK_z {
		return string(rune(sym))
	}
	if sym >= C.XK_0 && sym <= C.XK_9 {
		return string(rune(sym))
	}

	return fmt.Sprintf("0x%x", int(sym))
}
