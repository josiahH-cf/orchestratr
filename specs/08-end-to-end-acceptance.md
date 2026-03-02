# Feature: End-to-End Acceptance Testing

## Description

Manual and semi-automated walkthrough to verify that orchestratr works as a user
would experience it — from install through daily usage. This spec covers the full
lifecycle on a Linux/X11 host (the primary supported platform today) and documents
known platform limitations.

## Project State Summary

### Completed Specs (1–7)

| Spec | Feature | Status |
|------|---------|--------|
| 01 | Core Daemon & System Tray | Complete |
| 02 | Cross-Platform Hotkey Engine | Complete |
| 03 | App Registry & Configuration | Complete |
| 03a | Registry–Daemon Integration | Complete (wired in `runStart()`) |
| 03b | `orchestratr reload` CLI | Complete (handled in main.go) |
| 04 | HTTP API & IPC Protocol | Complete |
| 04a | Remove Dead HealthServer | Complete (no dead code remains) |
| 04b | Hot-Reload Completeness | Complete (`SwapChords` + `Sync` wired) |
| 04c | App Lifecycle Endpoints | Complete (`/state`, `/stopped` exist) |
| 04d | Version & Validation Housekeeping | Partially addressed |
| 05 | Cross-Environment Launching | Complete (WSL deferred — Windows-only) |
| 06 | Management GUI | Complete |
| 07 | Installer & Permissions | Complete |

### Known Limitations (Not Blockers for Linux/X11 Testing)

1. **No system tray implementation** — daemon runs headless; all interaction via CLI.
   Tray provider interface and callbacks are wired but use `HeadlessProvider`.
2. **macOS/Windows hotkey listeners are stubs** — hotkeys only work on Linux/X11.
3. **macOS/Windows launchers are stubs** — app launching only works on Linux.
4. **WSL `environment` field ignored** — `environment: wsl` runs as native on Linux.
5. **Wayland requires manual trigger** — no native global shortcut capture; user must
   bind `orchestratr trigger <chord>` in compositor settings.
6. **`xdotool` required** for bring-to-front focus on Linux/X11. Not auto-installed.

### What Works End-to-End (Linux/X11 with cgo)

- Install → autostart → daemon start → hotkey capture → chord match → app launch
- PID tracking → exit detection → state update
- Config hot-reload (file watcher + CLI + API)
- HTTP API (health, apps, lifecycle, reload, trigger)
- Web-based GUI for config editing
- CLI commands: start, stop, status, reload, list, trigger, configure, version

## Prerequisites

- Linux with X11 (Xorg) — not Wayland-only
- Go 1.22+ with cgo enabled (`CGO_ENABLED=1`)
- `libx11-dev` installed (for X11 hotkey listener)
- `xdotool` installed (for bring-to-front focus)
- A terminal emulator or simple app to use as a test target (e.g., `xterm`, `gnome-calculator`)
- No other instance of orchestratr running

## Acceptance Criteria

### Phase 1: Build & Install

- [ ] `go build -o orchestratr ./cmd/orchestratr` succeeds with no errors
- [ ] `./orchestratr version` prints a version string
- [ ] `./orchestratr install` completes: creates default config, configures autostart
- [ ] Default config exists at `~/.config/orchestratr/config.yml`
- [ ] Autostart unit exists at `~/.config/systemd/user/orchestratr.service`

### Phase 2: Daemon Lifecycle

- [ ] `./orchestratr start` launches daemon in background, prints PID
- [ ] `./orchestratr status` shows "running" with PID and port
- [ ] `curl http://127.0.0.1:9876/health` returns `{"status":"ok","version":"..."}`
- [ ] Port discovery file exists at `~/.config/orchestratr/port`
- [ ] Log file created at `~/.local/share/orchestratr/orchestratr.log`
- [ ] `./orchestratr stop` shuts down the daemon gracefully
- [ ] `./orchestratr status` shows "not running" after stop
- [ ] Second `./orchestratr start` does not conflict (lock released on stop)

### Phase 3: Configuration & Registry

- [ ] Edit `~/.config/orchestratr/config.yml` to add a test app:
      ```yaml
      leader_key: "ctrl+space"
      chord_timeout_ms: 2000
      api_port: 9876
      log_level: info
      apps:
        - name: calculator
          description: "GNOME Calculator"
          chord: "c"
          command: "gnome-calculator"
          environment: native
        - name: terminal
          description: "XTerm"
          chord: "t"
          command: "xterm"
          environment: native
      ```
- [ ] `./orchestratr reload` triggers config reload, confirms app count
- [ ] `./orchestratr list` shows both apps with status
- [ ] `curl http://127.0.0.1:9876/apps` returns JSON with both apps

### Phase 4: Hotkey & App Launching

- [ ] Press `Ctrl+Space` — hotkey engine enters chord-wait state (check log)
- [ ] Press `c` within timeout — gnome-calculator launches
- [ ] `./orchestratr list` shows calculator as "running"
- [ ] `curl http://127.0.0.1:9876/apps/calculator/state` shows launched state
- [ ] Press `Ctrl+Space` then `c` again — existing calculator window is focused (not a duplicate)
- [ ] Close calculator manually — daemon detects exit (check log)
- [ ] `./orchestratr list` shows calculator as "stopped"
- [ ] Press `Ctrl+Space` then `t` — xterm launches
- [ ] Press `Ctrl+Space` then wait for timeout — chord cancelled silently (check log)
- [ ] Press `Ctrl+Space` then an unmapped key — no action, returns to idle

### Phase 5: API Lifecycle Notifications

- [ ] `curl -X POST http://127.0.0.1:9876/apps/calculator/launched` returns 200
- [ ] `curl http://127.0.0.1:9876/apps/calculator/state` shows launched=true
- [ ] `curl -X POST http://127.0.0.1:9876/apps/calculator/ready` returns 200
- [ ] `curl http://127.0.0.1:9876/apps/calculator/state` shows ready=true
- [ ] `curl -X POST http://127.0.0.1:9876/apps/calculator/stopped` resets state
- [ ] `curl -X POST http://127.0.0.1:9876/apps/nonexistent/launched` returns 404

### Phase 6: Hot-Reload via File Edit

- [ ] With daemon running, edit config.yml to add a third app (e.g., `chord: "f"`)
- [ ] Within a few seconds, daemon reloads automatically (check log for "config reloaded")
- [ ] `./orchestratr list` now shows three apps
- [ ] New chord `f` is active — press `Ctrl+Space` then `f` to launch

### Phase 7: Web GUI

- [ ] `./orchestratr configure` opens browser to the management GUI
- [ ] GUI shows the app table with all configured apps
- [ ] Add a new app via the GUI form — config.yml updates and daemon reloads
- [ ] Remove an app via GUI — config.yml updates
- [ ] Edit leader key setting — config.yml updates
- [ ] Close browser tab — daemon continues running unaffected

### Phase 8: Manual Trigger (Wayland Fallback)

- [ ] `./orchestratr trigger c` launches calculator (or focuses if running)
- [ ] `./orchestratr trigger nonexistent` returns error

### Phase 9: Uninstall

- [ ] `./orchestratr stop` (if running)
- [ ] `./orchestratr uninstall` removes autostart entry
- [ ] Autostart unit no longer exists at `~/.config/systemd/user/orchestratr.service`
- [ ] Config file is preserved (not deleted by uninstall)

## Test Procedure

### Setup

```bash
# Install dependencies (Ubuntu/Debian)
sudo apt-get install -y libx11-dev xdotool gnome-calculator xterm

# Build
cd /path/to/orchestratr
CGO_ENABLED=1 go build -o orchestratr ./cmd/orchestratr

# Verify unit tests pass
go test ./...
```

### Run Through

Execute each phase in order. After each step, verify the acceptance criterion.
Open a second terminal to monitor the log in real time:

```bash
tail -f ~/.local/share/orchestratr/orchestratr.log
```

Record results as PASS/FAIL per criterion. Any FAIL blocks the phase; investigate
before continuing to the next phase.

### Cleanup

```bash
./orchestratr stop 2>/dev/null
./orchestratr uninstall
rm -f ./orchestratr
```

## Affected Areas

All packages — this is a full-system integration test:
- `cmd/orchestratr/` — CLI entry point
- `internal/daemon/` — lifecycle, lock, logging
- `internal/hotkey/` — leader key + chord capture
- `internal/registry/` — config parsing, validation, hot-reload
- `internal/api/` — HTTP server, state tracking
- `internal/launcher/` — process launch, PID tracking, focus
- `internal/gui/` — web management interface
- `internal/autostart/` — systemd user unit
- `internal/platform/` — WSL detection, accessibility
- `internal/tray/` — headless provider (no real tray)

## Constraints

- Must be run on a graphical Linux desktop with X11 (not headless CI)
- Tests involving hotkey capture require an active X display
- Tests are manual / semi-automated — not suitable for CI pipeline

## Out of Scope

- Automated CI integration testing (requires display server)
- macOS / Windows platform testing (stubs only)
- WSL2 cross-environment launch testing (requires Windows host)
- Performance benchmarking
- Security audits beyond localhost binding

## Dependencies

- Specs 01–07 implemented and passing unit tests
- Linux/X11 desktop environment
- System packages: `libx11-dev`, `xdotool`
- Test apps: `gnome-calculator`, `xterm` (or substitutes)

## Notes

- The system tray is headless — all daemon interaction is via CLI or API.
  A future spec should implement a real tray provider (e.g., `getlantern/systray`).
- If hotkeys don't register, check that another app isn't grabbing `Ctrl+Space`
  (common conflict with IBus input method). Change `leader_key` in config if needed.
- The daemon log (`~/.local/share/orchestratr/orchestratr.log`) is the primary
  debugging tool. Watch it during all phases.
- `orchestratr trigger` is the escape hatch when hotkey capture isn't available.

## Future Work (Candidates for New Specs)

1. **Real system tray** — implement `getlantern/systray` or `fyne.io/systray` provider
2. **Wayland native hotkeys** — `org.freedesktop.portal.GlobalShortcuts` D-Bus portal
3. **Version injection** — `go build -ldflags "-X main.Version=..."` pipeline
4. **xdotool dependency check** — warn at install/start if not found
5. **GUI cleanup on tray** — properly stop GUI server when launched from tray callback
6. **App readiness probing** — use `ready_cmd` to poll app health after launch
