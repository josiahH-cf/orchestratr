# Build, Test & Lint Commands

> Referenced from `AGENTS.md`. This is part of the canonical workflow — see `/governance/REGISTRY.md`.

These values are set with initial values during Phase 1 (initialization) and finalized during Phase 4 (Scaffold Project) based on architecture reasoning.

## Core Commands

| Command | Value |
|---------|-------|
| Install | `go mod download` |
| Build | `make build` |
| Test (all) | `make test` (`go test ./...`) |
| Test (single) | `go test -run <TestName> ./path/to/package` |
| Lint | `make lint` (`golangci-lint run ./...`) |
| Format | `make fmt` (`gofmt -w .`) |
| Type-check | N/A (Go compiler handles this) |
| Lint (workflow) | `scripts/workflow-lint.sh` |
| Review Bot | `/review-bot` (Claude) or `phase-7a-review-bot.prompt.md` (Copilot) |

## Code Conventions

**Language:** Go 1.22+ (module: `github.com/josiahH-cf/orchestratr`)

**Naming:**
- Functions and variables: `camelCase` (exported: `PascalCase`) per Go convention
- Files and directories: `snake_case` with build-tag suffixes (e.g., `listener_linux.go`, `registry_windows.go`)

**Architecture:**
- `cmd/orchestratr/` — CLI entrypoint (main, doctor, install subcommands)
- `internal/api/` — HTTP API server and state management
- `internal/autostart/` — Platform-specific autostart registration
- `internal/daemon/` — Daemon lifecycle, lock files, logging, port management
- `internal/gui/` — Web-based GUI with embedded static assets
- `internal/hotkey/` — Global hotkey binding (X11, Windows, Darwin)
- `internal/launcher/` — App process spawning, PID tracking, readiness checks
- `internal/platform/` — Platform abstraction (accessibility, WSL detection)
- `internal/registry/` — App configuration loading, validation, file watching
- `internal/tray/` — System tray integration
- `configs/` — Example configuration files
- `specs/` — Feature specifications
- `tasks/` — Implementation task files
- `decisions/` — Architecture Decision Records
