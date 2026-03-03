# Feature: Startup Diagnostic + Error Reporting

**Status:** Implemented
**Project:** orchestratr

## Description

orchestratr currently fails silently when things go wrong — unmatched chords produce no output, launch failures aren't surfaced to the user, and there's no way to verify the system is healthy without reading logs. This spec adds an `orchestratr doctor` diagnostic command, structured error reporting on launch failures, and a `--verbose` mode for real-time event streaming. The goal: when a user presses a chord and nothing happens, they can immediately find out why.

### Current State (Context for Implementation)

- **Launch failures** (`cmd/orchestratr/main.go`, `launchApp()`): If `executor.Launch()` fails, the error is logged but not surfaced to the user via the API, tray, or any visible channel. The app silently stays in `stopped` state.
- **Unmatched chords**: In `engine.go`, if a chord doesn't match any registered app, the engine returns to idle state with no log entry, no API event, nothing.
- **Daemon logs**: Written to `~/.config/orchestratr/orchestratr.log` (configured in `daemon/log.go`). Standard `log.Printf` — no structured logging, no level filtering.
- **State tracker** (`internal/api/state.go`): Tracks `launched`/`ready`/`stopped` per app. No `error` state or error message field.
- **Web GUI** (`internal/gui/static/index.html`): Shows status badges (`stopped`/`launched`/`ready`). No error state display.
- **CLI**: `orchestratr status` shows daemon status (running/stopped, PID, port). `orchestratr list` shows apps with state. Neither provides diagnostic information.
- **Tray**: `HeadlessProvider` — no real tray. Tray notifications are a no-op.

## Acceptance Criteria

- [ ] `orchestratr doctor` checks and reports: daemon running (PID + port), config valid (parse + validate), `apps.d/` scan results, all app commands resolvable (via `which`/`where`), WSL available (if any WSL-environment apps configured), per-app `ready_cmd` syntax validity
- [ ] `orchestratr doctor` output is structured: each check has a PASS/FAIL/WARN status with explanation
- [ ] Launch failures populate an `error` field on the app's state (visible via `/apps/{name}/state` API)
- [ ] Launch failure errors include: command attempted, error message, environment, and timestamp
- [ ] Web GUI shows an `error` status badge (red) with the error message in a tooltip when an app fails to launch
- [ ] Unmatched chords are logged at debug level with the chord character received
- [ ] `orchestratr start --verbose` streams events to stderr in real-time: leader key pressed, chord received, app launching, app launched, app focused, launch failed, ready check pass/fail, config reloaded
- [ ] `orchestratr doctor --json` outputs the diagnostic results as machine-readable JSON

## Affected Areas

| Area | Files |
|------|-------|
| **Create** | `cmd/orchestratr/main.go` — `doctor` subcommand implementation |
| **Modify** | `cmd/orchestratr/main.go` — `--verbose` flag on `start`, event logging in `launchApp()` and engine callback |
| **Modify** | `internal/api/state.go` — add `Error string` and `ErrorAt time.Time` fields to app state |
| **Modify** | `internal/api/server.go` — include error in `/apps/{name}/state` response |
| **Modify** | `internal/hotkey/engine.go` — log unmatched chords via a configurable callback or log function |
| **Modify** | `internal/gui/static/index.html` — error badge + tooltip display |

## Constraints

- `orchestratr doctor` must work even when the daemon is not running (it checks config and command availability independently)
- `--verbose` must not interfere with daemon backgrounding — verbose output goes to stderr of the foreground process
- Error state is transient — cleared on next successful launch
- Doctor checks should complete in under 5 seconds (with timeouts on network/WSL checks)
- No new dependencies for logging — use standard `log` package with level prefixes

## Out of Scope

- Real system tray notifications (separate feature, needs native tray implementation)
- Continuous health monitoring / watchdog behavior
- Structured logging library (e.g., zerolog, zap) — keep using standard `log` for now
- Remote diagnostics or telemetry

## Dependencies

- `drop-in-app-discovery.md` — doctor should report `apps.d/` scan results
- `ready-cmd-health-polling.md` — doctor should verify `ready_cmd` syntax; error state used by both
- `windows-platform-layer.md` — WSL availability check relevant on Windows

## Notes

### Doctor checks table

| Check | Condition | Status |
|-------|-----------|--------|
| Daemon running | PID file exists, process alive, API responds | PASS / FAIL |
| Config parseable | `config.yml` valid YAML | PASS / FAIL |
| Config valid | All required fields present, no duplicate chords | PASS / FAIL / WARN |
| `apps.d/` directory | Exists and readable | PASS / WARN (missing) |
| `apps.d/` files | Each file valid | PASS / FAIL per file |
| App commands | Each app's `command` binary resolvable via PATH | PASS / WARN per app |
| WSL available | `wsl.exe` in PATH, at least one distro listed | PASS / FAIL / SKIP (no WSL apps) |
| Ready commands | Each app's `ready_cmd` syntactically valid | PASS / WARN per app |

### Error state on AppState

```go
type AppState struct {
    Launched  bool      `json:"launched"`
    Ready     bool      `json:"ready"`
    LaunchedAt time.Time `json:"launched_at,omitempty"`
    ReadyAt   time.Time `json:"ready_at,omitempty"`
    StoppedAt time.Time `json:"stopped_at,omitempty"`
    Error     string    `json:"error,omitempty"`
    ErrorAt   time.Time `json:"error_at,omitempty"`
}
```

### Verbose event format

```
[2026-03-03T14:22:01] EVENT leader_key_pressed
[2026-03-03T14:22:01] EVENT chord_received chord=e
[2026-03-03T14:22:01] EVENT app_launching name=espansr command="espansr gui" env=wsl
[2026-03-03T14:22:02] EVENT app_launched name=espansr pid=12345
[2026-03-03T14:22:03] EVENT app_ready name=espansr elapsed=1.2s
```

Or on failure:

```
[2026-03-03T14:22:01] EVENT app_launch_failed name=espansr error="exec: espansr: not found"
```

### Doctor output format (text)

```
orchestratr doctor

  Daemon ............ PASS  (PID 4521, port 9876)
  Config ............ PASS  (2 apps in config.yml)
  apps.d/ ........... PASS  (1 file: espansr.yml)
  espansr command ... PASS  (wsl.exe → espansr gui)
  espansr ready_cmd . PASS  (wsl.exe → espansr status --json)
  templatr command .. WARN  (templatr not found in PATH)
  WSL available ..... PASS  (Ubuntu-22.04)

  3 passed, 1 warning, 0 failed
```
