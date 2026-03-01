# Feature: App Registry & Configuration

**Status:** Not started  
**Project:** orchestratr

## Description

The app registry is the central data store mapping chord keys to applications. It defines how apps are registered, what metadata orchestratr stores about each one, and where the configuration lives on disk. The config format must be human-readable, editable by hand, and usable by the management GUI.

## Acceptance Criteria

- [ ] Configuration is stored in a single YAML file at the platform-appropriate config directory (e.g., `~/.config/orchestratr/config.yml`, `%APPDATA%/orchestratr/config.yml`)
- [ ] Each app entry has: `name`, `chord` (single key), `command` (shell string), `environment` (native/wsl), and optional `description`
- [ ] Duplicate chord assignments are rejected with a clear error message at load time
- [ ] Config changes are detected and hot-reloaded without restarting the daemon (file watcher or explicit reload via tray menu)
- [ ] A `orchestratr list` CLI command prints the current app registry in a readable table
- [ ] Config file is created with sensible defaults on first run if it doesn't exist
- [ ] Invalid config entries (missing required fields, bad YAML) produce clear error messages and do not crash the daemon

## Affected Areas

| Area | Files |
|------|-------|
| **Create** | `orchestratr/registry.py` — app registry data model, load/save/validate |
| **Create** | `orchestratr/watcher.py` — file watcher for config hot-reload |
| **Modify** | `orchestratr/config.py` — integrate registry into global config |

## Constraints

- YAML format (human-readable, widely supported across languages)
- Config file must be valid even with zero app entries (daemon starts with empty registry)
- Chord keys are single characters or well-known key names (e.g., `e`, `s`, `1`, `f1`)
- Reserved chords: `?` (show help overlay), `space` (cancel) — cannot be assigned to apps
- No app-specific secrets in the config file

## Out of Scope

- GUI for editing the registry (see `management-gui.md`)
- App health monitoring or status tracking
- App-to-app routing rules (v2 IPC concern)
- Encrypted or remote config storage

## Dependencies

- `01-core-daemon.md` — daemon loads the registry at startup

## Notes

### Example config

```yaml
# ~/.config/orchestratr/config.yml
orchestratr:
  leader_key: "ctrl+space"
  chord_timeout_ms: 2000
  log_level: info

apps:
  - name: espansr
    chord: e
    command: "/home/josiah/R/espansr/.venv/bin/espansr gui"
    environment: wsl
    description: "Espanso template manager"

  - name: another-app
    chord: a
    command: "C:\\Tools\\another-app.exe"
    environment: native
    description: "Internal tool A"
```

### Environment field

The `environment` field tells orchestratr how to launch the app:

- `native` — run the command directly on the host OS
- `wsl` — wrap the command in `wsl.exe -d <distro> -- <command>` (Windows host only; on Linux, treated as `native`)
- `wsl:<distro>` — target a specific WSL distribution (e.g., `wsl:Ubuntu-22.04`)

This field is critical for the cross-environment launch spec.

### Hot-reload behavior

When the config file changes:
1. Daemon detects the change (filesystem watcher or manual reload via tray)
2. New config is parsed and validated
3. If valid: registry is swapped atomically, log entry written
4. If invalid: old config remains active, warning logged and shown via tray notification
