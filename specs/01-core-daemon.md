# Feature: Core Daemon & System Tray

**Status:** Not started  
**Project:** orchestratr

## Description

orchestratr runs as a lightweight background daemon with a system tray presence. It listens for a global leader key (default: Ctrl+Space), intercepts the subsequent chord keystroke, and dispatches the mapped action (launch app, bring to front, or send IPC message). The daemon starts at login and persists silently until the user interacts with it via the tray icon or a hotkey chord.

## Acceptance Criteria

- [ ] Daemon starts as a background process with no visible window
- [ ] System tray icon appears with a context menu (Configure, Pause, Quit)
- [ ] Leader key (Ctrl+Space) activates a "listening" state with a brief visual indicator (e.g., tray icon change or small overlay)
- [ ] If no chord key is pressed within a configurable timeout (default: 2s), listening state cancels silently
- [ ] Daemon logs startup, shutdown, and hotkey events to a rotating log file
- [ ] Daemon exposes a localhost HTTP health endpoint (`GET /health`) returning `{"status": "ok"}`
- [ ] Daemon gracefully shuts down on SIGTERM / tray Quit, releasing all hotkey registrations
- [ ] "Pause" mode temporarily disables all hotkey listening without killing the daemon

## Affected Areas

| Area | Files |
|------|-------|
| **Create** | `orchestratr/daemon.py` — main event loop, signal handling, lifecycle |
| **Create** | `orchestratr/tray.py` — system tray icon, context menu |
| **Create** | `orchestratr/config.py` — configuration loading, defaults |
| **Create** | `orchestratr/__main__.py` — CLI entry point (`orchestratr start`, `orchestratr stop`, `orchestratr status`) |

## Constraints

- Must work on Windows (native), macOS (native), and Linux (X11 and Wayland)
- Single-instance enforcement — starting a second daemon shows an error or brings the config GUI forward
- Memory footprint target: < 30 MB resident at idle
- No root/admin privileges required for basic operation (see installer spec for elevated scenarios)
- Language/framework choice must support system tray on all three platforms

## Out of Scope

- App registration and configuration (see `app-registry.md`)
- Hotkey registration mechanics (see `hotkey-engine.md`)
- HTTP API endpoints beyond `/health` (see `http-api.md`)
- Cross-environment launching (see `cross-env-launch.md`)
- Health dashboard or monitoring UI

## Dependencies

- `hotkey-engine.md` — leader key registration
- `app-registry.md` — config format for registered apps

## Notes

### Technology recommendation

Python with a Qt-based tray (PyQt6 or PySide6) is viable but adds ~80 MB to the install footprint. Alternatives:

- **Go + systray** — single binary, cross-platform tray, low memory, excellent for daemons. Apps connect via HTTP (language-agnostic). Would require a separate GUI framework for the management screen.
- **Rust + tauri** — smallest binary, native webview for GUI, but steeper learning curve.
- **Python + pystray** — lighter than Qt, pure Python tray support, but pystray has rough edges on Wayland.

The choice should prioritize: OS agnosticism > binary simplicity > developer familiarity. **Go is the recommended baseline** — it produces a single cross-compiled binary with no runtime dependencies, and the HTTP API is idiomatic.

This decision should be formalized in a `/decisions/` doc before implementation begins.

### Leader key model

The leader key pattern (inspired by tmux, Vim, i3wm) avoids stealing common shortcuts:

1. User presses **Ctrl+Space** (leader)
2. Tray icon flashes or a tiny overlay appears — "orchestratr listening..."
3. User presses **e** within 2 seconds
4. orchestratr dispatches the action mapped to chord "e" (e.g., launch espansr)

If the user presses an unmapped key, orchestratr ignores it and cancels listening mode. The leader key itself (Ctrl+Space) is the only system-wide hotkey orchestratr registers — all chords are handled internally during the listening window, avoiding broad hotkey hijacking.
