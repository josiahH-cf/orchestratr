# Release Notes — v0.1.0 (Initial Feature Set)

## Summary

All core features for the orchestratr hotkey launcher have been implemented and merged. The daemon, hotkey engine, app registry, web GUI, and platform-specific layers are functional on Linux, macOS, and Windows.

## Features Shipped

### E2E Hotkey Launch Pipeline
Full end-to-end pipeline: daemon starts → hotkey engine captures Ctrl+Space chord → registry resolves the bound app → launcher spawns/focuses the process → API and GUI reflect state.

### Drop-In App Discovery (`apps.d/`)
Apps self-register by placing a YAML manifest in the `apps.d/` directory. Hot-reload watches for file additions/removals. Schema validation rejects malformed manifests with clear errors.

### Ready Command Health Polling
After launch, `ready_cmd` health checks poll until the app reports ready (exit 0) or timeout expires. Integrated into the launch flow with configurable `ready_timeout_ms`.

### Startup Diagnostic + Error Reporting
`orchestratr doctor` validates configuration, binary paths, hotkey conflicts, and platform capabilities. Error states propagate through the API and display as badges in the web GUI. Verbose mode streams unmatched chord events for debugging.

### Native System Tray
System tray icon via `fyne.io/systray` (pure Go D-Bus on Linux). Menu provides status, pause/resume hotkeys, open GUI, and quit. Headless fallback when no display is available.

### Windows Platform Layer
- Hotkey listener using `RegisterHotKey` + `WH_KEYBOARD_LL` hook
- Process executor with native/WSL environment routing
- Window focus via `EnumWindows` + `SetForegroundWindow`
- Autostart via Windows registry + WSL2 startup guard

## Platform Support

| Platform | Status |
|----------|--------|
| Linux (X11) | Fully functional |
| macOS | Fully functional |
| Windows | Fully functional |
| WSL2 | Fully functional (launched from Windows daemon) |

## Dependencies

- Go 1.22+
- `fyne.io/systray` v1.12.0
- `fsnotify` v1.9.0
- `golang.org/x/sys` v0.30.0
- `gopkg.in/yaml.v3` v3.0.1
