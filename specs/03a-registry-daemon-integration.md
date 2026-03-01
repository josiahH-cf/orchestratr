# Feature: Registry–Daemon Integration

**Status:** Not started
**Parent:** /specs/03-app-registry.md
**Project:** orchestratr

## Description

The file watcher (hot-reload) and `EnsureDefaults()` (first-run config creation) were built and unit-tested in the registry package but never wired into the daemon's `runStart()` flow. This spec covers connecting those two components so the daemon actually uses them at runtime.

## Acceptance Criteria

- [ ] On first run with no config file, `EnsureDefaults()` is called before `LoadAndValidate()`, creating `~/.config/orchestratr/config.yml` with default values
- [ ] A `registry.Watcher` is started as part of `runStart()`, watching the active config file path for changes
- [ ] When the config file is modified on disk, the watcher triggers a reload that runs `LoadAndValidate()` and calls `reg.Swap()` on success
- [ ] When the config file is modified with invalid content, the old config remains active and a warning is logged
- [ ] The watcher is stopped cleanly via `defer` on daemon shutdown (no goroutine leak)
- [ ] All existing tests continue to pass unchanged

## Affected Areas

| Area | Files |
|------|-------|
| **Modify** | `cmd/orchestratr/main.go` — `runStart()`: call `EnsureDefaults()`, create and start `Watcher` |

## Constraints

- The reload function used by the watcher must be the **same** `reloadFn` closure already passed to `api.NewServer`, so that `POST /reload` and file-watcher reload share a single code path (avoids race conditions and inconsistent behavior)
- Extract the shared reload closure **before** passing it to both `api.NewServer` and `registry.NewWatcher`
- The watcher's `ReloadFunc` signature is `func(path string) error`, while the API's `ReloadFunc` is `func() (*registry.Config, error)`. Adapt with a thin wrapper — do not duplicate the load-validate-swap logic

## Out of Scope

- Updating the hotkey engine's chord map on reload (see `/specs/04b-hot-reload-completeness.md`)
- StateTracker reconciliation on reload (see `/specs/04b-hot-reload-completeness.md`)
- New CLI commands

## Dependencies

- `internal/registry` — `Watcher`, `EnsureDefaults`, `LoadAndValidate` (all already implemented)
- `internal/api` — `Server` (already wired in `runStart`)

## Notes

### Current state of `runStart()`

The reload closure is defined at [main.go lines 104–112](cmd/orchestratr/main.go#L104-L112):

```go
reloadFn := func() (*registry.Config, error) {
    newCfg, loadErr := registry.LoadAndValidate(cfgPath)
    if loadErr != nil {
        return nil, loadErr
    }
    if reg != nil {
        reg.Swap(*newCfg)
    }
    return newCfg, nil
}
```

The watcher needs a `func(path string) error` wrapper around this same logic. Suggested approach:

```go
watcherReload := func(path string) error {
    _, err := reloadFn()
    return err
}
```

### `EnsureDefaults` placement

Call `EnsureDefaults(cfgPath)` **before** `LoadAndValidate(cfgPath)` in `runStart()`. If `EnsureDefaults` fails (e.g., permission error), log a warning and continue — the daemon should still attempt to load whatever config exists.

### Watcher lifecycle

```go
w := registry.NewWatcher(cfgPath, watcherReload, registry.WithLogger(logger))
if err := w.Start(ctx); err != nil {
    logger.Printf("warning: file watcher not available: %v", err)
} else {
    defer w.Stop()
}
```

The watcher takes a `context.Context`. Use a cancellable context derived from the daemon's lifecycle. Since `d.Run(ctx)` blocks on signal, the watcher context should be cancelled when `Run` returns.
