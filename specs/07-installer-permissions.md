# Feature: Installer & Permissions

**Status:** Not started  
**Project:** orchestratr

## Description

orchestratr is a background daemon that must start at login and register system-wide hotkeys. This requires platform-specific installation steps: autostart configuration, accessibility permissions (macOS), and optional PATH integration. The installer experience is core functionality — if orchestratr can't start reliably or can't capture hotkeys, it has no value. This spec covers every installation and permission concern across all supported platforms.

## Acceptance Criteria

- [ ] `install.sh` installs orchestratr on Linux and macOS (venv or single binary depending on tech stack)
- [ ] `install.ps1` installs orchestratr on Windows
- [ ] Installer configures autostart at login (systemd user service on Linux, Launch Agent on macOS, Startup folder/registry on Windows)
- [ ] Installer detects and prompts for macOS Accessibility permission (required for `CGEventTap`)
- [ ] Installer verifies the chosen leader key is registerable and warns if it fails (e.g., Wayland without compositor support)
- [ ] `orchestratr start` / `orchestratr stop` / `orchestratr status` CLI commands work after install
- [ ] Uninstall path: `orchestratr uninstall` removes autostart config, config directory (with confirmation), and binary/venv
- [ ] Installer is idempotent — running it again updates without duplicating autostart entries
- [ ] WSL2 scenario: installer detects it's running inside WSL and warns that the daemon should run on the Windows side for hotkey support

## Affected Areas

| Area | Files |
|------|-------|
| **Create** | `install.sh` — Linux/macOS installer |
| **Create** | `install.ps1` — Windows installer |
| **Create** | `orchestratr/autostart/linux.py` (or equivalent) — systemd user service generation |
| **Create** | `orchestratr/autostart/macos.py` — Launch Agent plist generation |
| **Create** | `orchestratr/autostart/windows.py` — Startup registry/shortcut |
| **Create** | `orchestratr/permissions.py` — macOS accessibility check, Wayland capability check |

## Constraints

- **No root/admin required** for standard installation (user-level autostart, user-level hotkeys)
- macOS Accessibility permission requires user interaction (System Preferences) — installer must guide the user, not silently fail
- Linux autostart via systemd user service (`~/.config/systemd/user/orchestratr.service`) — no system-level service
- Windows autostart via `HKCU\Software\Microsoft\Windows\CurrentVersion\Run` registry key or Start Menu Startup folder
- Installer must not modify system-wide config files or require reboot

## Out of Scope

- Package manager distribution (apt, brew, winget, chocolatey) — follow-up concern
- Auto-update mechanism
- Multi-user installation (system-wide daemon serving multiple users)
- Container or VM deployment

## Dependencies

- `01-core-daemon.md` — the daemon is what gets installed and autostarted
- `02-hotkey-engine.md` — installer must verify hotkey registration capability

## Notes

### Platform-specific autostart

**Linux (systemd):**
```ini
# ~/.config/systemd/user/orchestratr.service
[Unit]
Description=orchestratr — app launcher daemon
After=graphical-session.target

[Service]
ExecStart=/path/to/orchestratr start --foreground
Restart=on-failure
RestartSec=5

[Install]
WantedBy=default.target
```

**macOS (launchd):**
```xml
<!-- ~/Library/LaunchAgents/com.orchestratr.daemon.plist -->
<plist version="1.0">
<dict>
  <key>Label</key><string>com.orchestratr.daemon</string>
  <key>ProgramArguments</key>
  <array>
    <string>/path/to/orchestratr</string>
    <string>start</string>
    <string>--foreground</string>
  </array>
  <key>RunAtLoad</key><true/>
  <key>KeepAlive</key><true/>
</dict>
</plist>
```

**Windows (registry):**
```
HKCU\Software\Microsoft\Windows\CurrentVersion\Run
  orchestratr = "C:\path\to\orchestratr.exe start --foreground"
```

### macOS accessibility permission flow

1. Installer runs `orchestratr start`
2. Daemon attempts to register hotkey via `CGEventTap`
3. macOS blocks the request → daemon detects the failure
4. Daemon shows a dialog: "orchestratr needs Accessibility permission to capture hotkeys"
5. Dialog includes a button that opens System Preferences → Privacy → Accessibility
6. User grants permission, daemon retries hotkey registration
7. If still denied: daemon runs in degraded mode (tray only, no hotkeys) with a persistent tray warning

### WSL2 detection

If the installer detects WSL2:
```
⚠ orchestratr is running inside WSL2.
  System-wide hotkeys require the daemon to run on the Windows side.
  
  Recommended: install orchestratr in Windows PowerShell instead:
    .\install.ps1
  
  Continuing will install, but hotkeys will not work from WSL2.
  Apps launched by orchestratr from Windows can still target WSL commands.
```

### Conflict detection

At install and first-start, orchestratr should:
1. Attempt to register the configured leader key
2. If registration fails: log the error and suggest an alternative
3. If the key is registered but known to conflict (e.g., Ctrl+Space on VS Code, IntelliJ): log a warning with the conflicting app name if detectable, but proceed (user chose orchestratr as source of truth)
