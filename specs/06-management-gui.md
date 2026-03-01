# Feature: Management GUI

**Status:** Not started  
**Project:** orchestratr

## Description

A minimal GUI for managing the app registry and orchestratr settings. Used rarely — primarily during initial setup and when adding new apps. Accessed via the system tray "Configure" menu item or a dedicated chord (e.g., leader + `?`). The GUI reads and writes the same YAML config file that can also be edited by hand.

## Acceptance Criteria

- [ ] GUI opens from the system tray "Configure" action or via the help chord (leader + `?`)
- [ ] Displays a table of registered apps with columns: Name, Chord, Command, Environment, Status (running/stopped)
- [ ] "Add App" form with fields: name, chord key, command, environment (dropdown: native/wsl/wsl:distro), description
- [ ] "Edit" and "Remove" actions on each app row
- [ ] Chord key field validates uniqueness in real-time (rejects duplicates before save)
- [ ] "Leader Key" setting is editable with a key-capture input (press the desired key combo)
- [ ] "Save" writes to the YAML config and triggers hot-reload in the daemon
- [ ] GUI is a separate window (not embedded in the tray) — closes without affecting the daemon

## Affected Areas

| Area | Files |
|------|-------|
| **Create** | `orchestratr/gui/main_window.py` — app table, settings panel |
| **Create** | `orchestratr/gui/app_form.py` — add/edit app dialog |
| **Create** | `orchestratr/gui/key_capture.py` — custom key-capture widget for leader key setting |

## Constraints

- Must match the same cross-platform requirements as the daemon (Windows, macOS, Linux)
- GUI framework should be the same as the tray implementation to avoid pulling in a second toolkit
- Window size target: single compact panel, no tabs or multi-screen navigation
- GUI does not need to be responsive to real-time app status changes (user can reload manually)
- Must be functional without the daemon running (can edit config file standalone)

## Out of Scope

- Drag-and-drop reordering of apps
- Visual hotkey conflict checker (shows conflicts with system-wide shortcuts outside orchestratr)
- App-specific settings or sub-configuration panels
- Theming or appearance customization
- Real-time log viewer

## Dependencies

- `03-app-registry.md` — GUI reads/writes the same config format
- `01-core-daemon.md` — tray icon triggers the GUI

## Notes

### Design philosophy

The GUI is a **config editor with guardrails**, not an app dashboard. It should feel like editing a YAML file with validation assistance. Power users will prefer editing the YAML directly; the GUI exists for initial setup and for users who don't want to touch config files.

### Layout sketch

```
┌─────────────────────────────────────────────────────┐
│ orchestratr — Configuration                        [×] │
├─────────────────────────────────────────────────────┤
│ Leader Key: [Ctrl+Space]  Timeout: [2000]ms         │
├─────────────────────────────────────────────────────┤
│ Registered Apps                                      │
│ ┌───────┬───────┬─────────────────┬─────────┬──────┐│
│ │ Chord │ Name  │ Command         │ Env     │      ││
│ ├───────┼───────┼─────────────────┼─────────┼──────┤│
│ │   e   │espansr│ espansr gui     │ wsl     │ ✎ ✕  ││
│ │   a   │app-a  │ app-a.exe       │ native  │ ✎ ✕  ││
│ └───────┴───────┴─────────────────┴─────────┴──────┘│
│                                      [+ Add App]     │
├─────────────────────────────────────────────────────┤
│ [Save]  [Cancel]                                     │
└─────────────────────────────────────────────────────┘
```

### Key capture widget

For the leader key setting, a custom input widget that:
1. Shows the current key combo as text (e.g., "Ctrl+Space")
2. On click, enters capture mode ("Press a key combo...")
3. Records the next key event and displays it
4. Validates it's a modifier + key combination (not a bare letter)
