# Tasks: Native System Tray

**Spec:** /specs/system-tray.md
**Branch:** `feat/system-tray`

## Status

- Total: 4
- Complete: 4
- Remaining: 0

## Task List

### Task 1: Add fyne.io/systray dependency + provider dispatch

- **Files:** `go.mod`, `go.sum`, `internal/tray/provider.go`, `internal/tray/provider_stub.go`
- **Done when:** `fyne.io/systray` is in `go.mod`; `NewPlatformProvider()` is defined in both the real file (with build tag `(linux || windows) && !notray`) and the stub (build tag `!(linux || windows) || notray`); `go build ./...` and `go build -tags notray ./...` both pass; stub returns `HeadlessProvider`; real file returns a `SystrayProvider` instance (not yet functional — that's Task 2)
- **Criteria covered:** `go build -tags notray` succeeds; Darwin returns HeadlessProvider
- **Status:** [x] Complete

### Task 2: SystrayProvider — Setup, menu, state, quit

- **Files:** `internal/tray/systray.go`, `internal/tray/systray_test.go`
- **Done when:** `SystrayProvider.Setup()` starts `systray.Run()` in a goroutine, waits for `onReady` signal (3s timeout), builds the Pause/Resume/Configure/Quit menu; `SetState()` updates tooltip; `OnPause`/`OnResume`/`OnQuit`/`OnConfigure` store callbacks that menu items invoke; `Quit()` calls `systray.Quit()`; unit tests verify: interface compliance, callback registration, state tracking, idempotent `Quit()`, and `Setup()` timeout when display unavailable
- **Criteria covered:** Tray icon on Linux/Windows; tooltip reflects state; menu callbacks fire; Quit stops daemon
- **Status:** [x] Complete

### Task 3: NotifyError — Linux notify-send + Windows tooltip flash

- **Files:** `internal/tray/systray.go` (extend), `internal/tray/systray_test.go` (extend), `internal/tray/notify_linux.go`, `internal/tray/notify_windows.go`
- **Done when:** On Linux, `NotifyError` execs `notify-send --urgency=critical` with a 3s context timeout, best-effort (failure is logged, not returned); on Windows, `NotifyError` briefly sets the tooltip to the error message then restores the previous state after 5s in a goroutine; tests verify no panic when `notify-send` is absent and that the state is restored after the Windows flash
- **Criteria covered:** NotifyError produces visible notification on Linux and Windows
- **Status:** [x] Complete

### Task 4: Wire into main.go + install.sh documentation

- **Files:** `cmd/orchestratr/main.go`, `install.sh`
- **Done when:** `main.go` line 307 uses `tray.NewPlatformProvider()` instead of `&tray.HeadlessProvider{}`; if `Setup()` returns an error, it falls back to `HeadlessProvider` with a log warning; `install.sh` notes the `libappindicator3-dev` compile-time requirement for Linux; `go test ./cmd/orchestratr/...` passes
- **Criteria covered:** Tray appears on startup; graceful fallback on headless; Linux compile deps documented
- **Status:** [x] Complete

## Test Strategy

| Criterion | Tested in |
|-----------|-----------|
| Darwin returns HeadlessProvider | Task 1 — provider_stub.go compile-time type check |
| `-tags notray` builds cleanly | Task 1 — CI build verification |
| Interface compliance | Task 2 — `var _ Provider = (*SystrayProvider)(nil)` |
| Callbacks registered and called | Task 2 — `systray_test.go` with mock menu items |
| SetState updates tooltip text | Task 2 — `systray_test.go` internal state check |
| NotifyError no-op when notify-send absent | Task 3 — `notify_linux.go` test with PATH override |
| Windows error flash restores state | Task 3 — `notify_windows.go` test |
| Graceful fallback for headless | Task 4 — `main_test.go` verifies no crash on headless |

## Session Log

<!-- Append after each session: date, completed, blockers -->

- 2025-07: Tasks 1-4 complete. Added fyne.io/systray v1.12.0 (pure Go D-Bus on Linux); provider dispatch with display detection; SystrayProvider with injectable runFn/quitFn (no real display needed in tests); NotifyError via notify-send (Linux) and tooltip flash (Windows); wired main.go to use NewPlatformProvider() with HeadlessProvider fallback. All 17 tray tests pass; vet clean on Linux + Windows; -tags notray verified.