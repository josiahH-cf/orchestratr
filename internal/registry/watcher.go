package registry

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// ReloadFunc is called when the config file changes. It receives the
// path to the config file and should return an error if the reload
// failed (e.g., invalid config). A nil return means the reload was
// successful.
type ReloadFunc func(path string) error

// Watcher monitors a config file and the apps.d directory for
// changes and triggers a reload callback when modifications are
// detected. It debounces rapid filesystem events to avoid redundant
// reloads.
type Watcher struct {
	path     string
	appsDir  string
	debounce time.Duration
	onReload ReloadFunc
	logger   *log.Logger

	mu      sync.Mutex
	running bool
	cancel  context.CancelFunc
	done    chan struct{}
}

// WatcherOption configures a Watcher.
type WatcherOption func(*Watcher)

// WithDebounce sets the debounce duration for filesystem events.
// The default is 100ms.
func WithDebounce(d time.Duration) WatcherOption {
	return func(w *Watcher) {
		w.debounce = d
	}
}

// WithLogger sets the logger for the watcher. If nil, the default
// log.Logger is used.
func WithLogger(l *log.Logger) WatcherOption {
	return func(w *Watcher) {
		w.logger = l
	}
}

// WithAppsDir sets the apps.d directory to watch for drop-in manifests.
// If empty, only the config file is watched.
func WithAppsDir(dir string) WatcherOption {
	return func(w *Watcher) {
		w.appsDir = dir
	}
}

// NewWatcher creates a Watcher that monitors the given config file
// path and calls onReload when changes are detected.
func NewWatcher(path string, onReload ReloadFunc, opts ...WatcherOption) *Watcher {
	w := &Watcher{
		path:     path,
		debounce: 100 * time.Millisecond,
		onReload: onReload,
		logger:   log.Default(),
	}
	for _, opt := range opts {
		opt(w)
	}
	return w
}

// Start begins watching the config file and optionally the apps.d
// directory for changes. It blocks until the context is cancelled
// or Stop is called. Start returns an error if the watcher cannot
// be initialized.
func (w *Watcher) Start(ctx context.Context) error {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return nil
	}

	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		w.mu.Unlock()
		return fmt.Errorf("creating file watcher: %w", err)
	}

	if err := fsw.Add(w.path); err != nil {
		fsw.Close()
		w.mu.Unlock()
		return fmt.Errorf("watching %s: %w", w.path, err)
	}

	// Also watch apps.d/ directory if configured and exists.
	if w.appsDir != "" {
		if _, statErr := os.Stat(w.appsDir); statErr == nil {
			if addErr := fsw.Add(w.appsDir); addErr != nil {
				w.logger.Printf("warning: could not watch apps.d directory %s: %v", w.appsDir, addErr)
			}
		}
	}

	ctx, cancel := context.WithCancel(ctx)
	w.cancel = cancel
	w.running = true
	w.done = make(chan struct{})
	w.mu.Unlock()

	go w.loop(ctx, fsw)
	return nil
}

// Stop stops the watcher and waits for the watch loop to exit.
func (w *Watcher) Stop() {
	w.mu.Lock()
	if !w.running {
		w.mu.Unlock()
		return
	}
	w.cancel()
	done := w.done
	w.mu.Unlock()

	<-done
}

// loop is the main watch loop that processes filesystem events.
func (w *Watcher) loop(ctx context.Context, fsw *fsnotify.Watcher) {
	defer func() {
		fsw.Close()
		w.mu.Lock()
		w.running = false
		close(w.done)
		w.mu.Unlock()
	}()

	var timer *time.Timer
	var timerC <-chan time.Time

	for {
		select {
		case <-ctx.Done():
			if timer != nil {
				timer.Stop()
			}
			return

		case event, ok := <-fsw.Events:
			if !ok {
				return
			}
			// React to write, create, and remove events.
			// Remove is needed for apps.d/ file deletions.
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove) == 0 {
				continue
			}
			// Debounce: reset the timer on each event.
			if timer != nil {
				timer.Stop()
			}
			timer = time.NewTimer(w.debounce)
			timerC = timer.C

		case err, ok := <-fsw.Errors:
			if !ok {
				return
			}
			w.logger.Printf("watcher error: %v", err)

		case <-timerC:
			timerC = nil
			if err := w.onReload(w.path); err != nil {
				w.logger.Printf("config reload failed: %v", err)
			} else {
				w.logger.Printf("config reloaded from %s", w.path)
			}
		}
	}
}
