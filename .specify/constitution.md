<!-- This file is generated during Phase 2 (Compass). During Phase 2, this is the primary write target — Compass directly populates sections based on the discovery interview. After Phase 2 completes, this file is read-only. Post-Compass edits require /compass-edit or equivalent. -->

# Project Constitution

This document defines the project's identity, goals, and boundaries. It is the source of truth that every downstream phase references. Nothing gets built that isn't traceable to this document.

The sections below are guiding themes — not a rigid checklist. Compass populates them with depth proportional to the project's needs. Some projects need extensive security coverage; others need minimal. The interview adapts.

## Problem & Context

Developers and power users build or collect small tools, scripts, and utilities across projects. Launching them is friction-heavy: remember the path, open a terminal, type the command, manage windows. There's no unified, cross-platform way to bind a global hotkey to "launch this tool right now." Without orchestratr, users fall back to OS-specific hacks (AutoHotkey, Hammerspoon, shell aliases) that don't transfer across machines or operating systems.

## Target User

The primary user is a developer or technical power user who:
- Works across Linux, macOS, and Windows (including WSL)
- Builds their own tools or uses tools from other projects
- Wants instant access to those tools via keyboard shortcuts
- Prefers GUI for configuration but accepts CLI where necessary
- Values "install once, works everywhere" — minimal OS-specific setup

## Success Criteria

- **Hotkey-to-launch in under 200ms**: Press Ctrl+Space then a key → the bound app launches or focuses instantly
- **Cross-platform install in one command**: `curl | sh` on Linux/macOS, installer or Scoop on Windows — no manual dependency wrangling
- **Easy first update**: Update mechanism that's as simple as the install
- **Intuitive GUI**: Open the management interface → immediately understand how to add an app, set a hotkey, see status
- **App discovery**: Drop a config into a known directory and orchestratr picks it up — zero manual registration for well-structured tools
- **All tests pass on all platforms**: CI green on Linux, macOS, Windows

## Core Capabilities

1. **Global hotkey engine**: Capture Ctrl+Space + second key as a chord, dispatch to the bound app. Native implementations per OS (X11/portal on Linux, CGEventTap on macOS, RegisterHotKey on Windows).
2. **App launcher with lifecycle management**: Spawn processes, track PIDs, detect readiness (via `ready_cmd` health polling), focus existing windows instead of re-launching.
3. **App registry and configuration**: YAML-based config files defining app name, binary path, hotkey binding, readiness checks. Validated on load.
4. **Drop-in app discovery (`apps.d/`)**: Watch a well-known directory for config files. New files are auto-registered; removed files are unregistered. Projects can ship an orchestratr config snippet.
5. **Web-based management GUI**: Localhost web UI (embedded via `embed`) for adding/editing/removing apps, viewing registered hotkeys, and checking app status. Opened in the default browser.
6. **System tray integration**: Tray icon with quick access to status, pause/resume, open GUI, and quit.
7. **Daemon with autostart**: Background daemon that registers hotkeys on boot. Autostart registration per OS (systemd, launchd, Windows registry).
8. **Startup diagnostics**: `orchestratr doctor` validates configuration, binary paths, hotkey conflicts, and platform capabilities.
9. **Cross-platform install/update**: Install scripts (`install.sh`, `install.ps1`), package manager support (Homebrew, Scoop), and a simple update path.

## Out-of-Scope Boundaries

- **Not a window manager**: Orchestratr launches and focuses apps — it does not tile, resize, or manage window layout.
- **Not a package manager**: It launches tools, it does not download, install, or manage their versions. Users provide the binary.
- **Not a scripting engine or macro tool**: One hotkey maps to one app launch. No chaining, no conditional logic, no text expansion.
- **Not a clipboard manager, text expander, or automation suite**: Adjacent tools exist for those; orchestratr stays in its lane.
- **Not per-compositor plugin development**: Wayland support uses the freedesktop GlobalShortcuts portal and XWayland fallback — no compositor-specific code (see Decision 0002).

## Inviolable Principles

1. **Cross-platform parity**: Every feature ships on Linux, macOS, and Windows or doesn't ship at all. Platform-specific code is isolated behind interfaces.
2. **Single binary, zero runtime dependencies**: The distributed artifact is one binary with embedded assets. No JRE, no Python, no Node.js at runtime.
3. **Trust the user**: If the user configured a binary path and a hotkey, orchestratr launches it. No sandboxing, no allow-listing. The config file is the trust boundary.
4. **TDD is mandatory**: Tests are written before implementation. No feature merges without passing tests covering every acceptance criterion.
5. **GUI-first for configuration, CLI for power users**: The default path is the web GUI. CLI exists for scripting, automation, and headless environments.

## Security Posture

- **Threat model**: Orchestratr executes arbitrary binaries specified by the user in config files. The trust boundary is the filesystem — whoever can write config files controls what gets launched.
- **No network exposure**: The daemon API binds to `127.0.0.1` only. The web GUI uses a separate ephemeral port on localhost.
- **No authentication**: Localhost-only services don't require auth. If the attacker has local access, they already own the machine.
- **Config file validation**: All config is validated on load. Malformed YAML is rejected with clear errors, not silently ignored.
- **No auto-download of binaries**: Orchestratr never fetches executables. It only launches what the user explicitly pointed at.

## Testing Strategy

- **TDD strictly enforced**: Write failing tests before implementation. No exceptions.
- **Coverage target**: Reasonable coverage of all code paths — aim for meaningful tests, not a vanity percentage. Every acceptance criterion from every spec must have a corresponding test.
- **Test types**:
  - **Unit tests**: Every package, alongside the source (`_test.go` files). Use Go's standard `testing` package.
  - **Integration tests**: Cross-package interactions (e.g., registry watcher → launcher).
  - **E2E tests**: Per-feature, testing the full hotkey-to-launch pipeline on each platform. See `docs/E2E-SMOKE-TEST.md` for the procedure.
- **CI enforcement**: Tests must pass on Linux, macOS, and Windows before merge. `go test ./...` is the gate.

## Ambiguity Tracking

- **Unknown: Wayland portal coverage in practice** — Decision 0002 chose the freedesktop GlobalShortcuts portal, but real-world coverage on compositors like Sway and Hyprland is unclear. If portal adoption stalls, the fallback is XWayland or CLI trigger. Monitor and revisit.
- **Unknown: Package manager distribution** — Homebrew and Scoop are planned but not yet implemented. Exact tap/bucket naming and release automation TBD.
- **Unknown: Update mechanism** — "Easy update" is a goal but the specific mechanism (self-update binary, package manager update, manual re-install) is not yet decided.
- **Deferred: macOS notarization** — Required for distribution outside the App Store. Deferred until the install/update story is solidified.
- **Deferred: Plugin/extension system** — Users may want custom launch behaviors (e.g., run a script before launch). Explicitly deferred to avoid scope creep — the current model is one hotkey → one binary.
