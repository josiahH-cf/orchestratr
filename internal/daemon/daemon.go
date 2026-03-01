package daemon

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// State represents the daemon's current operating state.
type State int

const (
	// StateStopped means the daemon is not running.
	StateStopped State = iota
	// StateRunning means the daemon is running and processing events.
	StateRunning
	// StatePaused means the daemon is running but hotkey listening is disabled.
	StatePaused
)

// String returns a human-readable representation of the state.
func (s State) String() string {
	switch s {
	case StateStopped:
		return "stopped"
	case StateRunning:
		return "running"
	case StatePaused:
		return "paused"
	default:
		return fmt.Sprintf("unknown(%d)", int(s))
	}
}

// Config holds daemon configuration values.
type Config struct {
	// LogLevel controls the log verbosity ("debug", "info", "warn", "error").
	LogLevel string
	// APIPort is the localhost port for the health endpoint.
	APIPort int
	// LockPath overrides the default PID lock file location.
	LockPath string
}

// Daemon manages the orchestratr background process lifecycle.
type Daemon struct {
	cfg    Config
	mu     sync.RWMutex
	state  State
	logger *log.Logger
}

// New creates a new Daemon with the given configuration.
func New(cfg Config) *Daemon {
	return &Daemon{
		cfg:    cfg,
		state:  StateStopped,
		logger: log.New(os.Stderr, "orchestratr: ", log.LstdFlags),
	}
}

// State returns the daemon's current state.
func (d *Daemon) State() State {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.state
}

// Run starts the daemon and blocks until the context is cancelled or a
// shutdown signal is received. It transitions through Running state and
// returns to Stopped on exit.
func (d *Daemon) Run(ctx context.Context) error {
	d.mu.Lock()
	d.state = StateRunning
	d.mu.Unlock()

	d.logger.Println("daemon starting")

	// Listen for OS shutdown signals.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	defer signal.Stop(sigCh)

	select {
	case <-ctx.Done():
		d.logger.Println("daemon stopping (context cancelled)")
	case sig := <-sigCh:
		d.logger.Printf("daemon stopping (received %s)", sig)
	}

	d.mu.Lock()
	d.state = StateStopped
	d.mu.Unlock()

	d.logger.Println("daemon stopped")
	return nil
}

// Pause transitions the daemon to paused state, disabling hotkey listening
// without stopping the process.
func (d *Daemon) Pause() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.state != StateRunning {
		return fmt.Errorf("cannot pause: daemon is %s", d.state)
	}
	d.state = StatePaused
	d.logger.Println("daemon paused")
	return nil
}

// Resume transitions the daemon from paused back to running state.
func (d *Daemon) Resume() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.state != StatePaused {
		return fmt.Errorf("cannot resume: daemon is %s", d.state)
	}
	d.state = StateRunning
	d.logger.Println("daemon resumed")
	return nil
}

// SetLogger replaces the daemon's logger.
func (d *Daemon) SetLogger(l *log.Logger) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.logger = l
}
