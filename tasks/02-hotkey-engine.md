# Tasks: Hotkey Engine

**Spec:** /specs/02-hotkey-engine.md
**Decision:** /decisions/0002-wayland-hotkey-strategy.md

## Status

- Total: 3
- Complete: 0
- Remaining: 3

## Task List

### Task 1: Core Interface, Key Types, and Chord Engine

- **Files:** `internal/hotkey/keys.go`, `internal/hotkey/keys_test.go`, `internal/hotkey/engine.go`, `internal/hotkey/engine_test.go`, `internal/hotkey/listener.go`
- **Done when:** Key parsing (`"ctrl+space"` → Key struct), Listener interface, and Engine state machine (Idle → ChordWait → dispatch/timeout) are implemented with passing tests. The chord engine matches incoming keystrokes against registered chords and dispatches named actions.
- **Criteria covered:**
  - Leader key is configurable via app config (not hardcoded)
  - After leader activation, chord keystrokes are captured for the timeout window
  - If mapped chord received: dispatch action; if unmapped or timeout: exit silently
- **Status:** [ ] Not started

### Task 2: Platform Listener Stubs with Build Tags

- **Files:** `internal/hotkey/listener_linux.go`, `internal/hotkey/listener_darwin.go`, `internal/hotkey/listener_windows.go`, `internal/hotkey/listener_stub.go`, `internal/hotkey/listener_test.go`
- **Done when:** Each platform file compiles under its build tag, returns a Listener that reports its platform/method at startup, and returns an `ErrNotImplemented` for platforms that don't have real capture yet. A stub listener (build tag `!linux,!darwin,!windows` or explicit `stub` tag) exists for CI testing. Conflict warning for common shortcuts is implemented in shared code.
- **Criteria covered:**
  - Global leader key captured on Windows / macOS / Linux/X11 / Linux/Wayland (stubs — full OS API implementations are follow-up work)
  - Engine reports current platform and registration method at startup
  - Warning emitted if leader key conflicts with common OS shortcuts
  - Hotkey registrations cleanly released on shutdown
- **Status:** [ ] Not started

### Task 3: Daemon Integration

- **Files:** `cmd/orchestratr/main.go`, `internal/daemon/daemon.go`
- **Done when:** The daemon creates a hotkey Engine from config, starts it alongside the API server, stops it cleanly on shutdown/pause, and reports the platform method in the startup log. The `orchestratr trigger` CLI subcommand exists for Wayland manual fallback.
- **Criteria covered:**
  - Hotkey registrations cleanly released on daemon pause, shutdown, or crash recovery
  - Engine reports platform and registration method at startup
  - Chord keystrokes captured without leaking to focused application (via Engine's suppression model)
- **Status:** [ ] Not started

## Test Strategy

| Acceptance Criterion | Test Location |
|---------------------|---------------|
| Leader key configurable | `keys_test.go` — ParseKey with various combos |
| Chord capture during timeout | `engine_test.go` — state machine transitions |
| Mapped chord dispatches action | `engine_test.go` — dispatch callback |
| Unmapped chord / timeout exits silently | `engine_test.go` — timeout and unknown chord |
| Platform method reported at startup | `listener_test.go` — stub reports method |
| Conflict warning on common shortcuts | `keys_test.go` or `listener_test.go` — known conflict detection |
| Clean release on shutdown | `engine_test.go` — Stop() releases listener |
| Key suppression in chord wait | `engine_test.go` — engine state during chord wait |

## Session Log

<!-- Append after each session: date, completed, blockers -->
