# Contributing to orchestratr

## Prerequisites

- Go 1.22+
- `golangci-lint` (for linting)
- CGO enabled (required for X11/systray on Linux)

## Development Workflow

```bash
make build        # compile binary
make test         # run all tests
make lint         # run linter
make fmt          # format all Go files
```

## Branch Naming

Format: `<agent>/type-short-description`

- **type:** `feat`, `bug`, `refactor`, `chore`, `docs`
- **short-description:** 2–4 word kebab-case summary

Examples: `josiah/feat-new-hotkey`, `claude/bug-login-crash`

## Commit Format

Use conventional commit prefixes:

- `feat(scope):` — new feature
- `fix(scope):` — bug fix
- `docs(scope):` — documentation only
- `refactor(scope):` — code restructuring
- `chore(scope):` — build, CI, tooling

Scope is the package name (e.g., `hotkey`, `launcher`, `tray`).

## Code Conventions

- **Language:** Go 1.22+, module `github.com/josiahH-cf/orchestratr`
- **Naming:** `camelCase` (unexported), `PascalCase` (exported) per Go convention
- **Files:** `snake_case` with build-tag suffixes (e.g., `listener_linux.go`)
- **Architecture:** `cmd/` for CLI, `internal/` for all library code — see README for package layout
- **Tests:** `_test.go` files alongside source, using Go's `testing` package
- **TDD:** Write failing tests before implementation

## Pull Requests

- State what changed, why, and how to verify
- Link the feature spec if applicable
- List files changed grouped by concern
- All tests must pass on Linux, macOS, and Windows before merge

## Project Structure

See [README.md](README.md) for the full architecture layout and [docs/CONNECTOR.md](docs/CONNECTOR.md) for the app integration protocol.
