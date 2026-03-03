package registry

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestWatcher_DetectsFileChange(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")

	// Write initial config.
	initial := "leader_key: ctrl+space\napps: []\n"
	if err := os.WriteFile(path, []byte(initial), 0o644); err != nil {
		t.Fatal(err)
	}

	var reloadCount atomic.Int32
	reload := func(p string) error {
		reloadCount.Add(1)
		return nil
	}

	w := NewWatcher(path, reload, WithDebounce(50*time.Millisecond))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := w.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer w.Stop()

	// Give the watcher time to register.
	time.Sleep(100 * time.Millisecond)

	// Modify the file.
	updated := "leader_key: alt+space\napps: []\n"
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		t.Fatal(err)
	}

	// Wait for the debounce + processing.
	deadline := time.After(2 * time.Second)
	for reloadCount.Load() == 0 {
		select {
		case <-deadline:
			t.Fatal("timed out waiting for reload callback")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}

	if got := reloadCount.Load(); got < 1 {
		t.Errorf("reload count = %d, want >= 1", got)
	}
}

func TestWatcher_InvalidConfigKeepsOld(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")

	initial := "leader_key: ctrl+space\napps: []\n"
	if err := os.WriteFile(path, []byte(initial), 0o644); err != nil {
		t.Fatal(err)
	}

	// Set up a registry and watcher that does a full load+validate cycle.
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	reg := NewRegistry(*cfg)

	reload := func(p string) error {
		newCfg, err := LoadAndValidate(p)
		if err != nil {
			return err
		}
		reg.Swap(*newCfg)
		return nil
	}

	w := NewWatcher(path, reload, WithDebounce(50*time.Millisecond))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := w.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer w.Stop()

	time.Sleep(100 * time.Millisecond)

	// Write invalid config (duplicate chords).
	invalid := `
apps:
  - name: a
    chord: x
    command: c1
  - name: b
    chord: x
    command: c2
`
	if err := os.WriteFile(path, []byte(invalid), 0o644); err != nil {
		t.Fatal(err)
	}

	// Wait for the reload attempt.
	time.Sleep(300 * time.Millisecond)

	// Registry should still have the original config (empty apps).
	if reg.Len() != 0 {
		t.Errorf("registry should keep old config on invalid reload, got %d apps", reg.Len())
	}
}

func TestWatcher_ContextCancellation(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")

	if err := os.WriteFile(path, []byte("apps: []\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	reload := func(p string) error { return nil }
	w := NewWatcher(path, reload, WithDebounce(50*time.Millisecond))

	ctx, cancel := context.WithCancel(context.Background())
	if err := w.Start(ctx); err != nil {
		t.Fatal(err)
	}

	// Cancel immediately.
	cancel()

	// Stop should return quickly (within 1 second).
	done := make(chan struct{})
	go func() {
		w.Stop()
		close(done)
	}()

	select {
	case <-done:
		// Good.
	case <-time.After(2 * time.Second):
		t.Fatal("Stop() did not return after context cancellation")
	}
}

func TestWatcher_StopWithoutStart(t *testing.T) {
	// Calling Stop on an unstarted watcher should not panic.
	w := NewWatcher("/nonexistent", func(string) error { return nil })
	w.Stop() // Should be a no-op.
}

func TestWatcher_DoubleStart(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")

	if err := os.WriteFile(path, []byte("apps: []\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	reload := func(p string) error { return nil }
	w := NewWatcher(path, reload)

	ctx := context.Background()
	if err := w.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer w.Stop()

	// Second Start should be a no-op, not an error.
	if err := w.Start(ctx); err != nil {
		t.Errorf("second Start() error = %v, want nil", err)
	}
}

func TestWatcher_DebounceCollapses(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")

	if err := os.WriteFile(path, []byte("apps: []\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var reloadCount atomic.Int32
	reload := func(p string) error {
		reloadCount.Add(1)
		return nil
	}

	// Use a longer debounce to make rapid writes collapse.
	w := NewWatcher(path, reload, WithDebounce(200*time.Millisecond))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := w.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer w.Stop()

	time.Sleep(100 * time.Millisecond)

	// Write rapidly 5 times within the debounce window.
	for i := range 5 {
		content := fmt.Sprintf("leader_key: key%d\napps: []\n", i)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		time.Sleep(20 * time.Millisecond)
	}

	// Wait for debounce to settle + processing.
	time.Sleep(500 * time.Millisecond)

	// Should have collapsed to 1 or at most 2 reloads, not 5.
	got := reloadCount.Load()
	if got > 2 {
		t.Errorf("reload count = %d after rapid writes, want <= 2 (debounce should collapse)", got)
	}
	if got < 1 {
		t.Errorf("reload count = %d, want >= 1", got)
	}
}

func TestWatcher_ReloadError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")

	if err := os.WriteFile(path, []byte("apps: []\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var reloadCalled atomic.Int32
	reload := func(p string) error {
		reloadCalled.Add(1)
		return fmt.Errorf("simulated reload error")
	}

	w := NewWatcher(path, reload, WithDebounce(50*time.Millisecond))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := w.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer w.Stop()

	time.Sleep(100 * time.Millisecond)

	if err := os.WriteFile(path, []byte("leader_key: new\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Wait for reload attempt. Watcher should not crash on error.
	time.Sleep(300 * time.Millisecond)

	if reloadCalled.Load() < 1 {
		t.Error("reload should have been called even when it returns an error")
	}
}

func TestWatcher_DetectsAppsDirChange(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yml")
	appsDir := filepath.Join(dir, "apps.d")

	if err := os.WriteFile(cfgPath, []byte("apps: []\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(appsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	var reloadCount atomic.Int32
	reload := func(p string) error {
		reloadCount.Add(1)
		return nil
	}

	w := NewWatcher(cfgPath, reload,
		WithDebounce(50*time.Millisecond),
		WithAppsDir(appsDir),
	)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := w.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer w.Stop()

	time.Sleep(100 * time.Millisecond)

	// Create a new drop-in file in apps.d/.
	if err := os.WriteFile(filepath.Join(appsDir, "newapp.yml"), []byte("name: newapp\nchord: n\ncommand: echo new\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	deadline := time.After(2 * time.Second)
	for reloadCount.Load() == 0 {
		select {
		case <-deadline:
			t.Fatal("timed out waiting for reload after apps.d change")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}

	if got := reloadCount.Load(); got < 1 {
		t.Errorf("reload count = %d, want >= 1", got)
	}
}

func TestWatcher_DetectsAppsDirDeletion(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yml")
	appsDir := filepath.Join(dir, "apps.d")

	if err := os.WriteFile(cfgPath, []byte("apps: []\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(appsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Pre-create a file so we can delete it.
	dropinPath := filepath.Join(appsDir, "toremove.yml")
	if err := os.WriteFile(dropinPath, []byte("name: toremove\nchord: r\ncommand: echo rm\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var reloadCount atomic.Int32
	reload := func(p string) error {
		reloadCount.Add(1)
		return nil
	}

	w := NewWatcher(cfgPath, reload,
		WithDebounce(50*time.Millisecond),
		WithAppsDir(appsDir),
	)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := w.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer w.Stop()

	time.Sleep(100 * time.Millisecond)

	// Delete the drop-in file.
	if err := os.Remove(dropinPath); err != nil {
		t.Fatal(err)
	}

	deadline := time.After(2 * time.Second)
	for reloadCount.Load() == 0 {
		select {
		case <-deadline:
			t.Fatal("timed out waiting for reload after apps.d file deletion")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}
