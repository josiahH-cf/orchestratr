// Package platform provides platform detection utilities for
// orchestratr, including WSL2 detection and macOS accessibility
// permission checking.
package platform

import "runtime"

// Name returns the current platform name (e.g. "linux", "darwin", "windows").
func Name() string {
	return runtime.GOOS
}
