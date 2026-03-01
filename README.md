# orchestratr

A system-wide hotkey launcher and app orchestrator.

Press a leader key chord to instantly launch, focus, or cycle between your internal tools — across native apps, WSL2, and terminal sessions.

## Status

**Pre-development** — specs complete, implementation not started.

## Architecture

- **Daemon:** Background process with system tray, single-instance lock
- **Hotkey Engine:** Leader key (Ctrl+Space) + chord (single keypress) to trigger actions
- **App Registry:** YAML config defining apps, their launch commands, and health checks
- **HTTP API:** Localhost JSON API on port 9876 for programmatic control
- **Cross-Environment:** Launch native and WSL2 apps from the same daemon

## Specs

See `/specs/` for the complete design:

| # | Spec | Description |
|---|------|-------------|
| 01 | Core Daemon | Background process, system tray, lifecycle |
| 02 | Hotkey Engine | Leader key + chord capture, per-platform |
| 03 | App Registry | YAML config, hot-reload, app state |
| 04 | HTTP API | Localhost REST endpoints |
| 05 | Cross-Environment Launch | Native + WSL2 bridging |
| 06 | Management GUI | Minimal config editor |
| 07 | Installer & Permissions | Autostart, OS permissions, uninstall |

## Platforms

- Linux (X11 + Wayland)
- macOS
- Windows
- WSL2 (apps launched from Windows-side daemon)

## Related

- [espansr](https://github.com/josiahH-cf/espansr) — First app to be orchestrated. See its `specs/espansr-orchestratr-connector.md` for the integration spec.

## License

[Choose license]
