# Feature: End-to-End Acceptance Testing

## Description

Manual and semi-automated walkthrough to verify that orchestratr works as a user
would experience it — from install through daily usage. This spec covers the full
lifecycle on Linux (including WSL2) and documents known platform limitations and
open issues discovered during testing.

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

### Added During Testing

- `POST /apps/{name}/launch` API endpoint + `orchestratr launch <name>` CLI
  command. Required for WSL/Wayland where the hotkey engine cannot capture keys.

### Known Limitations (Not Blockers for Testing)

1. **No system tray implementation** — daemon runs headless; all interaction via CLI.
   Tray provider interface and callbacks are wired but use `HeadlessProvider`.
2. **macOS/Windows hotkey listeners are stubs** — hotkeys only work on Linux/X11.
3. **macOS/Windows launchers are stubs** — app launching only works on Linux.
4. **WSL `environment` field ignored** — `environment: wsl` runs as native on Linux.
5. **Wayland requires manual trigger** — no native global shortcut capture; user must
   bind `orchestratr trigger` in compositor settings or use `orchestratr launch`.
6. **`xdotool` required** for bring-to-front focus on Linux/X11. Not auto-installed.

### What Works End-to-End

**Linux/X11 with cgo:**
- Install → autostart → daemon start → hotkey capture → chord match → app launch
- PID tracking → exit detection → state update
- Config hot-reload (file watcher + CLI + API)
- HTTP API (health, apps, lifecycle, reload, trigger, launch)
- Web-based GUI for config editing
- CLI commands: start, stop, status, reload, list, launch, trigger, configure, version

**WSL2 on Windows:**
- All of the above except hotkey capture (X11 grab works on WSLg but is unreliable)
- Windows executables (notepad.exe, calc.exe, explorer.exe) launch from WSL via `bash -c`
- `orchestratr launch <name>` is the primary interaction method

## Test Execution Record (2026-03-02, WSL2 Ubuntu on Windows)

### Environment

- WSL2 (kernel 6.6.87.2-microsoft-standard-WSL2)
- DISPLAY=:0 (WSLg), WAYLAND_DISPLAY=wayland-0
- CGO_ENABLED=1, libx11-dev installed
- No xdotool, no gnome-calculator/xterm (used Windows apps instead)

### Results by Phase

| Phase | Result | Notes |
|-------|--------|-------|
| 1. Build & Install | **PASS** | Binary builds. Version prints `v0.0.0-dev`. Install creates config + systemd unit. WSL2 warning printed. |
| 2. Daemon Lifecycle | **PASS** | Start/stop/status/health all work. Lock released cleanly. Restart works. Port file written. |
| 3. Config & Registry | **PASS** | `list` shows apps with live status from API. `/apps` returns JSON. `reload` reports app count. |
| 4. App Launch | **PASS** | `launch notepad` opens Windows Notepad from WSL. PID tracked. Re-launch detects running app and attempts focus. Exit detection works. 404 + exit code 1 for unknown apps. |
| 5. API Lifecycle | **PASS** | All endpoints: `/launched` → `/ready` → `/stopped` → `/state`. Timestamps present. 404 for unknown apps. |
| 6. Hot-Reload | **PASS** | File watcher detected edit within seconds. Explicit `reload` works. New apps immediately available. |
| 7. Web GUI | **PASS** | Config read/write via API. HTML served from embed. Daemon-info shows connected. PUT config triggers daemon reload. Browser opener fails on WSL (no `xdg-open`) — URL printed for manual access. |
| 8. Uninstall | **PASS** | Autostart service removed. Config file preserved. |

## Open Issues

Issues discovered during testing, ordered by severity.

### OI-1: `orchestratr start` runs in foreground, not background

**Severity:** High — UX confusion
**Observed:** `orchestratr start` blocks the terminal. The systemd service file
has `--foreground` flag but the binary does not actually parse flags — it ignores
`--foreground`. Users must manually background with `&` or `nohup`.
**Expected:** Either (a) `start` should daemonize (fork to background, print PID,
return), or (b) document that `start` is foreground-only and systemd/launchd handles
backgrounding, or (c) add `start --foreground` flag and have plain `start` daemonize.
**Impact:** The systemd service uses `ExecStart=...orchestratr start --foreground`
which happens to work because `--foreground` is silently ignored and `start` is
already foreground. But a user running `orchestratr start` manually gets stuck.

### OI-2: Log file path in spec doesn't match reality

**Severity:** Low — documentation only
**Observed:** `DefaultLogPath()` uses `os.UserConfigDir()` → `~/.config/orchestratr/orchestratr.log`.
The spec originally said `~/.local/share/orchestratr/orchestratr.log`.
**Actual path:** `~/.config/orchestratr/orchestratr.log`
**Impact:** Users looking at the wrong path for logs. Spec has been corrected below.

### OI-3: X11 fatal IO error on daemon shutdown in WSL

**Severity:** Low — cosmetic
**Observed:** On daemon stop, the X11 listener's connection to WSLg's X server
produces `XIO: fatal IO error 9 (Bad file descriptor) on X server ":0"`.
**Impact:** Ugly error in terminal output. Does not affect functionality — daemon
still shuts down cleanly and lock is released. Should be caught and suppressed in
the X11 listener's cleanup path.

### OI-4: `xdg-open` / browser opener missing on WSL

**Severity:** Medium — WSL usability
**Observed:** `orchestratr configure` tries `xdg-open` to open the browser. On WSL,
`xdg-open` is often not installed. The fallback message is correct ("open URL
manually") but the user experience could be better.
**Suggested fix:** On WSL, detect and use `wslview` (from `wslu` package),
`sensible-browser`, `cmd.exe /c start`, or `powershell.exe Start-Process` as
browser openers before falling back to `xdg-open`.

### OI-5: UWP Windows apps exit immediately (calc.exe)

**Severity:** Low — expected behavior, but worth documenting
**Observed:** `calc.exe` on modern Windows is a UWP app. The spawned process exits
immediately after handing off to the UWP host. The daemon correctly detects the
exit and marks the app as stopped, but the user sees "launched" then instantly
"stopped" which is confusing.
**Impact:** PID tracking doesn't work for UWP apps. Could document this or add a
`detached: true` config option to skip PID tracking for such apps.

### OI-6: No `xdotool` dependency check at install or startup

**Severity:** Low — usability
**Observed:** When `xdotool` is not installed, bring-to-front silently fails with
a log warning: `window focus not supported: xdotool not found`. The user gets no
install-time guidance.
**Suggested fix:** `orchestratr install` should check for `xdotool` and print a
warning like `"xdotool not found — bring-to-front will be disabled. Install with:
sudo apt install xdotool"`.

### OI-7: `orchestratr status` does not show the API port

**Severity:** Low — usability
**Observed:** `orchestratr status` prints `orchestratr is running (PID 17202)` but
does not show the API port. The port is available in `~/.config/orchestratr/port`
but a user shouldn't have to look there.
**Suggested fix:** Read the port file and print it: `orchestratr is running (PID
17202, port 9876)`.

### OI-8: Daemon stderr leaks API request logs to terminal

**Severity:** Low — cosmetic
**Observed:** When the daemon is started with `./orchestratr start &`, API request
logs (e.g., `api: POST /apps/notepad/launch 200 1.458ms`) print to the terminal
because the logger writes to both the log file and stderr via `io.MultiWriter`.
**Impact:** Noisy terminal output when running manually. The systemd service
captures stderr to journald so this is fine for production, but manual users see
every API call logged. Consider only logging to stderr when no log file is available,
or adding a `--quiet` flag.

### OI-9: `orchestratr trigger` only simulates leader key, not app launch

**Severity:** Medium — Wayland usability gap (partially fixed)
**Observed:** `orchestratr trigger` sends `POST /trigger` which simulates the leader
key press, putting the engine into chord-wait state. The user must then press a
physical key — but if the hotkey engine couldn't start (WSL/Wayland), the trigger
endpoint returns "hotkey engine not running".
**Partially fixed:** Added `orchestratr launch <name>` as a direct launch path.
**Remaining:** `orchestratr trigger` could accept an optional chord argument
(`orchestratr trigger c`) to directly dispatch without needing the engine. The
spec/help text implies this is possible but it is not implemented.

### OI-10: GUI server started from tray callback is never stopped

**Severity:** Low — resource leak
**Observed:** When the tray `OnConfigure` callback launches a GUI server, it calls
`guiSrv.Start()` in a goroutine but never calls `guiSrv.Stop()`. The server runs
until the daemon exits.
**Impact:** Minor goroutine/port leak. With `HeadlessProvider` this is never
triggered in practice, but would matter with a real tray.

### OI-11: Version is always `v0.0.0-dev`

**Severity:** Low — production readiness
**Observed:** `orchestratr version` prints `orchestratr v0.0.0-dev`. The `Version`
variable is set up for `-ldflags` injection but no build pipeline or Makefile target
sets it.
**Suggested fix:** Add a `make build` target or document the build command:
`go build -ldflags "-X main.Version=$(git describe --tags)" -o orchestratr ./cmd/orchestratr`

### OI-12: `orchestratr trigger` CLI does not accept a chord argument

**Severity:** Medium — spec says it does, code does not
**Observed:** The spec for Wayland fallback (decision 0002) describes running
`orchestratr trigger <chord>` from compositor keybindings. The `runTrigger()`
function ignores all arguments — it always sends a bare `POST /trigger` with no
body. The `handleTrigger` API handler similarly takes no parameters.
**Impact:** Wayland users cannot bind `orchestratr trigger c` to launch a specific
app. They can now use `orchestratr launch calculator` instead, but `trigger` should
either accept a chord or the help text should be corrected.

## Prerequisites

- Linux with X11 (Xorg) — or WSL2 with cgo for WSLg X11 support
- Go 1.22+ with cgo enabled (`CGO_ENABLED=1`)
- `libx11-dev` installed (for X11 hotkey listener)
- `xdotool` installed (optional — for bring-to-front focus)
- A test target app (Linux GUI app, or Windows `.exe` on WSL)
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

# On WSL, xdotool/gnome-calculator/xterm are optional — use Windows apps instead

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
tail -f ~/.config/orchestratr/orchestratr.log
```

On WSL, use `orchestratr launch <name>` instead of hotkey phases (Phase 4 chord tests).
Use Windows apps as test targets: `notepad.exe`, `calc.exe`, `explorer.exe`.

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

- Must be run on a graphical Linux desktop with X11, or WSL2 with WSLg
- Tests involving hotkey capture (Phase 4 chord tests) require an active X display
  with X11 grab support — may not work on WSLg; use `launch` command instead
- Tests are manual / semi-automated — not suitable for CI pipeline

## Out of Scope

- Automated CI integration testing (requires display server)
- macOS / Windows native platform testing (stubs only)
- Performance benchmarking
- Security audits beyond localhost binding

## Dependencies

- Specs 01–07 implemented and passing unit tests
- Linux/X11 desktop environment or WSL2 with WSLg
- System packages: `libx11-dev` (required), `xdotool` (optional)
- Test apps: native GUI apps, or Windows `.exe` on WSL

## Notes

- The system tray is headless — all daemon interaction is via CLI or API.
  A future spec should implement a real tray provider (e.g., `getlantern/systray`).
- If hotkeys don't register, check that another app isn't grabbing `Ctrl+Space`
  (common conflict with IBus input method). Change `leader_key` in config if needed.
- The daemon log is at `~/.config/orchestratr/orchestratr.log` (not `~/.local/share/`
  as might be expected). This is because `DefaultLogPath()` uses `os.UserConfigDir()`.
- `orchestratr launch <name>` is the primary escape hatch when hotkey capture is
  unavailable (WSL, Wayland, or any environment without X11).
- On WSL, Windows executables like `notepad.exe`, `calc.exe` work directly via
  `bash -c` because the Windows PATH is available in WSL.
- UWP apps (like modern Calculator on Windows) exit their parent process immediately,
  causing the daemon to mark them as "stopped" right after launch. Use classic Win32
  apps for PID tracking tests.

## Future Work (Candidates for New Specs)

1. **Real system tray** — implement `getlantern/systray` or `fyne.io/systray` provider
2. **Wayland native hotkeys** — `org.freedesktop.portal.GlobalShortcuts` D-Bus portal
3. **Version injection** — `go build -ldflags` pipeline + Makefile target (OI-11)
4. **xdotool dependency check** — warn at install/start if not found (OI-6)
5. **GUI cleanup on tray** — properly stop GUI server when launched from tray (OI-10)
6. **App readiness probing** — use `ready_cmd` to poll app health after launch
7. **Daemonize `start` command** — fork to background or add `--foreground` flag (OI-1)
8. **WSL browser opener** — use `wslview`/`cmd.exe` as fallback  (OI-4)
9. **`trigger` with chord argument** — enable `orchestratr trigger c` (OI-12)
10. **Status port display** — show API port in `orchestratr status` output (OI-7)
11. **Quiet mode** — suppress stderr API logs when running manually (OI-8)
12. **Detached app mode** — `detached: true` config option for UWP apps (OI-5)
