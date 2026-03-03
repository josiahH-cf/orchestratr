# orchestratr

A system-wide hotkey launcher and app orchestrator.

Press a leader key chord to instantly launch, focus, or cycle between your internal tools — across native apps, WSL2, and terminal sessions.

## Status

**In development** — core daemon, hotkey engine, app registry, HTTP API, and web GUI are implemented for Linux/X11. Windows platform layer and drop-in app discovery are in progress.

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

## Connecting Apps

Apps register with orchestratr by writing a YAML manifest to the `apps.d/` drop-in directory. See **[docs/CONNECTOR.md](docs/CONNECTOR.md)** for the full guide — manifest schema, path resolution, health checks, and examples.

For the complete protocol reference (including cross-platform path resolution and code templates): see the workspace-level `/specs/orchestratr-app-connector-protocol.md`.

## Related

- [espansr](https://github.com/josiahH-cf/espansr) — Espanso template manager. Connector: `espansr/integrations/orchestratr.py`
- [templatr](https://github.com/josiahH-cf/templatr) — Local-model prompt optimizer. Connector: `templatr/integrations/orchestratr.py`

## License

[Choose license]
