# Feature: End-to-End Hotkey Launch Pipeline

**Status:** Implemented
**Project:** orchestratr (primary), templatr, espansr
**Size:** L — cross-project, 7 tasks across 3 repos

## Description

Verify and fix the last-mile gaps preventing the core user story: **press Ctrl+Space → 'e' → espansr launches; Ctrl+Space → 't' → templatr launches**. The architecture (drop-in discovery, connectors, hotkey engine, launcher, API) is fully implemented across all three projects. What's missing is the build/install glue and a verified end-to-end test path.

The user runs WSL2 from Windows. orchestratr must run as a native Windows binary on the Windows host. espansr and templatr run inside WSL2. orchestratr launches them via `wsl.exe -d <distro> -- bash -lc "<command>"`. App manifests are written to the Windows-side `%APPDATA%\orchestratr\apps.d\` via `/mnt/c/` from WSL2.

### Why This Is One Feature

These tasks are independently small but form a single dependency chain: a Windows binary must exist (task 1) → app manifests must be generated at install time (tasks 2–3) → the pipeline must be testable (tasks 4–6) → real keyboard verification (task 7). Splitting into separate issues would lose the sequencing context.

## Current State

### What's Built and Working

**orchestratr (Go):**
- Leader key (Ctrl+Space) + chord capture: Linux/X11 via XGrabKey (482 lines), Windows via RegisterHotKey + WH_KEYBOARD_LL (511 lines) — both real implementations, not stubs
- Drop-in `apps.d/` directory scanning with fsnotify watcher, debounced hot-reload, merge conflict resolution
- Registry: thread-safe chord lookup, atomic config swap on reload
- Launcher: Linux `bash -c`, Windows `cmd.exe /c` (native) or `wsl.exe -d <distro> -- bash -lc` (WSL) — full environment routing
- `ready_cmd` health polling at 500ms intervals with configurable timeout
- Bring-to-front: Linux xdotool, Windows EnumWindows + SetForegroundWindow + AttachThreadInput
- REST API on localhost:9876 — `/health`, `/apps`, `/trigger` (with chord body), `/launch`, `/reload`
- `/trigger {"chord":"e"}` bypasses keyboard — looks up chord in registry, launches matching app directly. This is the **testable seam** for E2E verification.
- Daemon lifecycle: PID lock, daemonize, start/stop/status/reload, WSL2 guard (refuses to start inside WSL2 without `--force`)
- `orchestratr doctor`: config validation, apps.d scan, command resolution, WSL check, ready_cmd syntax — with `--json` output
- `orchestratr start --verbose`: real-time event streaming to stderr (leader_key, chord_received, app_launched, app_focused, config_reloaded, etc.)
- Autostart: systemd (Linux), Launch Agent (macOS), Windows registry (real `HKCU\...\Run` writes)
- Web GUI: embedded SPA for config editing
- CI: already tests on `{ubuntu, macos, windows}` × `{go 1.22, 1.23}` — confirms Windows builds succeed
- **433 tests, all passing**

**espansr (Python):**
- Connector module `espansr/integrations/orchestratr.py`: all 4 functions (`generate_manifest()`, `manifest_needs_update()`, `get_status_json()`, `resolve_orchestratr_apps_dir()`)
- `espansr status --json`: returns version, status, config_dir, template_count, last_sync
- `espansr setup` calls `generate_manifest()` conditionally — handles dry-run, cleans legacy files
- `install.sh` runs `espansr setup` → manifest IS generated at install time
- Flat YAML manifest: `name: espansr`, `chord: "e"`, `command: espansr gui`, `environment: wsl` (on WSL2)
- WSL2 path resolution: writes to `/mnt/c/Users/<win_user>/AppData/Roaming/orchestratr/apps.d/`
- **20 tests across 8 classes, all passing**

**templatr (Python):**
- Connector module `templatr/integrations/orchestratr.py`: all 4 functions implemented (254 lines)
- `templatr status --json`: returns version, status, config_dir, template_count, llm_server_status
- GUI Integrations dialog: File → Integrations, register button, staleness detection, startup check with status bar hint
- Flat YAML manifest: `name: templatr`, `chord: "t"`, `command: templatr`, `environment: wsl` (on WSL2)
- WSL2 path resolution: `cmd.exe /C echo %USERNAME%` with fallback scan of `/mnt/c/Users/`
- **621 lines of tests, 7 test classes**
- **Gap:** No `setup` CLI command. `generate_manifest()` is only callable from the GUI Integrations dialog or directly via Python import. `install.sh` never calls it.

### What's Missing (The 7 Tasks)

| # | Task | Project | Gap |
|---|------|---------|-----|
| 1 | Add `build-windows` Makefile target | orchestratr | Makefile hardcodes `CGO_ENABLED=1` (for X11). Windows needs `CGO_ENABLED=0 GOOS=windows`. |
| 2 | Add `templatr setup` CLI command | templatr | No CLI path to `generate_manifest()`. Only the GUI dialog. |
| 3 | Wire `templatr setup` into install.sh | templatr | install.sh never registers with orchestratr. |
| 4 | Write tests for `templatr setup` | templatr | No test coverage for the new CLI command. |
| 5 | Add cross-compile CI step | orchestratr | CI already builds on Windows, but doesn't cross-compile from Linux. |
| 6 | Create E2E smoke test procedure | orchestratr | No integration test exists anywhere. Need a documented, repeatable procedure. |
| 7 | Verify real hotkey on Windows | all | The Windows hotkey/launcher code is written but has never been tested on a real Windows host. |

## Acceptance Criteria

- [ ] `make build-windows` in orchestratr produces a working `orchestratr.exe` (exit 0, binary exists, no CGO required)
- [ ] `templatr setup` CLI command exists, calls `generate_manifest()`, prints registration status, supports `--dry-run`
- [ ] `templatr setup` is called by `install.sh` during installation (between `setup_config` and `smoke_test`)
- [ ] `templatr setup` has automated tests: manifest created, skip when absent, dry-run, idempotent
- [ ] CI cross-compile step builds `orchestratr.exe` from `ubuntu-latest` runner without failure
- [ ] E2E smoke test procedure is documented and exercises: cross-compile → install → manifest generation → daemon start → API trigger → app launch verification
- [ ] Real keyboard test: Ctrl+Space → 'e' on Windows desktop launches espansr GUI inside WSL2 via `wsl.exe`

## Task Details

### Task 1: Add `build-windows` Makefile Target

**File:** `orchestratr/Makefile`

**Change:** Add a new target after the existing `build` target:

```makefile
build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BINARY).exe ./cmd/orchestratr
```

**Why CGO_ENABLED=0:** The Windows platform layer uses pure syscalls via `golang.org/x/sys/windows` — no CGO needed. The only CGO dependency is the Linux X11 listener (`listener_linux_x11.go` links `-lX11`), which has a `//go:build linux && !cgo` fallback (`x11_available_nocgo.go`) that returns nil. On `GOOS=windows`, the X11 files are excluded entirely by build tags.

**Also update `clean`:** Add `rm -f $(BINARY).exe` alongside the existing `rm -f $(BINARY)`.

**Verification:** `make build-windows && file orchestratr.exe` — should show `PE32+ executable (console) x86-64`.

---

### Task 2: Add `templatr setup` CLI Command

**File:** `templatr/templatr/__main__.py`

**Current state:** The file has 110 lines. Subcommands are `status` (L62–73) and `gui` (L76). The dispatch is at L80–86.

**Change:** Add a `setup` subparser and `_cmd_setup()` handler matching espansr's pattern (see `espansr/espansr/__main__.py` L53–175).

Add subparser after the `gui` parser (around L76):

```python
# templatr setup
setup_parser = subparsers.add_parser("setup", help="Post-install setup and orchestratr registration")
setup_parser.add_argument(
    "--dry-run",
    action="store_true",
    help="Preview actions without writing files",
)
```

Add handler function:

```python
def _cmd_setup(args) -> int:
    """Handle the ``templatr setup`` subcommand.

    Registers templatr with orchestratr by writing a manifest to the
    apps.d/ drop-in directory. Skips silently when orchestratr is not
    installed. Supports --dry-run for previewing without writing.
    """
    from templatr.integrations.orchestratr import (
        generate_manifest,
        manifest_needs_update,
        resolve_orchestratr_apps_dir,
    )

    dry_run = getattr(args, "dry_run", False)
    apps_dir = resolve_orchestratr_apps_dir()

    if dry_run:
        if apps_dir is not None:
            manifest_path = apps_dir / "templatr.yml"
            print(f"[dry-run] Would write orchestratr manifest to {manifest_path}")
        else:
            print("[dry-run] orchestratr not found — would skip app registration")
        return 0

    if apps_dir is not None:
        if manifest_needs_update(apps_dir):
            result = generate_manifest()
            if result:
                print(f"Registered with orchestratr: {result}")
            else:
                print("orchestratr registration failed (could not resolve apps dir)")
        else:
            print("orchestratr manifest: up to date")
    else:
        print("orchestratr not found — skipping app registration")

    return 0
```

Wire into dispatch (add alongside the existing `args.command == "status"` check):

```python
if args.command == "setup":
    return _cmd_setup(args)
```

**Note on `generate_manifest()` API difference:** templatr's `generate_manifest()` takes no arguments (resolves `apps_dir` internally), unlike espansr's which takes `apps_dir: Path`. The handler should call `manifest_needs_update(apps_dir)` first (passing the resolved dir), then call `generate_manifest()` without args. Check the actual function signature at `templatr/integrations/orchestratr.py` L170 before implementing.

---

### Task 3: Wire `templatr setup` into install.sh

**File:** `templatr/install.sh`

**Current install flow** (L398–410):
```bash
main() {
    ...
    install_system_deps
    setup_python_env
    build_llama_cpp
    setup_config
    setup_alias       # ← insert new step before this
    smoke_test
    print_summary
}
```

**Change:** Add a `register_orchestratr()` function and call it between `setup_config` and `setup_alias`:

```bash
# Register with orchestratr (optional — skips if orchestratr not installed)
register_orchestratr() {
    log_info "Registering with orchestratr..."
    if "$VENV_DIR/bin/templatr" setup 2>/dev/null; then
        log_success "orchestratr registration complete"
    else
        log_warn "orchestratr not installed — skipping registration"
    fi
}
```

Update `main()`:
```bash
    setup_config
    register_orchestratr
    setup_alias
    smoke_test
```

**Behavior:** If orchestratr is not installed, `templatr setup` prints "orchestratr not found" and exits 0, so the installer continues. Non-blocking, passive.

---

### Task 4: Write Tests for `templatr setup`

**File:** `templatr/tests/test_orchestratr_connector.py`

**Existing patterns to follow:** `TestCLIStatusCommand` class (L425–476) tests CLI subcommands by importing `main` from `templatr.__main__` and mocking `sys.argv`.

**New test class: `TestSetupCommand`**

Tests to write:

1. **`test_setup_generates_manifest`** — mock `resolve_orchestratr_apps_dir()` to return a temp dir, call `templatr setup`, assert manifest file exists with correct content
2. **`test_setup_skips_when_absent`** — mock `resolve_orchestratr_apps_dir()` to return `None`, call `templatr setup`, assert exit 0 and "not found" in output
3. **`test_setup_dry_run`** — mock to return a temp dir, call `templatr setup --dry-run`, assert no manifest file written, "[dry-run]" in output
4. **`test_setup_idempotent`** — call setup twice, second call should print "up to date"
5. **`test_setup_subcommand_accepted`** — verify argparse accepts `setup` without error

Follow the fixture patterns established in `TestCLIStatusCommand` — use `monkeypatch` for `sys.argv`, `capsys` for output capture.

---

### Task 5: Add Cross-Compile CI Step

**File:** `orchestratr/.github/workflows/ci.yml`

**Current CI:** Matrix `{ubuntu-latest, macos-latest, windows-latest}` × `{go 1.22, 1.23}`. Steps: lint → test → build.

**Change:** Add a step to the `ubuntu-latest` runner (after the existing build step):

```yaml
    - name: Cross-compile for Windows
      if: matrix.os == 'ubuntu-latest'
      run: CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o orchestratr.exe ./cmd/orchestratr
```

This catches build tag isolation regressions — if someone adds a CGO dependency to a Windows-tagged file, CI will fail on the next PR.

---

### Task 6: Create E2E Smoke Test Procedure

**Deliverable:** A documented, repeatable procedure — NOT an automated script (the pipeline crosses OS boundaries that can't be automated in a single shell session).

**Location:** `orchestratr/docs/E2E-SMOKE-TEST.md`

**The procedure:**

#### Prerequisites
- WSL2 with a Linux distro running
- espansr and templatr installed in WSL2 (`pip install -e .` in each)
- Go toolchain in WSL2 for cross-compilation
- WSLg enabled (Windows 11 default) for GUI app display

#### Phase A: Build
```bash
cd ~/R/orchestratr
make build-windows
# Expect: orchestratr.exe in current directory
```

#### Phase B: Install on Windows side
```bash
WIN_USER="$(cmd.exe /C 'echo %USERNAME%' 2>/dev/null | tr -d '\r')"
INSTALL_DIR="/mnt/c/Users/$WIN_USER/AppData/Local/orchestratr"
CONFIG_DIR="/mnt/c/Users/$WIN_USER/AppData/Roaming/orchestratr"

mkdir -p "$INSTALL_DIR" "$CONFIG_DIR/apps.d"
cp orchestratr.exe "$INSTALL_DIR/"

# Seed default config if missing
[ -f "$CONFIG_DIR/config.yml" ] || cat > "$CONFIG_DIR/config.yml" << 'EOF'
leader_key: ctrl+space
chord_timeout_ms: 2000
api_port: 9876
apps: []
EOF
```

#### Phase C: Generate app manifests
```bash
# espansr (run espansr setup — will generate manifest at correct WSL2→Windows path)
cd ~/R/espansr && .venv/bin/espansr setup

# templatr (after task 2 is implemented)
cd ~/R/templatr && .venv/bin/templatr setup
```

#### Phase D: Verify manifests
```bash
cat "$CONFIG_DIR/apps.d/espansr.yml"
# Expected: name: espansr, chord: "e", environment: wsl, command: espansr gui

cat "$CONFIG_DIR/apps.d/templatr.yml"
# Expected: name: templatr, chord: "t", environment: wsl, command: templatr
```

#### Phase E: Start daemon (from PowerShell / Windows Terminal — NOT from WSL2)
```powershell
& "$env:LOCALAPPDATA\orchestratr\orchestratr.exe" start --verbose
```
Expected: daemon starts, prints PID, begins listening for hotkeys.

#### Phase F: Test via API (from WSL2 — shares localhost with Windows)
```bash
# Health check
curl -s http://127.0.0.1:9876/health
# Expected: {"status":"ok","version":"..."}

# List registered apps
curl -s http://127.0.0.1:9876/apps | python3 -m json.tool
# Expected: array with espansr (chord: e) and templatr (chord: t)

# Trigger espansr launch via API (bypasses keyboard — tests full launch pipeline)
curl -s -X POST http://127.0.0.1:9876/trigger \
  -H 'Content-Type: application/json' \
  -d '{"chord":"e"}'
# Expected: {"status":"ok","app":"espansr","pid":...}
# Expected in --verbose output: app_launched event

# Check app state
curl -s http://127.0.0.1:9876/apps/espansr/state
# Expected: state: "launched" or "ready"

# Trigger templatr
curl -s -X POST http://127.0.0.1:9876/trigger \
  -H 'Content-Type: application/json' \
  -d '{"chord":"t"}'

# Run diagnostics
powershell.exe -Command "& '$env:LOCALAPPDATA\orchestratr\orchestratr.exe' doctor"
```

#### Phase G: Test real keyboard
1. With daemon running, press **Ctrl+Space** on Windows desktop
2. Within 2 seconds, press **e**
3. espansr GUI should appear (via WSLg)
4. Press **Ctrl+Space** → **e** again — existing window should come to front (not a new instance)
5. Press **Ctrl+Space** → **t** — templatr should launch
6. Press **Ctrl+Space** → **?** — help overlay (if configured)

#### Phase H: Cleanup
```powershell
& "$env:LOCALAPPDATA\orchestratr\orchestratr.exe" stop
```

#### Pass Criteria Table
| Check | Method | Pass |
|-------|--------|------|
| Binary builds | `make build-windows` exit 0 | `file orchestratr.exe` shows PE32+ |
| Config loads | `orchestratr.exe doctor` | PASS on config checks |
| Apps discovered | `curl /apps` | Both apps listed |
| API trigger works | `POST /trigger {"chord":"e"}` | 200 OK, PID returned |
| WSL exec works | espansr GUI appears | Window visible via WSLg |
| Re-activation focuses | Trigger same chord twice | Existing window focused, no new process |
| Doctor passes | `orchestratr.exe doctor` | No FAIL checks |
| Real hotkey works | Ctrl+Space → e on keyboard | Same as API trigger result |

---

### Task 7: Real Keyboard Verification

This is NOT a code task — it's the manual Phase G from the smoke test above, performed after all previous tasks are complete. It verifies that Windows `RegisterHotKey` and `WH_KEYBOARD_LL` actually capture keystrokes in the real Windows desktop environment.

**Potential issues to watch for:**
- **WSLg not enabled:** `echo $DISPLAY` inside WSL2 should show `:0` or similar. If empty, WSLg needs to be enabled in Windows Settings → Optional features.
- **Firewall blocking localhost:** WSL2 and Windows share `127.0.0.1` by default (networkingMode=mirrored in modern WSL2). If using NAT mode, the API may not be reachable from WSL2.
- **`wsl.exe` distro resolution:** `wsl --list --quiet` must return at least one distro. The launcher uses the first line as the default.
- **PATH inside WSL2:** When orchestratr runs `wsl.exe -d Ubuntu -- bash -lc "espansr gui"`, the `bash -l` (login shell) sources `.bashrc`/`.profile`. The `espansr` alias or venv must be on PATH in a login shell context.
- **Qt platform plugin:** WSLg provides a Wayland compositor. PyQt6 should auto-detect it. If not, `QT_QPA_PLATFORM=xcb` may be needed in the shell profile.

## Affected Areas

| Project | Files | Change Type |
|---------|-------|-------------|
| orchestratr | `Makefile` | Modify — add `build-windows` + `clean` |
| orchestratr | `.github/workflows/ci.yml` | Modify — add cross-compile step |
| orchestratr | `docs/E2E-SMOKE-TEST.md` | Create — smoke test procedure |
| templatr | `templatr/__main__.py` | Modify — add `setup` subcommand |
| templatr | `install.sh` | Modify — add `register_orchestratr()` step |
| templatr | `tests/test_orchestratr_connector.py` | Modify — add `TestSetupCommand` class |

**Not modified:** espansr (already complete), orchestratr Go source (already complete).

## Constraints

- No new dependencies in any project
- templatr `setup` command must exit 0 even when orchestratr is absent (passive behavior)
- `install.sh` change must not break installation when orchestratr is not installed
- Cross-compile must work without installing a Windows cross-compiler toolchain (CGO_ENABLED=0)
- E2E smoke test procedure must be runnable by a human in under 15 minutes

## Out of Scope

- System tray implementation (separate spec: `orchestratr/specs/system-tray.md`)
- Wayland hotkey support (separate decision: `orchestratr/decisions/0002-wayland-hotkey-strategy.md`)
- macOS support (not the target platform)
- Automated E2E test CI (would require a Windows VM with WSL2 — complex, deferred)
- espansr code changes (connector is complete and tested)

## Dependencies

- **Implemented:** `orchestratr/specs/drop-in-app-discovery.md` — apps.d scanning, fsnotify watcher
- **Implemented:** `orchestratr/specs/ready-cmd-health-polling.md` — ready_cmd polling
- **Implemented:** `orchestratr/specs/windows-platform-layer.md` — Win32 hotkey, launcher, focus, autostart
- **Implemented:** `orchestratr/specs/startup-diagnostic.md` — doctor command, --verbose events
- **Implemented:** `espansr/specs/archive/manifest-schema-alignment.md` — flat schema connector
- **Implemented (code, not install.sh):** `templatr/specs/orchestratr-connector.md` — connector module, status CLI, GUI dialog
- **Protocol reference:** `/specs/orchestratr-app-connector-protocol.md`
- **Connector docs:** `orchestratr/docs/CONNECTOR.md`

## Notes

### Key Architectural Insight: `/trigger` as Test Seam

The `/trigger` API endpoint accepts `{"chord":"e"}` in the request body and performs the **full launch pipeline** — registry lookup → environment routing → `wsl.exe` invocation → PID tracking → state update. This bypasses the keyboard entirely, making the entire launch pipeline testable from a `curl` command in WSL2. This is the critical path for Phase F of the smoke test.

### templatr `generate_manifest()` API differs from espansr

- **espansr:** `generate_manifest(apps_dir: Path) -> Path` — caller resolves and passes the apps dir
- **templatr:** `generate_manifest() -> Optional[Path]` — resolves apps dir internally

The `_cmd_setup()` handler must account for this: call `resolve_orchestratr_apps_dir()` for the needs-update check, but call `generate_manifest()` with no arguments for the actual write.

### espansr `install.sh` Already Works

espansr's installer runs `espansr setup`, which conditionally calls `generate_manifest()`. No changes needed. The manifest is generated at install time if orchestratr's config directory exists on the Windows side.

### Login Shell PATH Requirement

When orchestratr runs `wsl.exe -d <distro> -- bash -lc "espansr gui"`, the `-l` flag makes bash read `~/.profile` and `~/.bashrc`. Both espansr and templatr's `install.sh` scripts add aliases to the shell RC file, but aliases aren't expanded in non-interactive `bash -lc` contexts. The venv binary path (e.g., `~/R/espansr/.venv/bin/espansr`) must be on `$PATH`, or the manifest `command` field must use the absolute path.

**Current espansr manifest:** `command: espansr gui` — relies on the command being on PATH.
**Current templatr manifest:** `command: templatr` — same.

If this fails during the smoke test, the fix is to update `generate_manifest()` in each connector to use the absolute venv path. This is a known risk but should be verified before adding complexity.
