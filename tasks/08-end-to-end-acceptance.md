# Tasks: End-to-End Acceptance — Open Issue Fixes

**Spec:** /specs/08-end-to-end-acceptance.md

## Status

- Total: 4
- Complete: 0
- Remaining: 4

## Decisions

- **OI-1:** Re-exec with `--foreground` over true fork/daemonize (cross-platform, simpler)
- **OI-8:** Tied to OI-1 — background child logs to file only; `--foreground` keeps stderr
- **OI-5:** Config field `detached: true` over doc-only (actionable for users)
- **OI-9/12:** Chord lookup bypasses engine entirely via registry, works even without hotkey listener

## Task List

### Task 1: Daemon startup & logging (OI-1, OI-8, OI-7)

- **Files:** `cmd/orchestratr/main.go`, `internal/autostart/linux.go`, `cmd/orchestratr/main_test.go`
- **Done when:** `orchestratr start` backgrounds itself via re-exec with `--foreground`; stderr quiet when daemonized; `status` shows port
- **Criteria covered:** Phase 2 (background start, status with port), OI-1, OI-7, OI-8
- **Status:** [ ] Not started

### Task 2: Trigger with chord argument (OI-9, OI-12)

- **Files:** `internal/registry/registry.go`, `internal/registry/registry_test.go`, `internal/api/server.go`, `internal/api/routes_test.go`, `cmd/orchestratr/main.go`
- **Done when:** `orchestratr trigger c` launches the app matching chord `c` via API, bypassing the hotkey engine
- **Criteria covered:** Phase 8 (manual trigger), OI-9, OI-12
- **Status:** [ ] Not started

### Task 3: Platform UX — browser, xdotool, detached (OI-4, OI-5, OI-6)

- **Files:** `internal/gui/gui.go`, `internal/registry/config.go`, `internal/launcher/launcher.go`, `internal/launcher/native_linux.go`, `cmd/orchestratr/install.go`, `configs/example.yml`
- **Done when:** WSL browser opener works; `detached` apps skip PID tracking; install warns if xdotool missing
- **Criteria covered:** OI-4, OI-5, OI-6
- **Status:** [ ] Not started

### Task 4: Cleanup & build polish (OI-2, OI-3, OI-10, OI-11)

- **Files:** `internal/hotkey/listener_linux_x11.go`, `cmd/orchestratr/main.go`, `Makefile` (new), `specs/08-end-to-end-acceptance.md`
- **Done when:** No X11 IO error on shutdown; GUI server properly stopped; `make build` injects version; spec log path corrected
- **Criteria covered:** OI-2, OI-3, OI-10, OI-11
- **Status:** [ ] Not started

## Test Strategy

- Each task writes tests before implementation (TDD per AGENTS.md)
- `go test ./...` must pass after each task
- `golangci-lint run ./...` must be clean
- Manual smoke test per OI listed in verification section

## Session Log

- 2026-03-02: Plan created from spec OI analysis. Decisions documented above.
