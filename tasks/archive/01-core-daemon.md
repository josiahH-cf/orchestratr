# Tasks: Core Daemon

**Spec:** /specs/01-core-daemon.md

## Status

- Total: 3
- Complete: 3
- Remaining: 0

## Task List

### Task 1: Daemon lifecycle, lock, and signal handling

- **Files:** `internal/daemon/daemon.go`, `internal/daemon/lock.go`, `internal/daemon/daemon_test.go`, `internal/daemon/lock_test.go`
- **Done when:** Daemon struct starts, acquires PID lock, handles SIGTERM/SIGINT gracefully, supports Pause/Resume state transitions, and rejects duplicate instances
- **Criteria covered:** 1 (background process), 7 (graceful shutdown), 8 (pause mode)
- **Status:** [x] Complete

### Task 2: Logging and health endpoint

- **Files:** `internal/daemon/log.go`, `internal/daemon/log_test.go`, `internal/daemon/health.go`, `internal/daemon/health_test.go`
- **Done when:** Daemon writes logs to a rotating log file with configurable level; localhost-only HTTP server on `/health` returns `{"status":"ok"}`; port is written to discovery file
- **Criteria covered:** 5 (rotating log), 6 (health endpoint)
- **Status:** [x] Complete

### Task 3: CLI commands and tray interface

- **Files:** `cmd/orchestratr/main.go`, `cmd/orchestratr/main_test.go`, `internal/tray/tray.go`, `internal/tray/tray_test.go`
- **Done when:** `orchestratr start` launches daemon (foreground for now), `orchestratr stop` sends shutdown, `orchestratr status` reports running/stopped; tray Provider interface defined with a headless stub for CI
- **Criteria covered:** 2 (tray interface), CLI integration
- **Status:** [x] Complete

## Test Strategy

| Criterion | Tested in |
|-----------|-----------|
| Background process / no window | Task 1 — daemon starts without blocking |
| Graceful shutdown (SIGTERM) | Task 1 — signal handler test |
| Pause mode | Task 1 — state transition tests |
| Single-instance lock | Task 1 — concurrent start rejection |
| Rotating log | Task 2 — log output verification |
| Health endpoint | Task 2 — HTTP response test |
| CLI start/stop/status | Task 3 — integration tests |
| Tray Provider interface | Task 3 — stub implementation test |

## Session Log

### 2026-03-01

- Completed all 3 tasks
- Daemon lifecycle with PID lock, signal handling, pause/resume
- Health endpoint on localhost with port discovery file
- CLI start/stop/status commands
- Tray Provider interface with HeadlessProvider stub
- All tests passing (4 packages, 0 failures)
