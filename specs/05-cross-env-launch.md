# Feature: Cross-Environment Launching

**Status:** Not started  
**Project:** orchestratr

## Description

orchestratr must launch apps regardless of where they live — natively on the host OS, or inside a WSL2 distribution when the daemon runs on Windows. It must also track whether a launched app is still running so it can bring an existing instance to the foreground instead of spawning duplicates. This spec covers the process spawning, environment bridging, and instance tracking mechanics.

## Acceptance Criteria

- [ ] Apps with `environment: native` are launched directly via the host shell
- [ ] Apps with `environment: wsl` are launched via `wsl.exe -d <distro> -- <command>` when the daemon runs on Windows
- [ ] Apps with `environment: wsl` are launched directly (as native) when the daemon itself runs on Linux/WSL2
- [ ] A specific WSL distro can be targeted via `environment: wsl:Ubuntu-22.04` syntax
- [ ] Before launching, orchestratr checks if the app is already running; if so, it attempts to bring the existing window to the foreground
- [ ] If bring-to-front fails or is unsupported, orchestratr falls back to launching a new instance
- [ ] Launch failures (command not found, permission denied, WSL not available) produce a tray notification with the error message
- [ ] Process tracking uses PID monitoring — no polling; the daemon is notified when a process exits
- [ ] `orchestratr list` shows running/stopped status for each registered app

## Affected Areas

| Area | Files |
|------|-------|
| **Create** | `orchestratr/launcher.py` — process spawning, environment detection, PID tracking |
| **Create** | `orchestratr/launcher/native.py` — native OS process spawning |
| **Create** | `orchestratr/launcher/wsl.py` — WSL2 bridge logic |
| **Create** | `orchestratr/window.py` — bring-to-front logic per platform |

## Constraints

- Bring-to-front is best-effort: it works reliably on Windows (`SetForegroundWindow`), partially on macOS (`NSRunningApplication.activateWithOptions`), and is compositor-dependent on Linux
- WSL2 path translation may be needed for commands that reference Linux paths from the Windows side
- Process tracking must handle zombie/orphan processes gracefully
- App commands may include pipes, environment variables, or shell features — they must be executed via a shell (not `exec` directly)
- No assumption about the app's technology stack — any executable that can be run from a shell command is valid

## Out of Scope

- Restarting crashed apps automatically (no watchdog behavior)
- Resource monitoring (CPU, memory) of launched apps
- Launching apps on remote machines or containers
- Managing multiple instances of the same app (one instance per registered app)

## Dependencies

- `01-core-daemon.md` — daemon dispatches launch commands
- `03-app-registry.md` — registry provides command, environment, and chord mapping
- `02-hotkey-engine.md` — chord triggers the launch

## Notes

### Launch flow

```
Chord received → look up app in registry
  → is app already running? (PID check)
    → YES: attempt bring-to-front
      → success: done
      → fail: log warning, done (don't launch duplicate)
    → NO: spawn process
      → environment == native: spawn directly
      → environment == wsl: wrap in wsl.exe
      → store PID for tracking
      → emit launch event to HTTP API
```

### WSL2 bridging details

When the daemon runs on Windows and needs to launch a WSL app:

```
wsl.exe -d Ubuntu -- bash -lc "/home/user/app/.venv/bin/myapp gui"
```

Key considerations:
- `bash -lc` ensures the user's profile is loaded (PATH, aliases, etc.)
- GUI apps in WSL2 require WSLg (Windows 11) or an X server — orchestratr should detect and warn if display forwarding isn't available
- The default distro is read from `wsl --list --quiet` if `environment: wsl` doesn't specify one

### Bring-to-front platforms

| Platform | Method | Reliability |
|----------|--------|-------------|
| Windows | `SetForegroundWindow` via Win32 API | High (with foreground lock workaround) |
| macOS | `NSRunningApplication.activateWithOptions` | High |
| Linux/X11 | `_NET_ACTIVE_WINDOW` via `xdotool` or Xlib | Medium |
| Linux/Wayland | Compositor-specific D-Bus (limited) | Low — document as best-effort |
