# Project

- Project name: orchestratr
- Description: A system-wide hotkey launcher and app orchestrator. Background daemon with leader-key chords, localhost HTTP API, and cross-environment launching.
- Primary language/framework: Go (single binary, cross-platform)
- Scope: Hotkey registration, app lifecycle management, IPC via HTTP, system tray, cross-env launching (WSL2 ↔ Windows)
- Non-goals: cloud APIs, multi-tenant, mobile/web, plugin system, package manager distribution (v1)

# Build

- Install: `go install ./...` or `make install`
- Build: `go build -o orchestratr ./cmd/orchestratr`
- Test (all): `go test ./...`
- Test (single): `go test ./internal/hotkey -run TestLeaderKeyCapture`
- Lint: `golangci-lint run ./...`
- Format: `gofmt -w .`
- Type-check: Go compiler handles this natively

# Architecture

- `cmd/orchestratr/` — Main binary entrypoint (daemon start, CLI commands)
- `internal/daemon/` — Background daemon lifecycle, single-instance lock, signal handling
- `internal/hotkey/` — Platform-specific hotkey capture (evdev, CGEventTap, RegisterHotKey, Wayland)
- `internal/registry/` — App registry: YAML config parsing, hot-reload, app state tracking
- `internal/api/` — Localhost HTTP API server (JSON, port 9876)
- `internal/launcher/` — App launching, PID tracking, bring-to-front, cross-env (WSL2)
- `internal/tray/` — System tray icon, minimal status menu
- `internal/gui/` — Management GUI (config editor, app table) — lightweight, infrequent use
- `configs/` — Default/example YAML configs

# Feature Lifecycle

1. **Ideate** — Human files a GitHub issue or describes the feature
2. **Scope** — Agent explores the codebase and writes `/specs/[feature-name].md` using the template
3. **Plan** — Agent decomposes the spec into `/tasks/[feature-name].md` (2–5 tasks)
4. **Test** — Agent writes failing tests for each acceptance criterion
5. **Implement** — Agent makes tests pass, one task per session
6. **Review** — A different agent or human reviews the PR

GitHub Issues are the human intake mechanism. Agents read issues but do not create, edit, or close them.
All agent-driven planning happens in local files (`/specs/`, `/tasks/`, `/decisions/`).

# Communication

- Ask questions as plain text in chat — never use interactive UI elements (buttons, toggles, quick-picks)

# Conventions

- Functions and variables: Go standard (`camelCase` local, `PascalCase` exported)
- Files: lowercase with underscores where needed (e.g., `hotkey_linux.go`)
- Build tags for platform: `//go:build linux`, `//go:build darwin`, `//go:build windows`
- Prefer explicit error handling — `if err != nil { return err }` over silent swallow
- No dead code — remove unused imports, variables, and functions
- Every exported function has a doc comment
- No hardcoded secrets, URLs, or environment-specific values
- Use `internal/` for all non-entrypoint packages (enforced by Go)

# Cross-Platform Strategy

- Platform-specific code isolated in files with build tags: `*_linux.go`, `*_darwin.go`, `*_windows.go`
- Shared interface per platform concern:
  - `hotkey.Listener` — platform-specific hotkey capture
  - `launcher.Executor` — platform-specific process launch and bring-to-front
  - `tray.Provider` — platform-specific system tray
- WSL2 bridging: daemon runs on Windows host; launches WSL apps via `wsl.exe -d <distro> -- <cmd>`

# Testing

- Write tests before implementation
- Place tests alongside source files using `_test.go` naming
- Use table-driven tests where multiple inputs test the same function
- Each acceptance criterion requires at least one test
- Do not modify existing tests to accommodate new code — fix the implementation
- Run the full test suite before committing
- Tests must be deterministic — no flaky tests in the main suite
- Platform-specific tests use build tags; CI matrix covers Linux, macOS, Windows

# Dependencies

- Minimize external dependencies — Go stdlib covers HTTP, JSON, YAML (with a small library), and OS interaction
- Hotkey capture will need platform-specific C interop (cgo) or syscall wrappers
- System tray: evaluate `getlantern/systray` or `fyne.io/systray`
- GUI: evaluate `fyne.io/fyne` for the minimal management interface
- YAML: `gopkg.in/yaml.v3`

# Planning

- Features with more than 3 implementation steps require a written plan
- Plans go in `/tasks/[feature-name].md` or as an ExecPlan per `/.codex/PLANS.md`
- Plans are living documents — update progress, decisions, and surprises as work proceeds
- A plan that cannot fit in 5 tasks indicates the feature should be split. Call this out.
- Small-fix fast path: if a change is <= 3 files and has no behavior change, a full spec/task lifecycle is optional; still document intent in the PR and run lint + relevant tests.

# Commits

- One logical change per commit
- Present-tense imperative subject line, under 72 characters
- Reference the spec or task file in the commit body when applicable
- Commit after each completed task, not after all tasks

# Branches

- Branch from the latest target branch immediately before starting work
- One feature per branch
- Delete after merge
- Never commit directly to the target branch
- Naming: `[type]/[slug]` (e.g., `feat/hotkey-engine`, `fix/pid-tracking`). Include the issue number if one exists: `feat/42-hotkey-engine`

# Worktrees

- Use git worktrees for concurrent features across agents
- Worktree root: `.trees/[branch-name]/`
- Each worktree is isolated: agents operate only within their assigned worktree
- Artifacts (specs, tasks, decisions) live in the main worktree and are shared read-only
- Never switch branches inside a worktree — create a new one

# Pull Requests

- Link to the spec file
- Diff under 300 lines; if larger, split the feature
- All CI checks pass before requesting review
- PR description states: what changed, why, how to verify

# Review

- Reviewable in under 15 minutes
- Tests cover every acceptance criterion
- No unrelated changes in the diff
- Cross-agent review encouraged: use a different model than the one that wrote the code

# Security

- No secrets in code or instruction files
- Use environment variables for all credentials
- Sanitize all external input
- Log security-relevant events
- HTTP API binds to localhost only (127.0.0.1) — never 0.0.0.0

# Agent Boundaries

- Agents do not create or modify GitHub issues, labels, milestones, or projects
- Agents do not push to main/master directly
- Agents do not modify CI/CD workflows without explicit human instruction
- Agents work within local files: specs, tasks, decisions, and source code

# Related Projects

- **espansr** — Espanso template manager (Python/PyQt6). First app to be orchestrated.
  - Connector spec: see `espansr/specs/espansr-orchestratr-connector.md` in the espansr repo
  - Provides `orchestratr.yml` manifest and `espansr status --json` for health checks
