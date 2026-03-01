# Feature: Cross-Platform Hotkey Engine

**Status:** Not started  
**Project:** orchestratr

## Description

The hotkey engine is the OS abstraction layer that registers the global leader key and captures chord keystrokes during the listening window. It must work across Windows, macOS, and Linux (X11 and Wayland) without requiring the user to install additional system software. The engine exposes a uniform API to the daemon regardless of platform.

## Acceptance Criteria

- [ ] Global leader key (default: Ctrl+Space) is captured on Windows via `RegisterHotKey` / low-level keyboard hook
- [ ] Global leader key is captured on macOS via `CGEventTap` or `NSEvent.addGlobalMonitorForEvents`
- [ ] Global leader key is captured on Linux/X11 via `XGrabKey`
- [ ] Global leader key is captured on Linux/Wayland via a supported mechanism (e.g., D-Bus portal, compositor protocol, or documented fallback)
- [ ] After leader activation, chord keystrokes are captured for the timeout window without leaking to the focused application
- [ ] The leader key is configurable via the app config (not hardcoded to Ctrl+Space)
- [ ] A warning is emitted at registration time if the leader key is known to conflict with common OS shortcuts
- [ ] Hotkey registrations are cleanly released on daemon pause, shutdown, or crash recovery
- [ ] The engine reports the current platform and registration method at startup (for diagnostics)

## Affected Areas

| Area | Files |
|------|-------|
| **Create** | `orchestratr/hotkey/engine.py` — platform-agnostic interface |
| **Create** | `orchestratr/hotkey/windows.py` — Win32 implementation |
| **Create** | `orchestratr/hotkey/macos.py` — macOS implementation |
| **Create** | `orchestratr/hotkey/linux_x11.py` — X11 implementation |
| **Create** | `orchestratr/hotkey/linux_wayland.py` — Wayland implementation (may be stub with documented limitations) |

## Constraints

- **No elevated privileges** for basic hotkey registration on Windows and Linux/X11
- macOS requires Accessibility permission — the engine must detect this and prompt the user (see installer spec)
- Wayland is the hardest target: most compositors don't allow global hotkey grabbing by design. The spec must document a pragmatic fallback strategy (e.g., compositor-specific D-Bus protocols for GNOME/KDE, or requiring X11 compatibility mode)
- Chord capture during the listening window must suppress key pass-through to avoid triggering actions in the focused app
- The engine must be testable in CI without a real display (mock/stub the OS layer)

## Out of Scope

- Registering per-app hotkeys (orchestratr registers only the leader key globally)
- Remapping or intercepting arbitrary key combinations system-wide
- Gamepad, mouse, or non-keyboard input
- Wayland compositor plugin development

## Dependencies

- `01-core-daemon.md` — the daemon owns the event loop that drives the engine

## Notes

### Wayland reality check

Wayland's security model intentionally prevents applications from grabbing global keys. Practical approaches:

1. **GNOME/Mutter:** `org.gnome.Shell.Extensions` D-Bus interface or a GNOME Shell extension
2. **KDE/KWin:** `org.kde.kglobalaccel` D-Bus interface (well-supported)
3. **Sway/wlroots:** Custom IPC or manual `swaymsg` bindings
4. **Fallback:** Document that on unsupported Wayland compositors, the user must configure the leader key in their compositor's native keybinding settings to run `orchestratr trigger` (a CLI command that simulates leader activation)

This may warrant a `/decisions/` document before implementation.

### Chord capture model

During the listening window (leader pressed, awaiting chord):

```
[Leader pressed] → engine enters CHORD_WAIT state
  → all keyboard input is captured (not forwarded)
  → if mapped chord received: dispatch action, exit state
  → if unmapped chord received: ignore, exit state
  → if timeout expires: exit state silently
```

The key suppression during CHORD_WAIT is critical — without it, pressing "e" after the leader would also type "e" into whatever app has focus.
