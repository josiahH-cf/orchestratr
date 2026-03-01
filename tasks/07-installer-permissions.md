# Tasks: Installer & Permissions

**Spec:** /specs/07-installer-permissions.md

## Status

- Total: 4
- Complete: 0
- Remaining: 4

## Task List

### Task 1: Autostart package — cross-platform autostart management

- **Files:** `internal/autostart/autostart.go`, `internal/autostart/autostart_test.go`, `internal/autostart/linux.go`, `internal/autostart/linux_test.go`, `internal/autostart/darwin.go`, `internal/autostart/darwin_test.go`, `internal/autostart/windows.go`, `internal/autostart/windows_test.go`, `internal/autostart/stub.go`
- **Done when:** `AutostartManager` interface with `Install`, `Uninstall`, and `IsInstalled` methods is implemented for all platforms; unit tests verify idempotent install/uninstall and correct file/content generation for systemd user service (Linux), Launch Agent plist (macOS), and registry key (Windows).
- **Criteria covered:** Autostart at login (systemd, Launch Agent, registry), idempotent installs, uninstall removes autostart config
- **Status:** [ ] Not started

### Task 2: Platform detection — WSL2 detection and macOS accessibility check

- **Files:** `internal/platform/platform.go`, `internal/platform/platform_test.go`, `internal/platform/wsl.go`, `internal/platform/wsl_test.go`, `internal/platform/accessibility_darwin.go`, `internal/platform/accessibility_stub.go`, `internal/platform/accessibility_test.go`
- **Done when:** `IsWSL2()` correctly detects WSL2 environment (via `/proc/version`); `CheckAccessibility()` returns whether macOS Accessibility permission is granted (stub returns true on non-macOS); unit tests cover detection logic with mock data.
- **Criteria covered:** WSL2 detection and warning, macOS Accessibility permission detection
- **Status:** [ ] Not started

### Task 3: Install/uninstall CLI commands

- **Files:** `cmd/orchestratr/main.go`, `cmd/orchestratr/install.go`, `cmd/orchestratr/install_test.go`
- **Done when:** `orchestratr install` configures autostart, verifies hotkey registration, detects WSL2, checks macOS accessibility, and prints actionable warnings; `orchestratr uninstall` removes autostart config and prompts for config directory removal; both commands are idempotent; tests verify output and behavior for each scenario.
- **Criteria covered:** `orchestratr install`/`orchestratr uninstall` CLI, hotkey registration verification, WSL2 warning, macOS Accessibility prompt, idempotent installs
- **Status:** [ ] Not started

### Task 4: Install scripts — install.sh and install.ps1

- **Files:** `install.sh`, `install.ps1`
- **Done when:** `install.sh` builds the Go binary and runs `orchestratr install` (Linux/macOS); `install.ps1` builds the Go binary and runs `orchestratr install` (Windows); scripts are idempotent and detect missing prerequisites (Go toolchain).
- **Criteria covered:** `install.sh` installs on Linux/macOS, `install.ps1` installs on Windows
- **Status:** [ ] Not started

## Test Strategy

| Criterion | Tested by |
|-----------|-----------|
| Autostart at login (systemd, Launch Agent, registry) | Task 1 unit tests |
| Idempotent installs | Task 1 + Task 3 tests |
| Uninstall removes autostart config | Task 1 + Task 3 tests |
| macOS Accessibility detection | Task 2 unit tests |
| WSL2 detection and warning | Task 2 + Task 3 tests |
| Hotkey registration verification | Task 3 tests (uses existing `Listener.Register`) |
| `orchestratr install` / `uninstall` CLI | Task 3 integration tests |
| `install.sh` / `install.ps1` | Task 4 (manual verification — shell scripts) |

## Session Log

<!-- Append after each session: date, completed, blockers -->
