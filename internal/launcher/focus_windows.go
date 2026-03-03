//go:build windows

package launcher

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32DLL                    = windows.NewLazySystemDLL("user32.dll")
	procEnumWindows              = user32DLL.NewProc("EnumWindows")
	procGetWindowThreadProcessId = user32DLL.NewProc("GetWindowThreadProcessId")
	procSetForegroundWindow      = user32DLL.NewProc("SetForegroundWindow")
	procIsIconic                 = user32DLL.NewProc("IsIconic")
	procShowWindow               = user32DLL.NewProc("ShowWindow")
	procGetForegroundWindow      = user32DLL.NewProc("GetForegroundWindow")
	procGetWindowThreadProcId2   = user32DLL.NewProc("GetWindowThreadProcessId")
	procAttachThreadInput        = user32DLL.NewProc("AttachThreadInput")
	procIsWindowVisible          = user32DLL.NewProc("IsWindowVisible")

	kernel32DLL                  = windows.NewLazySystemDLL("kernel32.dll")
	procGetCurrentThreadIdFocus  = kernel32DLL.NewProc("GetCurrentThreadId")
)

const (
	_SW_RESTORE = 9
)

// FocusWindow brings the window belonging to the given PID to the
// foreground on Windows. It uses EnumWindows to find the target
// window by PID, then SetForegroundWindow with the AttachThreadInput
// workaround to bypass the foreground lock.
func FocusWindow(pid int) error {
	if pid <= 0 {
		return fmt.Errorf("invalid PID %d", pid)
	}

	hwnd, err := findWindowByPID(uint32(pid))
	if err != nil {
		return err
	}

	// If the window is minimized, restore it first.
	iconic, _, _ := procIsIconic.Call(hwnd)
	if iconic != 0 {
		procShowWindow.Call(hwnd, _SW_RESTORE)
	}

	// Try SetForegroundWindow directly first.
	ret, _, _ := procSetForegroundWindow.Call(hwnd)
	if ret != 0 {
		return nil // success
	}

	// Foreground lock workaround: attach our thread's input to the
	// current foreground window's thread, call SetForegroundWindow,
	// then detach.
	return focusWithAttach(hwnd)
}

// findWindowByPID enumerates all top-level windows and returns the
// first visible window handle owned by the given process ID.
func findWindowByPID(targetPID uint32) (uintptr, error) {
	type result struct {
		hwnd uintptr
	}
	var found result

	cb := windows.NewCallback(func(hwnd uintptr, lParam uintptr) uintptr {
		// Check if window is visible.
		visible, _, _ := procIsWindowVisible.Call(hwnd)
		if visible == 0 {
			return 1 // continue enumeration
		}

		var windowPID uint32
		procGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&windowPID)))
		if windowPID == targetPID {
			// Store result via the lParam pointer.
			r := (*result)(unsafe.Pointer(lParam))
			r.hwnd = hwnd
			return 0 // stop enumeration
		}
		return 1 // continue
	})

	procEnumWindows.Call(cb, uintptr(unsafe.Pointer(&found)))

	if found.hwnd == 0 {
		return 0, fmt.Errorf("%w: no visible window found for PID %d", ErrFocusNotSupported, targetPID)
	}

	return found.hwnd, nil
}

// focusWithAttach uses the AttachThreadInput technique to steal the
// foreground lock and set the target window as foreground.
func focusWithAttach(hwnd uintptr) error {
	// Get the foreground window's thread.
	fgHwnd, _, _ := procGetForegroundWindow.Call()
	if fgHwnd == 0 {
		// No foreground window — SetForegroundWindow should have
		// worked. Return a generic error.
		return fmt.Errorf("SetForegroundWindow failed and no foreground window exists")
	}

	var fgPID uint32
	fgThread, _, _ := procGetWindowThreadProcId2.Call(fgHwnd, uintptr(unsafe.Pointer(&fgPID)))

	// Get our thread ID.
	ourThread, _, _ := procGetCurrentThreadIdFocus.Call()

	if fgThread != ourThread {
		// Attach our input to the foreground thread.
		procAttachThreadInput.Call(ourThread, fgThread, 1) // TRUE
		defer procAttachThreadInput.Call(ourThread, fgThread, 0) // FALSE
	}

	ret, _, _ := procSetForegroundWindow.Call(hwnd)
	if ret == 0 {
		return fmt.Errorf("SetForegroundWindow failed after AttachThreadInput")
	}

	return nil
}
