# Tasks: Startup Diagnostic + Error Reporting

**Spec:** /specs/startup-diagnostic.md
**Branch:** `feat/startup-diagnostic`

## Status

- Total: 5
- Complete: 5
- Remaining: 0

## Task List

### Task 1: Error State on AppState

- **Files:** `internal/api/state.go`, `internal/api/state_test.go`
- **Done when:** `AppState` has `Error` and `ErrorAt` fields; `SetError(name, msg)` and `ClearError(name)` methods exist; `SetLaunched` clears error; all existing + new tests pass
- **Criteria covered:** "Launch failures populate an `error` field on the app's state", "Error state is transient — cleared on next successful launch"
- **Status:** [x] Complete

### Task 2: Wire Error State into Launch Flow

- **Files:** `cmd/orchestratr/main.go`, `cmd/orchestratr/main_test.go`, `internal/api/server.go`
- **Done when:** `launchApp()` and the launch API endpoint call `SetError()` on failure; `/apps/{name}/state` includes `error` and `error_at` in JSON; tests verify error propagation
- **Criteria covered:** "Launch failure errors include: command attempted, error message, environment, and timestamp", "Launch failures populate an `error` field visible via `/apps/{name}/state`"
- **Status:** [x] Complete

### Task 3: `orchestratr doctor` Command

- **Files:** `cmd/orchestratr/doctor.go`, `cmd/orchestratr/doctor_test.go`
- **Done when:** `orchestratr doctor` runs all checks (daemon, config, apps.d, commands, WSL, ready_cmd); output has PASS/FAIL/WARN per check; `--json` flag outputs machine-readable JSON; works even when daemon is not running
- **Criteria covered:** All `orchestratr doctor` acceptance criteria (6 items)
- **Status:** [x] Complete

### Task 4: Unmatched Chord Logging + Verbose Mode

- **Files:** `internal/hotkey/engine.go`, `internal/hotkey/engine_test.go`, `cmd/orchestratr/main.go`
- **Done when:** Unmatched chords are logged at debug level with the chord character; `orchestratr start --verbose` streams structured events to stderr (leader pressed, chord received, app launching/launched/focused, launch failed, ready pass/fail, config reloaded)
- **Criteria covered:** "Unmatched chords are logged at debug level", "`orchestratr start --verbose` streams events to stderr"
- **Status:** [x] Complete

### Task 5: Web GUI Error Badge

- **Files:** `internal/gui/static/index.html`
- **Done when:** GUI shows a red `error` status badge when an app has an error; tooltip displays the error message; existing status badges still work
- **Criteria covered:** "Web GUI shows an `error` status badge (red) with the error message in a tooltip"
- **Status:** [x] Complete

## Test Strategy

| Criterion | Task | Test |
|-----------|------|------|
| Error field on app state | 1 | `state_test.go`: `TestSetError`, `TestClearErrorOnRelaunch` |
| Error in API response | 2 | `main_test.go` or `server_test.go`: launch failure → GET state includes error |
| Doctor checks | 3 | `doctor_test.go`: each check type with pass/fail scenarios |
| Doctor --json | 3 | `doctor_test.go`: JSON output parsing |
| Unmatched chord logging | 4 | `engine_test.go`: unmatched chord produces log entry |
| Verbose event streaming | 4 | `main_test.go`: --verbose flag produces event lines on stderr |
| GUI error badge | 5 | Manual verification (HTML/JS) |

## Session Log

<!-- Append after each session: date, completed, blockers -->
