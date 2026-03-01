# Feature: Hot-Reload Completeness

**Status:** Not started
**Parent:** /specs/03-app-registry.md, /specs/04-http-api.md
**Project:** orchestratr

## Description

When the config is hot-reloaded (via file watcher or `POST /reload`), the registry is swapped but two other stateful components are not updated: the hotkey engine still uses its original chord map, and the StateTracker retains entries for apps that may have been removed. This spec ensures a config reload propagates fully through the system.

## Acceptance Criteria

- [ ] After a successful config reload, the hotkey engine's chord map is updated to match the new config's apps (new chords added, removed chords dropped, changed chords updated)
- [ ] The `Engine` exposes a `SwapChords(chords []Chord) error` method that atomically replaces the chord lookup table while the engine is running
- [ ] `SwapChords` rejects duplicate chords (same validation as `NewEngine`) and returns an error without modifying state if duplicates are found
- [ ] After a successful config reload, the `StateTracker` removes entries for apps no longer in the registry (stale state cleanup)
- [ ] The `StateTracker` exposes a `Sync(appNames []string)` method that removes entries whose names are not in the provided list
- [ ] The shared reload function in `runStart()` calls both `SwapChords` and `Sync` after a successful `reg.Swap()`
- [ ] Existing tests continue to pass; new methods have dedicated unit tests

## Affected Areas

| Area | Files |
|------|-------|
| **Modify** | `internal/hotkey/engine.go` — add `SwapChords()` method |
| **Modify** | `internal/hotkey/engine_test.go` — tests for `SwapChords` |
| **Modify** | `internal/api/state.go` — add `Sync()` method to `StateTracker` |
| **Modify** | `internal/api/state_test.go` — tests for `Sync` |
| **Modify** | `cmd/orchestratr/main.go` — extend `reloadFn` to call `SwapChords` and `Sync` |

## Constraints

- `SwapChords` must be safe to call while the engine event loop is running (use the existing `mu` mutex)
- `SwapChords` must not disrupt in-flight chord matching (if the engine is in `StateChordWait`, the swap takes effect after the current chord window expires or completes)
- The reload function must handle `SwapChords` errors gracefully — log a warning but don't roll back the registry swap (the registry and API should reflect the new config even if the hotkey engine couldn't update)

## Out of Scope

- Re-registering the leader key on reload (the leader key change requires an engine restart, which is a larger concern)
- Process tracking for removed apps (spec 05 territory)

## Dependencies

- `/specs/03a-registry-daemon-integration.md` — watcher must be wired first, or this can be tested via `POST /reload` alone
- `internal/hotkey/engine.go` — existing `Engine` struct and `chords` map
- `internal/api/state.go` — existing `StateTracker` struct

## Notes

### `SwapChords` design

```go
// SwapChords atomically replaces the engine's chord lookup table.
// Returns an error if the new chord set contains duplicates.
func (e *Engine) SwapChords(chords []Chord) error {
    newMap := make(map[string]string, len(chords))
    for _, c := range chords {
        canonical := c.Key.String()
        if _, dup := newMap[canonical]; dup {
            return fmt.Errorf("duplicate chord %q", canonical)
        }
        newMap[canonical] = c.Action
    }

    e.mu.Lock()
    e.chords = newMap
    e.mu.Unlock()
    return nil
}
```

The engine's `loop()` already reads `e.chords` under a read lock (implicitly via the select — but actually it accesses it directly). The `loop()` reads `e.chords[canonical]` without holding `mu`. This is a **pre-existing issue**: `chords` is a map read from the loop goroutine while `SwapChords` writes from another goroutine. The fix: wrap `e.chords` access in the loop with `e.mu.RLock()`.

### `StateTracker.Sync` design

```go
// Sync removes state entries for apps not in the provided name list.
func (st *StateTracker) Sync(appNames []string) {
    allowed := make(map[string]bool, len(appNames))
    for _, name := range appNames {
        allowed[name] = true
    }

    st.mu.Lock()
    defer st.mu.Unlock()
    for name := range st.apps {
        if !allowed[name] {
            delete(st.apps, name)
        }
    }
}
```

### Reload function extension

In `runStart()`, the reload closure becomes:

```go
reloadFn := func() (*registry.Config, error) {
    newCfg, loadErr := registry.LoadAndValidate(cfgPath)
    if loadErr != nil {
        return nil, loadErr
    }
    if reg != nil {
        reg.Swap(*newCfg)
    }
    // Update hotkey chords.
    if hotkeyEngine != nil {
        chords := buildChords(newCfg.Apps)
        if err := hotkeyEngine.SwapChords(chords); err != nil {
            logger.Printf("warning: hotkey chord update failed: %v", err)
        }
    }
    // Clean stale app state.
    apiSrv.SyncState(newCfg.Apps) // or access state tracker directly
    return newCfg, nil
}
```

Note: `hotkeyEngine` is declared after `reloadFn` in the current code. The reload closure must capture a pointer or the engine setup must move before the closure definition. Consider using a `*hotkey.Engine` pointer variable declared before the closure, then assigned later.
