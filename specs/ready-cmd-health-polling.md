# Feature: `ready_cmd` Health Polling

**Status:** Implemented
**Project:** orchestratr

## Description

After launching an app, orchestratr should poll the app's `ready_cmd` (if configured) to confirm the app is responsive, then transition the app's lifecycle state from `launched` to `ready`. The config schema already supports `ready_cmd` and `ready_timeout_ms` fields, and the API state tracker already models `launched`/`ready`/`stopped` states — but nothing currently triggers the `launched → ready` transition. This spec wires the polling.

### Current State (Context for Implementation)

- **Config fields**: `AppEntry.ReadyCmd` (string, optional) and `AppEntry.ReadyTimeoutMs` (int, optional) are parsed from YAML, validated as non-negative, and stored. They are **never read** after loading.
- **State tracker** (`internal/api/state.go`): Thread-safe per-app state with `SetLaunched(name)`, `SetReady(name)`, `SetStopped(name)`. Timestamps tracked. The API exposes these via `/apps/{name}/state`. `SetReady()` is functional but **never called** from anywhere in the codebase.
- **Launch flow** (`cmd/orchestratr/main.go`, `launchApp()`): After `executor.Launch(entry)` succeeds, calls `stateTracker.SetLaunched(appName)`. Spawns a goroutine for `cmd.Wait()` to detect process exit → `stateTracker.SetStopped(appName)`. No readiness check between launched and ready.
- **API endpoints**: `/apps/{name}/launched` (POST) and `/apps/{name}/ready` (POST) exist for external callers. `/apps/{name}/state` (GET) returns current state. The `ready` endpoint works but is only useful for external health reporters — orchestratr itself should also drive readiness.

### Protocol

The `ready_cmd` is a shell command that the app provides. orchestratr executes it and interprets:
- **Exit code 0** = app is ready
- **Non-zero exit code** = app is not yet ready (retry)
- If the command also outputs JSON with `"status": "ok"`, that's a stronger signal, but exit code 0 alone is sufficient

This matches the contract established by espansr (`espansr status --json` returns exit 0 with `{"status": "ok"}`) and will be adopted by templatr.

## Acceptance Criteria

- [ ] After a successful launch, if `ready_cmd` is non-empty, orchestratr spawns a polling goroutine
- [ ] Polling executes `ready_cmd` at 500ms intervals
- [ ] On exit code 0, the app state transitions from `launched` to `ready` and polling stops
- [ ] On timeout (`ready_timeout_ms`, default 5000ms if unset), polling stops with a warning log; app remains in `launched` state (not killed)
- [ ] If the process exits before becoming ready, polling stops immediately
- [ ] `ready_cmd` is executed with the same environment routing as the app's `command` (WSL wrapping for `wsl` environment apps)
- [ ] If `ready_cmd` is empty, the app transitions to `ready` immediately after launch (no polling)
- [ ] Polling respects daemon shutdown (context cancellation)
- [ ] `orchestratr list` and the web GUI display the `ready` state when achieved
- [ ] API endpoint `/apps/{name}/state` reflects the `ready` transition with timestamp

## Affected Areas

| Area | Files |
|------|-------|
| **Modify** | `cmd/orchestratr/main.go` — add `pollReadiness()` goroutine call after launch |
| **Create** | `internal/launcher/readiness.go` — `PollReadiness()` function: environment-aware command execution, interval timer, timeout, context cancellation |
| **Modify** | `internal/api/state.go` — no structural changes, but verify `SetReady()` is correctly thread-safe (it should be) |
| **Modify** | `internal/api/server.go` — no changes needed (state endpoint already works) |

## Constraints

- Polling must not block the daemon's main loop or hotkey event processing
- The polling goroutine must be cancellable (daemon shutdown, app exit, manual stop)
- `ready_cmd` execution must timeout individually (each poll attempt times out after 5s to avoid hangs)
- Commands that output to stdout/stderr should have output discarded (or logged at debug level) — only exit code matters
- If `ready_timeout_ms` is 0 or negative in config, use a sensible default (5000ms)

## Out of Scope

- Continuous health monitoring after `ready` state (this is one-time readiness, not a watchdog)
- Restarting apps that fail readiness (user decides what to do)
- Custom readiness strategies (HTTP endpoints, TCP probes) — shell command is the universal interface
- Parsing `ready_cmd` JSON output beyond exit code (exit code 0 is the contract; JSON is informational)

## Dependencies

- `windows-platform-layer.md` — environment routing for `ready_cmd` execution on Windows (WSL wrapping)
- `drop-in-app-discovery.md` — `ready_cmd` field in drop-in manifests (no code dependency, just schema alignment)

## Notes

### Polling goroutine design

```go
func pollReadiness(ctx context.Context, entry registry.AppEntry, tracker *api.StateTracker, executor Executor) {
    if entry.ReadyCmd == "" {
        tracker.SetReady(entry.Name)
        return
    }

    timeout := time.Duration(entry.ReadyTimeoutMs) * time.Millisecond
    if timeout <= 0 {
        timeout = 5 * time.Second
    }

    deadline := time.After(timeout)
    ticker := time.NewTicker(500 * time.Millisecond)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return // daemon shutting down
        case <-deadline:
            log.Printf("WARN: %s did not become ready within %v", entry.Name, timeout)
            return // leave in "launched" state
        case <-ticker.C:
            if !executor.IsRunning(pid) {
                return // process exited before ready
            }
            if runReadyCmd(entry) == nil {
                tracker.SetReady(entry.Name)
                return
            }
        }
    }
}
```

### Environment-aware `ready_cmd` execution

The `ready_cmd` must be routed through the same environment logic as the main `command`:

- `environment: native` → execute `ready_cmd` via host shell
- `environment: wsl` → execute via `wsl.exe -d <distro> -- bash -lc "<ready_cmd>"`
- `environment: wsl:<distro>` → same with specific distro

This should reuse the environment routing from the executor (either by calling a shared helper or by adding a `RunCommand(env, cmd)` method to the `Executor` interface).

### Integration with launch flow

In `launchApp()` in `main.go`, after the current code:

```go
stateTracker.SetLaunched(appName)
go func() { cmd.Wait(); stateTracker.SetStopped(appName) }()
```

Add:

```go
go pollReadiness(ctx, entry, stateTracker, executor)
```

The poll goroutine must be aware of the exit goroutine — if the process exits, polling should stop. Use a shared context derived from the process lifecycle.
