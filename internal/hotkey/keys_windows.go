//go:build windows

package hotkey

import "strings"

// Win32 Virtual-Key codes used by RegisterHotKey and WH_KEYBOARD_LL.
// Reference: https://learn.microsoft.com/en-us/windows/win32/inputdev/virtual-key-codes
const (
	_VK_BACK   uint32 = 0x08
	_VK_TAB    uint32 = 0x09
	_VK_RETURN uint32 = 0x0D
	_VK_ESCAPE uint32 = 0x1B
	_VK_SPACE  uint32 = 0x20
	_VK_PRIOR  uint32 = 0x21 // Page Up
	_VK_NEXT   uint32 = 0x22 // Page Down
	_VK_END    uint32 = 0x23
	_VK_HOME   uint32 = 0x24
	_VK_LEFT   uint32 = 0x25
	_VK_UP     uint32 = 0x26
	_VK_RIGHT  uint32 = 0x27
	_VK_DOWN   uint32 = 0x28
	_VK_INSERT uint32 = 0x2D
	_VK_DELETE uint32 = 0x2E
	_VK_F1     uint32 = 0x70
	_VK_F2     uint32 = 0x71
	_VK_F3     uint32 = 0x72
	_VK_F4     uint32 = 0x73
	_VK_F5     uint32 = 0x74
	_VK_F6     uint32 = 0x75
	_VK_F7     uint32 = 0x76
	_VK_F8     uint32 = 0x77
	_VK_F9     uint32 = 0x78
	_VK_F10    uint32 = 0x79
	_VK_F11    uint32 = 0x7A
	_VK_F12    uint32 = 0x7B
)

// Win32 modifier flags for RegisterHotKey.
const (
	_MOD_ALT     uint32 = 0x0001
	_MOD_CONTROL uint32 = 0x0002
	_MOD_SHIFT   uint32 = 0x0004
	_MOD_WIN     uint32 = 0x0008
)

// Win32 message constants.
const (
	_WM_HOTKEY      uint32 = 0x0312
	_WM_QUIT        uint32 = 0x0012
	_WM_DESTROY     uint32 = 0x0002
	_WH_KEYBOARD_LL        = 13
	_WM_KEYDOWN     uint32 = 0x0100
	_WM_SYSKEYDOWN  uint32 = 0x0104
)

// keyToVK converts an internal key code string to a Win32 Virtual-Key code.
func keyToVK(code string) (uint32, bool) {
	code = strings.ToLower(code)

	switch code {
	case "space":
		return _VK_SPACE, true
	case "return", "enter":
		return _VK_RETURN, true
	case "escape", "esc":
		return _VK_ESCAPE, true
	case "tab":
		return _VK_TAB, true
	case "backspace":
		return _VK_BACK, true
	case "delete", "del":
		return _VK_DELETE, true
	case "insert":
		return _VK_INSERT, true
	case "home":
		return _VK_HOME, true
	case "end":
		return _VK_END, true
	case "pageup", "prior":
		return _VK_PRIOR, true
	case "pagedown", "next":
		return _VK_NEXT, true
	case "up":
		return _VK_UP, true
	case "down":
		return _VK_DOWN, true
	case "left":
		return _VK_LEFT, true
	case "right":
		return _VK_RIGHT, true
	case "f1":
		return _VK_F1, true
	case "f2":
		return _VK_F2, true
	case "f3":
		return _VK_F3, true
	case "f4":
		return _VK_F4, true
	case "f5":
		return _VK_F5, true
	case "f6":
		return _VK_F6, true
	case "f7":
		return _VK_F7, true
	case "f8":
		return _VK_F8, true
	case "f9":
		return _VK_F9, true
	case "f10":
		return _VK_F10, true
	case "f11":
		return _VK_F11, true
	case "f12":
		return _VK_F12, true
	}

	// Single ASCII letter a-z → VK code is uppercase ASCII.
	if len(code) == 1 {
		c := code[0]
		if c >= 'a' && c <= 'z' {
			return uint32(c - 'a' + 'A'), true
		}
		if c >= '0' && c <= '9' {
			return uint32(c), true
		}
	}

	return 0, false
}

// vkToKeyCode converts a Win32 Virtual-Key code back to an internal
// key code string. Returns empty string if the VK code is unknown.
func vkToKeyCode(vk uint32) string {
	switch vk {
	case _VK_SPACE:
		return "space"
	case _VK_RETURN:
		return "return"
	case _VK_ESCAPE:
		return "escape"
	case _VK_TAB:
		return "tab"
	case _VK_BACK:
		return "backspace"
	case _VK_DELETE:
		return "delete"
	case _VK_INSERT:
		return "insert"
	case _VK_HOME:
		return "home"
	case _VK_END:
		return "end"
	case _VK_PRIOR:
		return "pageup"
	case _VK_NEXT:
		return "pagedown"
	case _VK_UP:
		return "up"
	case _VK_DOWN:
		return "down"
	case _VK_LEFT:
		return "left"
	case _VK_RIGHT:
		return "right"
	case _VK_F1:
		return "f1"
	case _VK_F2:
		return "f2"
	case _VK_F3:
		return "f3"
	case _VK_F4:
		return "f4"
	case _VK_F5:
		return "f5"
	case _VK_F6:
		return "f6"
	case _VK_F7:
		return "f7"
	case _VK_F8:
		return "f8"
	case _VK_F9:
		return "f9"
	case _VK_F10:
		return "f10"
	case _VK_F11:
		return "f11"
	case _VK_F12:
		return "f12"
	}

	// Single ASCII letter A-Z.
	if vk >= 0x41 && vk <= 0x5A {
		return string(rune(vk - 0x41 + 'a'))
	}
	// Digits 0-9.
	if vk >= 0x30 && vk <= 0x39 {
		return string(rune(vk))
	}

	return ""
}

// modifiersToWin32 converts Modifier flags to Win32 RegisterHotKey
// modifier flags.
func modifiersToWin32(m Modifier) uint32 {
	var flags uint32
	if m&ModCtrl != 0 {
		flags |= _MOD_CONTROL
	}
	if m&ModAlt != 0 {
		flags |= _MOD_ALT
	}
	if m&ModShift != 0 {
		flags |= _MOD_SHIFT
	}
	if m&ModSuper != 0 {
		flags |= _MOD_WIN
	}
	return flags
}
