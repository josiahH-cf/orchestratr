package hotkey

import "runtime"

// NewStubListener returns a Listener that reports the current platform
// but returns ErrNotImplemented from Start. It is used for CI testing
// and on platforms where the real capture API is not yet implemented.
func NewStubListener() Listener {
	return &stubListener{
		info: PlatformInfo{
			OS:     runtime.GOOS,
			Method: "stub",
		},
	}
}

type stubListener struct {
	info   PlatformInfo
	leader Key
}

func (s *stubListener) Info() PlatformInfo { return s.info }

func (s *stubListener) Register(leader Key) (string, error) {
	s.leader = leader
	return CheckConflicts(leader), nil
}

// Start returns ErrNotImplemented. The stub listener cannot capture
// real keyboard events.
func (s *stubListener) Start(events chan<- KeyEvent) error {
	return ErrNotImplemented
}

// GrabKeyboard is a no-op on the stub listener.
func (s *stubListener) GrabKeyboard() error { return nil }

// UngrabKeyboard is a no-op on the stub listener.
func (s *stubListener) UngrabKeyboard() {}

// Stop is a no-op on the stub listener.
func (s *stubListener) Stop() error {
	return nil
}
