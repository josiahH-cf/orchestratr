# Tasks: Windows Platform Layer

**Spec:** /specs/windows-platform-layer.md

## Status

- Total: 4
- Complete: 4
- Remaining: 0

## Task List

### Task 1: Windows hotkey listener

- **Files:** `internal/hotkey/listener_windows.go`, `internal/hotkey/keys_windows.go`, `internal/hotkey/listener_windows_test.go`
- **Done when:** `NewPlatformListener()` on Windows returns a `WindowsListener` that registers a leader key via `RegisterHotKey`, captures chords via `WH_KEYBOARD_LL` hook, and passes events through the engine's `Start`/`GrabKeyboard`/`UngrabKeyboard`/`Stop` lifecycle. Unit tests verify key mapping, `PlatformInfo`, and the register/stop lifecycle.
- **Criteria covered:** Leader key registration, chord capture via low-level hook, engine integration via `OnAction`
- **Status:** [x] Complete

### Task 2: Windows executor with environment routing

- **Files:** `internal/launcher/native_windows.go`, `internal/launcher/native_windows_test.go`, `internal/launcher/executor_stub.go` (narrow build tag)
- **Done when:** `WindowsExecutor` launches native apps via `cmd.exe /c`, WSL apps via `wsl.exe -d <distro> -- bash -lc`, resolves default WSL distro from `wsl --list --quiet`, tracks PIDs, fires exit callbacks, and uses `CREATE_NEW_PROCESS_GROUP`. Stub build tag narrowed to `!linux && !windows`. Unit tests cover native launch, WSL routing, distro resolution, PID tracking, exit callback, and stop.
- **Criteria covered:** Native launch, WSL launch (default + named distro), PID tracking, exit detection, process group
- **Status:** [x] Complete

### Task 3: Windows window focus

- **Files:** `internal/launcher/focus_windows.go`, `internal/launcher/focus_windows_test.go`, `internal/launcher/focus_stub.go` (narrow build tag)
- **Done when:** `FocusWindow(pid)` uses `EnumWindows` + `GetWindowThreadProcessId` to find the target HWND, calls `ShowWindow(SW_RESTORE)` if minimized, then `SetForegroundWindow` with `AttachThreadInput` workaround. Stub build tag narrowed to `!linux && !windows`. Unit tests cover invalid PID and no-window error paths.
- **Criteria covered:** `SetForegroundWindow` focus by PID via `EnumWindows`
- **Status:** [x] Complete

### Task 4: Autostart registry + WSL2 startup guard

- **Files:** `internal/autostart/windows.go`, `internal/autostart/windows_test.go`, `cmd/orchestratr/main.go`, `cmd/orchestratr/main_test.go`
- **Done when:** `WindowsManager` uses `golang.org/x/sys/windows/registry` for real `HKCU\...\Run` writes (build-tagged `windows`). WSL2 startup guard in `main.go` prints `WSL2Warning()` and exits non-zero when `platform.IsWSL2()` is true, unless `--force` is passed. Tests cover autostart round-trip and WSL2 guard behavior.
- **Criteria covered:** Autostart via registry, WSL2 startup guard, `--force` flag
- **Status:** [x] Complete

## Test Strategy

| Criterion | Tested in |
|-----------|-----------|
| Leader key via RegisterHotKey | Task 1 — `listener_windows_test.go` |
| Chord capture via WH_KEYBOARD_LL | Task 1 — `listener_windows_test.go` |
| Native app launch (cmd.exe /c) | Task 2 — `native_windows_test.go` |
| WSL launch (wsl.exe -d) | Task 2 — `native_windows_test.go` |
| Default WSL distro resolution | Task 2 — `native_windows_test.go` |
| PID tracking + exit callback | Task 2 — `native_windows_test.go` |
| Focus via SetForegroundWindow | Task 3 — `focus_windows_test.go` (error paths; happy path is manual) |
| Autostart registry write | Task 4 — `windows_test.go` |
| WSL2 startup guard + --force | Task 4 — `main_test.go` |

## Session Log

<!-- Append after each session: date, completed, blockers -->
