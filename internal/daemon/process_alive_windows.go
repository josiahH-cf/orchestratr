//go:build windows

package daemon

import "golang.org/x/sys/windows"

// processAlive checks process liveness on Windows via OpenProcess.
func processAlive(pid int) bool {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return false
	}
	_ = windows.CloseHandle(h)
	return true
}
