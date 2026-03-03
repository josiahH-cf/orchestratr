//go:build windows

package hotkey

import (
	"fmt"
	"runtime"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Win32 API constants not defined in golang.org/x/sys/windows.
const (
	_HWND_MESSAGE = ^uintptr(2) // (HWND)-3 — message-only window parent
	_HOTKEY_ID    = 1           // ID for our single RegisterHotKey call

	// Custom window messages for cross-thread communication between
	// the engine goroutine and the Win32 message pump thread.
	_WM_APP_GRAB    = 0x8000 + 1 // request keyboard hook install
	_WM_APP_UNGRAB  = 0x8000 + 2 // request keyboard hook removal
	_WM_APP_HOOKEVT = 0x8000 + 3 // wake pump to drain hook events
)

var (
	user32                  = windows.NewLazySystemDLL("user32.dll")
	procCreateWindowExW     = user32.NewProc("CreateWindowExW")
	procDestroyWindow       = user32.NewProc("DestroyWindow")
	procRegisterHotKey      = user32.NewProc("RegisterHotKey")
	procUnregisterHotKey    = user32.NewProc("UnregisterHotKey")
	procGetMessageW         = user32.NewProc("GetMessageW")
	procPostThreadMessageW  = user32.NewProc("PostThreadMessageW")
	procPostMessageW        = user32.NewProc("PostMessageW")
	procSetWindowsHookExW   = user32.NewProc("SetWindowsHookExW")
	procUnhookWindowsHookEx = user32.NewProc("UnhookWindowsHookEx")
	procCallNextHookEx      = user32.NewProc("CallNextHookEx")
	procRegisterClassExW    = user32.NewProc("RegisterClassExW")
	procDefWindowProcW      = user32.NewProc("DefWindowProcW")

	kernel32               = windows.NewLazySystemDLL("kernel32.dll")
	procGetCurrentThreadId = kernel32.NewProc("GetCurrentThreadId")
)

// _MSG is the Win32 MSG structure for GetMessage/PeekMessage.
type _MSG struct {
	HWnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      struct{ X, Y int32 }
}

// _KBDLLHOOKSTRUCT is the data passed to the WH_KEYBOARD_LL callback.
type _KBDLLHOOKSTRUCT struct {
	VKCode      uint32
	ScanCode    uint32
	Flags       uint32
	Time        uint32
	DwExtraInfo uintptr
}

// _WNDCLASSEXW for RegisterClassExW.
type _WNDCLASSEXW struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     uintptr
	HIcon         uintptr
	HCursor       uintptr
	HbrBackground uintptr
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       uintptr
}

// windowsListener implements the Listener interface on Windows using
// RegisterHotKey for the leader key and WH_KEYBOARD_LL for chord
// capture during the chord-wait window.
type windowsListener struct {
	mu        sync.Mutex
	leader    Key
	leaderVK  uint32
	leaderMod uint32

	hwnd     uintptr         // hidden message-only window
	threadID uint32          // message pump thread ID
	hookH    uintptr         // WH_KEYBOARD_LL hook handle
	events   chan<- KeyEvent // output channel
	stopCh   chan struct{}
	stopped  bool

	// hookEvents receives key events captured by the low-level hook
	// callback. Buffered so the hook proc returns quickly.
	hookEvents chan KeyEvent

	// grabDone and ungrabDone synchronize GrabKeyboard /
	// UngrabKeyboard with the message pump thread, which is the
	// only thread allowed to install / remove the LL hook.
	grabDone   chan error
	ungrabDone chan struct{}
}

// newWindowsListener creates a new windowsListener. The listener does
// not acquire any system resources until Register/Start are called.
func newWindowsListener() *windowsListener {
	return &windowsListener{
		stopCh:     make(chan struct{}),
		hookEvents: make(chan KeyEvent, 64),
		grabDone:   make(chan error, 1),
		ungrabDone: make(chan struct{}, 1),
	}
}

// NewPlatformListener returns the Windows hotkey listener.
func NewPlatformListener() Listener {
	return newWindowsListener()
}

// Info returns platform diagnostics.
func (l *windowsListener) Info() PlatformInfo {
	return PlatformInfo{OS: "windows", Method: "registerhotkey"}
}

// Register stores the leader key and validates that it can be mapped
// to a Win32 Virtual-Key code. The actual RegisterHotKey call happens
// in Start when the message pump thread is running.
func (l *windowsListener) Register(leader Key) (string, error) {
	vk, ok := keyToVK(leader.Code)
	if !ok {
		return "", fmt.Errorf("unknown key %q for Windows", leader.Code)
	}

	l.mu.Lock()
	l.leader = leader
	l.leaderVK = vk
	l.leaderMod = modifiersToWin32(leader.Modifiers)
	l.mu.Unlock()

	return CheckConflicts(leader), nil
}

// Start creates a hidden message-only window, registers the leader
// key hotkey, and runs the Win32 message pump. It blocks until Stop
// is called or a fatal error occurs. Key events are sent to the
// provided channel.
func (l *windowsListener) Start(events chan<- KeyEvent) error {
	l.mu.Lock()
	l.events = events
	l.mu.Unlock()

	// The message pump must run on a dedicated OS thread because
	// Win32 hooks and message queues are thread-local.
	errCh := make(chan error, 1)

	go func() {
		// Lock this goroutine to its OS thread for the lifetime of
		// the message pump — required by RegisterHotKey and
		// SetWindowsHookEx.
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		tid, _, _ := procGetCurrentThreadId.Call()
		l.mu.Lock()
		l.threadID = uint32(tid)
		l.mu.Unlock()

		hwnd, err := l.createMessageWindow()
		if err != nil {
			errCh <- fmt.Errorf("creating message window: %w", err)
			return
		}
		l.mu.Lock()
		l.hwnd = hwnd
		l.mu.Unlock()

		// Register the leader key hotkey.
		ret, _, callErr := procRegisterHotKey.Call(
			hwnd,
			_HOTKEY_ID,
			uintptr(l.leaderMod),
			uintptr(l.leaderVK),
		)
		if ret == 0 {
			procDestroyWindow.Call(hwnd)
			errCh <- fmt.Errorf("RegisterHotKey failed: %v", callErr)
			return
		}

		// Signal that startup succeeded.
		errCh <- nil

		// Run the message pump.
		l.messagePump()
	}()

	// Wait for the message pump to start or fail.
	if err := <-errCh; err != nil {
		return err
	}

	// Block until stopCh is closed (same contract as other listeners).
	<-l.stopCh
	return nil
}

// messagePump runs the Win32 GetMessage loop. It processes WM_HOTKEY
// messages, handles cross-thread hook management requests, and
// forwards hook-captured events. It exits when WM_QUIT is received.
// Deferred cleanup releases the hotkey, hook, and window on the
// correct OS thread.
func (l *windowsListener) messagePump() {
	defer l.cleanup()

	var msg _MSG
	for {
		// GetMessage blocks until a message is available. Returns 0
		// for WM_QUIT, -1 on error.
		ret, _, _ := procGetMessageW.Call(
			uintptr(unsafe.Pointer(&msg)),
			0, // all windows
			0,
			0,
		)
		if int32(ret) <= 0 {
			// WM_QUIT or error — exit the pump.
			return
		}

		switch msg.Message {
		case _WM_HOTKEY:
			// Leader key pressed — send it to the engine.
			l.mu.Lock()
			ch := l.events
			l.mu.Unlock()
			if ch != nil {
				ch <- KeyEvent{Key: l.leader, Pressed: true}
			}

		case _WM_APP_GRAB:
			// Engine requested keyboard hook installation.
			l.grabDone <- l.installHook()

		case _WM_APP_UNGRAB:
			// Engine requested keyboard hook removal.
			l.uninstallHook()
			l.ungrabDone <- struct{}{}

		case _WM_APP_HOOKEVT:
			// Wake message from hook callback — just need to reach
			// drainHookEvents below.
		}

		// Drain any hook events accumulated since the last pump
		// iteration.
		l.drainHookEvents()
	}
}

// drainHookEvents forwards any pending events from the low-level
// keyboard hook to the engine's event channel.
func (l *windowsListener) drainHookEvents() {
	l.mu.Lock()
	ch := l.events
	l.mu.Unlock()
	if ch == nil {
		return
	}

	for {
		select {
		case evt := <-l.hookEvents:
			ch <- evt
		default:
			return
		}
	}
}

// GrabKeyboard requests the message pump thread to install a
// WH_KEYBOARD_LL hook. The hook must be installed on the pump thread
// because Windows delivers LL hook callbacks on the installing thread,
// and that thread must have an active message pump.
func (l *windowsListener) GrabKeyboard() error {
	l.mu.Lock()
	tid := l.threadID
	l.mu.Unlock()

	if tid == 0 {
		return fmt.Errorf("message pump not running")
	}

	// Drain any stale response from a previous call.
	select {
	case <-l.grabDone:
	default:
	}

	ret, _, callErr := procPostThreadMessageW.Call(
		uintptr(tid), uintptr(_WM_APP_GRAB), 0, 0,
	)
	if ret == 0 {
		return fmt.Errorf("PostThreadMessage(WM_APP_GRAB) failed: %v", callErr)
	}

	select {
	case grabErr := <-l.grabDone:
		return grabErr
	case <-l.stopCh:
		return fmt.Errorf("listener stopped")
	}
}

// installHook installs the WH_KEYBOARD_LL hook. Must be called on the
// message pump thread.
func (l *windowsListener) installHook() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.hookH != 0 {
		return nil // already hooked
	}

	h, _, err := procSetWindowsHookExW.Call(
		_WH_KEYBOARD_LL,
		windows.NewCallback(l.llKeyboardProc),
		0, // hMod — 0 for LL hooks (global)
		0, // dwThreadId — 0 for global hook
	)
	if h == 0 {
		return fmt.Errorf("SetWindowsHookEx(WH_KEYBOARD_LL) failed: %v", err)
	}

	l.hookH = h
	return nil
}

// UngrabKeyboard requests the message pump thread to remove the
// WH_KEYBOARD_LL hook. Blocks until the hook is removed or the
// listener is stopped.
func (l *windowsListener) UngrabKeyboard() {
	l.mu.Lock()
	tid := l.threadID
	l.mu.Unlock()

	if tid == 0 {
		return
	}

	// Drain any stale response.
	select {
	case <-l.ungrabDone:
	default:
	}

	procPostThreadMessageW.Call(
		uintptr(tid), uintptr(_WM_APP_UNGRAB), 0, 0,
	)

	select {
	case <-l.ungrabDone:
	case <-l.stopCh:
	}
}

// uninstallHook removes the WH_KEYBOARD_LL hook. Must be called on
// the message pump thread.
func (l *windowsListener) uninstallHook() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.hookH != 0 {
		procUnhookWindowsHookEx.Call(l.hookH)
		l.hookH = 0
	}
}

// Stop posts WM_QUIT to the message pump thread and signals Start to
// return. The pump thread's deferred cleanup releases the hotkey,
// hook, and window on the correct OS thread. Safe to call multiple
// times.
func (l *windowsListener) Stop() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.stopped {
		return nil
	}
	l.stopped = true

	// Post WM_QUIT to the message pump thread to exit GetMessage.
	// The pump's deferred cleanup() handles resource release on the
	// correct thread.
	if l.threadID != 0 {
		procPostThreadMessageW.Call(
			uintptr(l.threadID),
			uintptr(_WM_QUIT),
			0,
			0,
		)
	}

	// Signal the blocking Start() call to return.
	select {
	case <-l.stopCh:
	default:
		close(l.stopCh)
	}

	return nil
}

// cleanup releases Win32 resources. It must be called on the message
// pump thread (via defer in messagePump) so that DestroyWindow is
// invoked from the same thread that created the window.
func (l *windowsListener) cleanup() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.hookH != 0 {
		procUnhookWindowsHookEx.Call(l.hookH)
		l.hookH = 0
	}
	if l.hwnd != 0 {
		procUnregisterHotKey.Call(l.hwnd, _HOTKEY_ID)
		procDestroyWindow.Call(l.hwnd)
		l.hwnd = 0
	}
}

// llKeyboardProc is the WH_KEYBOARD_LL callback. It captures keydown
// events and sends them to the hookEvents channel, then posts a wake
// message so the pump drains them promptly. The callback must return
// quickly to avoid the OS removing the hook.
//
// lParam is declared as unsafe.Pointer (rather than uintptr) so that
// the conversion does not trigger a go-vet unsafeptr warning. The
// Win32 API guarantees lParam is a valid pointer to KBDLLHOOKSTRUCT
// for the callback's lifetime.
func (l *windowsListener) llKeyboardProc(nCode int, wParam uintptr, lParam unsafe.Pointer) uintptr {
	if nCode >= 0 && (uint32(wParam) == _WM_KEYDOWN || uint32(wParam) == _WM_SYSKEYDOWN) {
		kb := (*_KBDLLHOOKSTRUCT)(lParam)
		code := vkToKeyCode(kb.VKCode)
		if code != "" {
			evt := KeyEvent{
				Key:     Key{Code: code},
				Pressed: true,
			}
			// Non-blocking send — drop if buffer full to avoid
			// stalling the hook.
			select {
			case l.hookEvents <- evt:
			default:
			}
		}
	}

	// Post a wake message so the pump's drainHookEvents runs
	// promptly after this callback returns. Without this, the pump
	// would block in GetMessage until an unrelated message arrives,
	// delaying chord delivery to the engine.
	procPostMessageW.Call(l.hwnd, uintptr(_WM_APP_HOOKEVT), 0, 0)

	// Always call the next hook in the chain.
	ret, _, _ := procCallNextHookEx.Call(0, uintptr(nCode), wParam, uintptr(lParam))
	return ret
}

// createMessageWindow registers a window class and creates a hidden
// message-only window for receiving WM_HOTKEY messages.
func (l *windowsListener) createMessageWindow() (uintptr, error) {
	className, err := windows.UTF16PtrFromString("orchestratr_hotkey")
	if err != nil {
		return 0, err
	}

	wc := _WNDCLASSEXW{
		CbSize:        uint32(unsafe.Sizeof(_WNDCLASSEXW{})),
		LpfnWndProc:   windows.NewCallback(defWindowProc),
		LpszClassName: className,
	}

	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	// RegisterClassEx may fail if already registered — that's OK.

	hwnd, _, createErr := procCreateWindowExW.Call(
		0,                                  // dwExStyle
		uintptr(unsafe.Pointer(className)), // lpClassName
		0,                                  // lpWindowName
		0,                                  // dwStyle
		0, 0, 0, 0,                         // x, y, w, h
		_HWND_MESSAGE, // hWndParent (message-only)
		0,             // hMenu
		0,             // hInstance
		0,             // lpParam
	)
	if hwnd == 0 {
		return 0, fmt.Errorf("CreateWindowEx failed: %v", createErr)
	}
	return hwnd, nil
}

// defWindowProc is the default window procedure for the message-only window.
func defWindowProc(hwnd uintptr, msg uint32, wParam uintptr, lParam uintptr) uintptr {
	ret, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
	return ret
}
