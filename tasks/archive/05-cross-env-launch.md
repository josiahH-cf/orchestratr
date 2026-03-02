# Tasks: 05-cross-env-launch

**Spec:** /specs/05-cross-env-launch.md

## Status

- Total: 5
- Complete: 4
- Remaining: 1

## Task List

### Task 1: Executor interface, NativeExecutor, and stub

- **Files:** `internal/launcher/launcher.go`, `internal/launcher/native_linux.go`, `internal/launcher/executor_stub.go`
- **Done when:** `Executor` interface defined, `NativeExecutor` launches via `bash -c`, tracks PIDs, detects exit via `cmd.Wait()` goroutine (no polling), `StubExecutor` returns `ErrNotImplemented` on non-Linux
- **Criteria covered:** native launch, process tracking via PID, no-polling exit detection
- **Status:** [x] Complete

### Task 2: Launcher tests

- **Files:** `internal/launcher/launcher_test.go`, `internal/launcher/native_linux_test.go`
- **Done when:** Tests cover: launch + IsRunning, already-running guard, stop, stop-not-running, natural exit with callback, exit-with-error callback, bad command, StopAll, relaunch after exit
- **Criteria covered:** all Task 1 criteria via automated tests
- **Status:** [x] Complete

### Task 3: Wire launcher into daemon and update list

- **Files:** `cmd/orchestratr/main.go`
- **Done when:** `OnAction` callback calls `launchApp()`, exit callback updates `StateTracker`, `orchestratr list` shows STATUS column when daemon is running
- **Criteria covered:** chord dispatch launches apps, `orchestratr list` shows running/stopped
- **Status:** [x] Complete

### Task 4: WSL bridge launcher (Windows)

- **Files:** `internal/launcher/wsl_windows.go`, `internal/launcher/wsl_windows_test.go`
- **Done when:** `environment: wsl` apps launch via `wsl.exe -d <distro> -- bash -lc <command>`, default distro detection, `wsl:Ubuntu-22.04` syntax supported
- **Criteria covered:** wsl environment, distro targeting, WSL on Linux falls back to native
- **Status:** [ ] Not started — deferred (requires Windows)

### Task 5: Bring-to-front and tray notifications

- **Files:** `internal/launcher/focus_linux.go`, `internal/launcher/focus_stub.go`, `internal/tray/tray.go`, `internal/launcher/launcher.go`, `internal/launcher/native_linux.go`, `internal/launcher/executor_stub.go`, `cmd/orchestratr/main.go`
- **Done when:** Before launching, executor attempts bring-to-front for running apps; launch failures produce tray notification
- **Criteria covered:** bring-to-front, tray notification on failure
- **Status:** [x] Complete

## Test Strategy

| Criterion | Tested by |
|-----------|-----------|
| Native launch via host shell | Task 2: `TestLaunchAndIsRunning` |
| Already-running guard | Task 2: `TestLaunchAlreadyRunning` |
| PID tracking, no polling | Task 2: `TestNaturalExit`, `TestNaturalExitWithError` |
| Stop + StopAll | Task 2: `TestStopRunningApp`, `TestStopAll` |
| Relaunch after exit | Task 2: `TestRelaunchAfterExit` |
| Bad command error | Task 2: `TestLaunchBadCommand`, `TestLaunchEmptyCommand` |
| Bring-to-front attempt | Task 5: `TestFocusWindow_PIDZero`, `TestFocusWindow_NegativePID`, `TestFocusWindow_NonexistentPID`, `TestFocusWindow_NoXdotool` |
| PID lookup for running app | Task 5: `TestPIDRunningApp`, `TestPIDNotRunning` |
| Tray notification on failure | Task 5: `TestHeadlessProvider_NotifyError`, `TestHeadlessProvider_NotifyError_Multiple` |
| `list` shows running/stopped | Task 3: existing `TestRun_ListWithApps` (offline), manual (online) |

## Session Log

- **2026-03-01:** Completed task 5 (bring-to-front + tray notifications). Added `FocusWindow(pid)` function (xdotool-based on Linux, stub on other platforms). Added `PID(name)` method to `Executor` interface. Added `NotifyError(title, message)` to `tray.Provider` interface with `HeadlessProvider` recording for tests. Updated `launchApp` flow: if app running → attempt focus → log result; if launch fails → tray notification. 8 new tests, all passing. Full suite green, `go vet` clean, `gofmt` clean. Task 4 (WSL) remains deferred (requires Windows).
