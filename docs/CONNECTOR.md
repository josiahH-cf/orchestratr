# Connecting Apps to orchestratr

This document explains how orchestratr discovers and launches apps, and how to build a connector so your app is launchable via hotkey.

## How It Works

orchestratr is a system-wide hotkey daemon. It captures a **leader key** (`Ctrl+Space` by default) followed by a single-character **chord** (e.g., `e`) to launch, focus, or health-check registered apps.

Apps register by writing a YAML manifest to orchestratr's **drop-in directory** (`apps.d/`). orchestratr scans this directory at startup and on config reload. No API calls, no runtime coupling — just one file.

## Quick Start

### 1. Write a manifest

Create `<appname>.yml` in orchestratr's `apps.d/` directory:

```yaml
# apps.d/myapp.yml
name: myapp
chord: "m"
command: "myapp gui"
environment: native
description: "My application"
ready_cmd: "myapp status --json"
ready_timeout_ms: 5000
```

### 2. Verify

```bash
orchestratr list          # should show your app
# Ctrl+Space → m          # launch it
```

That's it for manual registration. For automated registration (recommended), read on.

## Manifest Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `name` | string | **yes** | — | Unique app identifier |
| `chord` | string | **yes** | — | Hotkey chord (single char or key name like `f1`) |
| `command` | string | **yes** | — | Shell command to launch the app |
| `environment` | string | no | `native` | `native`, `wsl`, or `wsl:<distro>` |
| `description` | string | no | — | Human-readable description |
| `ready_cmd` | string | no | — | Health check command (exit 0 = ready) |
| `ready_timeout_ms` | int | no | `5000` | Max polling time for `ready_cmd` |
| `detached` | bool | no | `false` | Skip PID tracking (for UWP apps etc.) |

### Reserved Chords

- `?` — help overlay
- `space` — cancel

### Environment Values

- `native` — run directly on the host OS
- `wsl` — run inside the default WSL2 distribution (`wsl.exe -d <default> -- bash -lc "<command>"`)
- `wsl:<distro>` — target a specific distro (e.g., `wsl:Ubuntu-22.04`)

**Important:** The `command` field is always the **native** command. Don't pre-wrap with `wsl.exe` — orchestratr handles environment routing.

## Drop-In Directory Location

| Platform | Path |
|----------|------|
| Linux | `~/.config/orchestratr/apps.d/` |
| macOS | `~/Library/Application Support/orchestratr/apps.d/` |
| Windows | `%APPDATA%\orchestratr\apps.d\` |

From WSL2 (targeting Windows-side orchestratr): `/mnt/c/Users/<username>/AppData/Roaming/orchestratr/apps.d/`

## Merge Rules

- `config.yml` entries take precedence over `apps.d/` entries on name or chord conflict
- Conflicts between two `apps.d/` files are rejected with an error naming both files
- `apps.d/` files are read-only from orchestratr's perspective — the owning app manages them

## Health Check Protocol (`ready_cmd`)

After launching an app, orchestratr polls its `ready_cmd` at 500ms intervals:

- **Exit code 0** → app is ready (polling stops)
- **Non-zero exit code** → not ready yet (retry)
- **Timeout** → app stays in "launched" state with a warning

The recommended pattern: implement `<appname> status --json` that returns:

```json
{
  "version": "1.0.0",
  "status": "ok",
  "config_dir": "/path/to/config"
}
```

## Building an Automated Connector

For apps that should self-register, implement a connector module. The full protocol with code examples is documented in the workspace reference: `/specs/orchestratr-app-connector-protocol.md`.

### Summary

1. **Module**: `<app>/integrations/orchestratr.py` (or equivalent)
2. **Functions**: `generate_manifest()`, `manifest_needs_update()`, `get_status_json()`, `resolve_orchestratr_apps_dir()`
3. **CLI**: `<app> status --json` — health check endpoint
4. **Setup hook**: Call `generate_manifest()` during app setup/install
5. **GUI hint**: On startup, check `manifest_needs_update()` and show a non-intrusive hint if stale
6. **Passive**: If orchestratr isn't installed, do nothing — no errors, no side effects

### Reference Implementations

- **espansr**: `espansr/espansr/integrations/orchestratr.py` (complete)
- **templatr**: `templatr/templatr/integrations/orchestratr.py` (spec: `templatr/specs/orchestratr-connector.md`)

## Troubleshooting

```bash
# Is the daemon running?
orchestratr status

# See registered apps and their source
orchestratr list

# Check a specific manifest
cat ~/.config/orchestratr/apps.d/myapp.yml

# Full diagnostic (planned — see specs/startup-diagnostic.md)
# orchestratr doctor

# Watch events in real-time (planned — see specs/startup-diagnostic.md)
# orchestratr start --verbose
```

## Architecture

```
                    Windows Desktop
                    ┌──────────────────────────────┐
                    │  orchestratr daemon           │
                    │  (native Windows binary)      │
                    │                               │
                    │  RegisterHotKey (leader key)  │
                    │  WH_KEYBOARD_LL (chord)       │
                    │                               │
                    │  ┌─────────────────────────┐  │
                    │  │ Registry                 │  │
                    │  │  config.yml              │  │
                    │  │  apps.d/espansr.yml      │  │
                    │  │  apps.d/templatr.yml     │  │
                    │  └─────────────────────────┘  │
                    │                               │
                    │  on chord match:              │
                    │   env=native → cmd.exe /c     │
                    │   env=wsl    → wsl.exe -d ... │
                    └──────────────────────────────┘
                              │
                    ┌─────────┴──────────┐
                    │                    │
              native apps          WSL2 apps
              (cmd.exe /c)     (wsl.exe -d distro)
                                       │
                              ┌────────┴────────┐
                              │ WSL2 (Ubuntu)    │
                              │  espansr gui     │
                              │  templatr        │
                              │  (WSLg display)  │
                              └─────────────────┘
```
