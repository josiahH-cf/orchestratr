//go:build !linux && !windows

package launcher

import (
	"github.com/josiahH-cf/orchestratr/internal/registry"
)

// StubExecutor is a placeholder for platforms where launching is not
// yet implemented. All methods return ErrNotImplemented.
type StubExecutor struct{}

// NewPlatformExecutor creates the platform-appropriate Executor. On
// unsupported platforms this returns a StubExecutor.
func NewPlatformExecutor(_ ...Option) Executor {
	return NewStubExecutor()
}

// NewStubExecutor creates a StubExecutor.
func NewStubExecutor() *StubExecutor { return &StubExecutor{} }

// Launch returns ErrNotImplemented.
func (s *StubExecutor) Launch(_ registry.AppEntry) (*Result, error) {
	return nil, ErrNotImplemented
}

// Stop returns ErrNotImplemented.
func (s *StubExecutor) Stop(_ string) error { return ErrNotImplemented }

// StopAll is a no-op on unsupported platforms.
func (s *StubExecutor) StopAll() {}

// IsRunning always returns false on unsupported platforms.
func (s *StubExecutor) IsRunning(_ string) bool { return false }

// PID always returns 0, false on unsupported platforms.
func (s *StubExecutor) PID(_ string) (int, bool) { return 0, false }
