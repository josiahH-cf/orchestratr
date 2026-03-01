//go:build linux && cgo

package hotkey

import (
	"golang.org/x/sys/unix"
)

// waitForFd blocks until the file descriptor is readable or the
// timeout (in milliseconds) expires. Returns true if readable.
func waitForFd(fd int, timeoutMs int) bool {
	fds := &unix.FdSet{}
	fds.Set(fd)

	tv := unix.Timeval{
		Sec:  int64(timeoutMs / 1000),
		Usec: int64((timeoutMs % 1000) * 1000),
	}

	n, err := unix.Select(fd+1, fds, nil, nil, &tv)
	return err == nil && n > 0
}
