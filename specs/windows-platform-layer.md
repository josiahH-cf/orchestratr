# Feature: Windows Platform Layer (Hotkey Listener + Executor + Focus)

**Status:** Not started
**Project:** orchestratr

## Description

orchestratr is designed to run as a native Windows binary that captures system-wide hotkeys and launches apps — both native Windows apps and apps inside WSL2 distributions. Currently, the Windows hotkey listener, process executor, and window focus modules are all stubs returning `ErrNotImplemented`. This spec implements the full Windows platform layer so the daemon can fulfill its core promise: always listening, always ready, always supersedes the local key combo.

### Current State (Context for Implementation)

- **Hotkey listener** (`internal/hotkey/listener_windows.go`): Stub, returns `ErrNotImplemented`. The comment says "The full implementation will use RegisterHotKey". Build tag: `//go:build windows`.
- **Process executor** (`internal/launcher/executor_stub.go`): Stub for all non-Linux platforms (`//go:build !linux`), returns `ErrNotImplemented` for `Launch()`, `Kill()`, `IsRunning()`.
- **Window focus** (`internal/launcher/focus_stub.go`): Stub for non-Linux (`//go:build !linux`), returns `ErrFocusNotSupported`.
- **Linux executor** (`internal/launcher/native_linux.go`): Working implementation — spawns via `bash -c`, tracks PID, uses process groups. This is the reference for the Windows implementation.
- **Linux focus** (`internal/launcher/focus_linux.go`): Working implementation — uses `xdotool` to find window by PID and activate.
- **Environment field**: `AppEntry.Environment` is validated (`native`/`wsl`/`wsl:<distro>`) but **never consumed** by the launcher. `NativeExecutor.Launch()` runs `bash -c entry.Command` regardless of environment.
- **WSL2 detection** (`internal/platform/wsl.go`): `IsWSL2()` reads `/proc/version`, fully tested. `WSL2Warning()` returns a warning string. **Neither is called from `main.go`** — dead code today.
- **Hotkey engine** (`internal/hotkey/engine.go`): Platform-agnostic state machine. Calls `listener.GrabKey()` for leader key, `listener.GrabKeyboard()` for chord capture. On chord match, calls `OnAction(appName)`. Fully tested.
- **Autostart Windows** (`internal/autostart/windows.go`): Stub — `Enable()` and `Disable()` return `ErrNotImplemented`.

### Architecture

The hotkey engine is already abstracted behind the `Listener` interface:

```go
type Listener interface {
    GrabKey(combo KeyCombo) error
    GrabKeyboard() error
    UngrabKeyboard() error
    NextEvent(ctx context.Context) (Event, error)
    Close() error
    Info() PlatformInfo
}
```

The Windows implementation needs to satisfy this interface using Win32 APIs.

The launcher is abstracted behind the `Executor` interface:

```go
type Executor interface {
    Launch(entry registry.AppEntry) (int, error)  // returns PID
    Kill(pid int) error
    IsRunning(pid int) bool
}
```

The Windows implementation must handle environment-based routing (`native` vs `wsl`).

## Acceptance Criteria

- [ ] On Windows, `orchestratr start` registers the configured leader key as a system-wide hotkey via `RegisterHotKey`
- [ ] Pressing the leader key enters chord-wait state; the next keypress is captured via a low-level keyboard hook (`SetWindowsHookEx` with `WH_KEYBOARD_LL`)
- [ ] Chord match triggers `OnAction(appName)` through the existing engine
- [ ] Apps with `environment: native` (or empty) are launched via the Windows shell (`cmd.exe /c` or `CreateProcess`)
- [ ] Apps with `environment: wsl` are launched via `wsl.exe -d <default_distro> -- bash -lc "<command>"`
- [ ] Apps with `environment: wsl:<distro>` are launched via `wsl.exe -d <distro> -- bash -lc "<command>"`
- [ ] If `environment: wsl` without a specific distro, the default is resolved from `wsl --list --quiet` (first line)
- [ ] Running apps are focused via `SetForegroundWindow` (found by PID via `EnumWindows`) instead of re-launched
- [ ] If `orchestratr start` is run inside WSL2 (detected via `platform.IsWSL2()`), it prints the `WSL2Warning()` message and exits with a non-zero code, unless `--force` is passed
- [ ] PID tracking works for both native Windows processes and WSL2-launched processes (`wsl.exe` PID)
- [ ] Process exit detection works via `cmd.Wait()` goroutine (same pattern as Linux)
- [ ] Windows autostart via registry key `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`

## Affected Areas

| Area | Files |
|------|-------|
| **Replace** | `internal/hotkey/listener_windows.go` — full Win32 `RegisterHotKey` + `WH_KEYBOARD_LL` implementation |
| **Create** | `internal/launcher/native_windows.go` — `WindowsExecutor` with environment routing |
| **Create** | `internal/launcher/focus_windows.go` — `SetForegroundWindow` + `EnumWindows` by PID |
| **Remove** | `internal/launcher/executor_stub.go` — replaced by platform-specific files (keep for darwin only) |
| **Remove** | `internal/launcher/focus_stub.go` — replaced by platform-specific files (keep for darwin only) |
| **Modify** | `cmd/orchestratr/main.go` — WSL2 startup guard, `--force` flag, platform-appropriate executor selection |
| **Modify** | `internal/autostart/windows.go` — real registry implementation |

## Constraints

- All Win32 calls must use Go's `syscall` or `golang.org/x/sys/windows` — no CGo (unlike the Linux X11 listener which uses CGo). This keeps cross-compilation simple.
- `RegisterHotKey` registers with a hidden message-only window (standard pattern for Go services on Windows)
- The low-level keyboard hook (`WH_KEYBOARD_LL`) must be installed only during the chord-wait window (2s default) and then uninstalled — keeping it permanently would be intrusive and flagged by security software
- `bash -lc` is required for WSL2 commands so the user's profile (PATH, aliases, venv activation) is loaded
- Focus via `SetForegroundWindow` requires the foreground lock workaround (`AllowSetForegroundWindow` or `AttachThreadInput` dance) — document the chosen approach in a decision record if the standard workaround is insufficient
- Process group handling: native Windows apps should be launched in a new process group (`CREATE_NEW_PROCESS_GROUP`) for clean kill support

## Out of Scope

- macOS platform layer (separate spec if needed)
- Wayland/Linux platform improvements (covered by existing Decision 0002)
- Tray icon on Windows (currently headless; real tray is a separate feature)
- Windows installer / MSI packaging (the existing `install.ps1` is sufficient for now)

## Dependencies

- None — this can be implemented independently as it replaces stubs with real implementations

## Notes

### Win32 hotkey registration pattern

```
1. Create a hidden message-only window (CreateWindowEx with HWND_MESSAGE parent)
2. RegisterHotKey(hwnd, id, MOD_CONTROL, VK_SPACE) for leader key
3. Run a message pump (GetMessage loop) in a dedicated goroutine
4. On WM_HOTKEY: signal the engine that leader key was pressed
5. Engine calls GrabKeyboard() → install WH_KEYBOARD_LL hook
6. Hook captures next keypress → return as Event to engine
7. Engine calls UngrabKeyboard() → uninstall hook
8. On shutdown: UnregisterHotKey, DestroyWindow
```

### Environment routing in WindowsExecutor

```go
func (e *WindowsExecutor) Launch(entry registry.AppEntry) (int, error) {
    switch {
    case entry.Environment == "" || entry.Environment == "native":
        // cmd.exe /c "entry.Command"
        cmd := exec.Command("cmd.exe", "/c", entry.Command)
    case entry.Environment == "wsl":
        distro := resolveDefaultWSLDistro()
        cmd := exec.Command("wsl.exe", "-d", distro, "--", "bash", "-lc", entry.Command)
    case strings.HasPrefix(entry.Environment, "wsl:"):
        distro := strings.TrimPrefix(entry.Environment, "wsl:")
        cmd := exec.Command("wsl.exe", "-d", distro, "--", "bash", "-lc", entry.Command)
    }
    // SysProcAttr: CREATE_NEW_PROCESS_GROUP
    // Start, track PID, goroutine for exit detection
}
```

### WSL default distro resolution

```go
func resolveDefaultWSLDistro() (string, error) {
    out, err := exec.Command("wsl.exe", "--list", "--quiet").Output()
    // First non-empty line is the default distro
    // Cache for the lifetime of the daemon (invalidated on config reload)
}
```

### Focus flow

```
1. EnumWindows callback: for each top-level window, GetWindowThreadProcessId
2. If PID matches target → store HWND
3. AllowSetForegroundWindow(targetPID) (may need current foreground window's thread)
4. SetForegroundWindow(hwnd)
5. If window is minimized: ShowWindow(hwnd, SW_RESTORE) first
```

### Stub file restructuring

The current `executor_stub.go` has `//go:build !linux`. After this spec:
- `native_linux.go` — `//go:build linux` (exists)
- `native_windows.go` — `//go:build windows` (new)
- `executor_stub.go` — `//go:build !linux && !windows` (darwin/other)
- Same pattern for `focus_*.go` files
