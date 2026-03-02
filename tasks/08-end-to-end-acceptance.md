# Tasks: End-to-End Acceptance — Open Issue Fixes

**Spec:** /specs/08-end-to-end-acceptance.md

## Status

- Total: 4
- Complete: 4
- Remaining: 0

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
- **Status:** [x] Complete

### Task 2: Trigger with chord argument (OI-9, OI-12)

- **Files:** `internal/registry/registry.go`, `internal/registry/registry_test.go`, `internal/api/server.go`, `internal/api/routes_test.go`, `cmd/orchestratr/main.go`
- **Done when:** `orchestratr trigger c` launches the app matching chord `c` via API, bypassing the hotkey engine
- **Criteria covered:** Phase 8 (manual trigger), OI-9, OI-12
- **Status:** [x] Complete

### Task 3: Platform UX — browser, xdotool, detached (OI-4, OI-5, OI-6)

- **Files:** `internal/gui/gui.go`, `internal/registry/config.go`, `internal/launcher/launcher.go`, `internal/launcher/native_linux.go`, `cmd/orchestratr/install.go`, `configs/example.yml`
- **Done when:** WSL browser opener works; `detached` apps skip PID tracking; install warns if xdotool missing
- **Criteria covered:** OI-4, OI-5, OI-6
- **Status:** [x] Complete

### Task 4: Cleanup & build polish (OI-2, OI-3, OI-10, OI-11)

- **Files:** `internal/hotkey/listener_linux_x11.go`, `cmd/orchestratr/main.go`, `Makefile` (new), `specs/08-end-to-end-acceptance.md`
- **Done when:** No X11 IO error on shutdown; GUI server properly stopped; `make build` injects version; spec log path corrected
- **Criteria covered:** OI-2, OI-3, OI-10, OI-11
- **Status:** [x] Complete

## Test Strategy

- Each task writes tests before implementation (TDD per AGENTS.md)
- `go test ./...` must pass after each task
- `golangci-lint run ./...` must be clean
- Manual smoke test per OI listed in verification section

## Session Log

- 2026-03-02: Plan created from spec OI analysis. Decisions documented above.
- 2026-03-02: Task 1 complete — OI-1 (--foreground + daemonize), OI-7 (status port), OI-8 (file-only logging when backgrounded). Commit bde556f.
- 2026-03-02: Task 2 complete — OI-9/OI-12 (trigger with chord arg). FindByChord already existed. API reads optional chord from body, bypasses engine. Commit 9e63cde.
- 2026-03-02: Task 3 complete — OI-4 (WSL browser openers), OI-5 (detached field + skip PID tracking), OI-6 (xdotool install check). Commit 284990a.
- 2026-03-02: Task 4 complete — OI-2 (spec log path fix), OI-3 (X11 IO error handler), OI-10 (GUI server cleanup), OI-11 (Makefile with version injection). Commit fc43dbd.
