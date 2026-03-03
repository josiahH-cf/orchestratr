# End-to-End Smoke Test Procedure

A documented, repeatable procedure for verifying the full hotkey → launch pipeline. This tests the integration across orchestratr (Windows host), espansr (WSL2), and templatr (WSL2).

**Time to complete:** ~15 minutes

## Prerequisites

- Windows 11 with WSL2 and a Linux distro (e.g., Ubuntu)
- WSLg enabled (default on Windows 11) — verify with `echo $DISPLAY` inside WSL2
- espansr and templatr installed in WSL2 (`pip install -e .` in each)
- Go toolchain available in WSL2 for cross-compilation
- espansr and templatr venv binaries on the **login shell** `$PATH` (see PATH setup below)

### PATH Setup (Required)

orchestratr launches WSL apps via `wsl.exe -d <distro> -- bash -lc "<command>"`. The
`-l` flag makes bash source `~/.profile` (not `~/.bashrc` in non-interactive mode).
Aliases are not expanded. The venv binaries must be on `$PATH` via `~/.profile`:

```bash
# Add to ~/.profile (NOT just ~/.bashrc)
echo 'export PATH="$HOME/R/espansr/.venv/bin:$HOME/R/templatr/.venv/bin:$PATH"' >> ~/.profile

# Verify it works in a login shell
bash -l -c 'which espansr && which templatr'
```

## Phase A: Build

From WSL2:

```bash
cd ~/R/orchestratr
make build-windows
# Expect: orchestratr.exe in current directory
file orchestratr.exe
# Expect: PE32+ executable (console) x86-64
```

## Phase B: Install on Windows Side

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

## Phase C: Generate App Manifests

```bash
# espansr (already wired into install.sh)
cd ~/R/espansr && .venv/bin/espansr setup

# templatr
cd ~/R/templatr && .venv/bin/templatr setup
```

## Phase D: Verify Manifests

```bash
cat "$CONFIG_DIR/apps.d/espansr.yml"
# Expected: name: espansr, chord: "e", environment: wsl, command: espansr gui

cat "$CONFIG_DIR/apps.d/templatr.yml"
# Expected: name: templatr, chord: "t", environment: wsl, command: templatr
```

## Phase E: Start Daemon

**From PowerShell / Windows Terminal — NOT from WSL2:**

```powershell
# Ensure port 9876 is free (kill stale processes if needed)
netstat -ano | Select-String '9876'

# Start the daemon
& "$env:LOCALAPPDATA\orchestratr\orchestratr.exe" start

# Verify it's running
& "$env:LOCALAPPDATA\orchestratr\orchestratr.exe" doctor
```

Expected: daemon starts, prints PID, doctor shows all checks PASS.

> **Important:** The daemon must be started from native PowerShell, not via
> WSL2 interop (`/mnt/c/.../orchestratr.exe`). The WSL2 interop path cannot
> register Windows hotkeys. If you see `RegisterHotKey failed`, another
> instance may already hold the hotkey — stop it first.
>
> **Debugging:** If the daemonized process exits silently, run in foreground
> to see errors:
> ```powershell
> & "$env:LOCALAPPDATA\orchestratr\orchestratr.exe" start --foreground --verbose
> ```

## Phase F: Test via API

From PowerShell (WSL2 → Windows localhost may not work in NAT networking mode):

```powershell
# Health check
Invoke-RestMethod -Uri 'http://127.0.0.1:9876/health'
# Expected: status=ok, version=...

# List registered apps
(Invoke-RestMethod -Uri 'http://127.0.0.1:9876/apps') | ConvertTo-Json -Depth 5
# Expected: array with espansr (chord: e) and templatr (chord: t)

# Trigger espansr launch (bypasses keyboard — tests full launch pipeline)
Invoke-RestMethod -Uri 'http://127.0.0.1:9876/trigger' -Method Post -ContentType 'application/json' -Body '{"chord":"e"}'
# Expected: status=ok, app=espansr, pid=...

# Check app state
Invoke-RestMethod -Uri 'http://127.0.0.1:9876/apps/espansr/state'
# Expected: launched=True

# Trigger templatr
Invoke-RestMethod -Uri 'http://127.0.0.1:9876/trigger' -Method Post -ContentType 'application/json' -Body '{"chord":"t"}'

# Run diagnostics
& "$env:LOCALAPPDATA\orchestratr\orchestratr.exe" doctor
```

> **Note:** If WSL2 is in mirrored networking mode, `curl` from WSL2 also works.
> In NAT mode (default for older WSL2 configurations), use PowerShell instead.

## Phase G: Test Real Keyboard

1. With daemon running, press **Ctrl+Space** on the Windows desktop
2. Within 2 seconds, press **e**
3. espansr GUI should appear (via WSLg)
4. Press **Ctrl+Space** → **e** again — existing window should come to front (not a new instance)
5. Press **Ctrl+Space** → **t** — templatr should launch
6. Press **Ctrl+Space** → **?** — help overlay (if configured)

## Phase H: Cleanup

```powershell
& "$env:LOCALAPPDATA\orchestratr\orchestratr.exe" stop
```

## Pass Criteria

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

## Troubleshooting

### WSLg Not Enabled

`echo $DISPLAY` inside WSL2 should show `:0` or similar. If empty, enable WSLg in Windows Settings → Optional features, or add to `%USERPROFILE%\.wslconfig`:

```ini
[wsl2]
guiApplications=true
```

Then restart WSL2: `wsl --shutdown` from PowerShell.

### Firewall Blocking Localhost

WSL2 and Windows share `127.0.0.1` by default (networkingMode=mirrored in modern WSL2). If using NAT mode, the API may not be reachable from WSL2. Check `%USERPROFILE%\.wslconfig`:

```ini
[wsl2]
networkingMode=mirrored
```

### PATH Issues with bash -lc (Exit Code 127)

When orchestratr runs `wsl.exe -d <distro> -- bash -lc "espansr gui"`, the `-l` flag makes bash source `~/.profile` **but not** `~/.bashrc` (non-interactive). Aliases are never expanded in `bash -lc`. If apps exit with code 127 (command not found), add the venv bin dirs to `~/.profile`:

```bash
echo 'export PATH="$HOME/R/espansr/.venv/bin:$HOME/R/templatr/.venv/bin:$PATH"' >> ~/.profile
```

Verify: `bash -l -c 'which espansr && which templatr'` should print both paths.

### Port 9876 Already in Use

If the daemon fails to start with "bind: Only one usage of each socket address", another process holds port 9876. Find and kill it:

```powershell
netstat -ano | Select-String '9876'
Stop-Process -Id <PID> -Force
```

### Qt Platform Plugin

WSLg provides a Wayland compositor. PyQt6 should auto-detect it. If not, add to
your shell profile:

```bash
export QT_QPA_PLATFORM=xcb
```

### wsl.exe Distro Resolution

`wsl --list --quiet` must return at least one distro. The launcher uses the first line as the default. If you have multiple distros, set the default:

```powershell
wsl --set-default Ubuntu
```
