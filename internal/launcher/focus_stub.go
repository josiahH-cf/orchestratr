//go:build !linux

package launcher

import "fmt"

// FocusWindow is not implemented on non-Linux platforms. It always
// returns ErrFocusNotSupported.
func FocusWindow(pid int) error {
	return fmt.Errorf("%w: not implemented on this platform", ErrFocusNotSupported)
}
