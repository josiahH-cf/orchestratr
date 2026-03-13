//go:build !windows

package daemon

import "syscall"

// processAlive checks process liveness on Unix-like systems via signal 0.
func processAlive(pid int) bool {
	err := syscall.Kill(pid, 0)
	if err == nil {
		return true
	}
	// EPERM means the process exists but we lack permissions.
	return err == syscall.EPERM
}
