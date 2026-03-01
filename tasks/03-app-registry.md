# Tasks: App Registry & Configuration

**Spec:** /specs/03-app-registry.md

## Status

- Total: 4
- Complete: 4
- Remaining: 0

## Task List

### Task 1: Data model & validation

- **Files:** `internal/registry/config.go`, `internal/registry/validate.go`, `internal/registry/registry.go`, `internal/registry/config_test.go`, `internal/registry/validate_test.go`, `internal/registry/registry_test.go`
- **Done when:** Config/AppEntry structs parse YAML correctly; validation rejects duplicates, reserved chords, missing fields; Registry provides thread-safe query methods
- **Criteria covered:** App entry fields, duplicate chord rejection, reserved chords, invalid config handling
- **Status:** [x] Complete

### Task 2: Config loading & defaults

- **Files:** `internal/registry/loader.go`, `internal/registry/loader_test.go`
- **Done when:** `Load()` reads YAML from disk; `DefaultConfigPath()` returns platform path; `EnsureDefaults()` creates config with sensible defaults if missing
- **Criteria covered:** Platform-appropriate config directory, config created with defaults on first run, clear error messages
- **Status:** [x] Complete

### Task 3: File watcher & hot-reload

- **Files:** `internal/registry/watcher.go`, `internal/registry/watcher_test.go`
- **Done when:** Watcher detects config file changes, debounces rapid writes, reloads valid config atomically, keeps old config on invalid change
- **Criteria covered:** Hot-reload without restart
- **Status:** [x] Complete

### Task 4: CLI `list` command

- **Files:** `cmd/orchestratr/main.go`, `cmd/orchestratr/main_test.go`
- **Done when:** `orchestratr list` loads config from default path and prints a formatted table of registered apps
- **Criteria covered:** `orchestratr list` CLI command
- **Status:** [x] Complete

## Test Strategy

| Criterion | Tested in |
|-----------|-----------|
| Platform config directory | Task 2: `loader_test.go` |
| App entry fields | Task 1: `config_test.go` |
| Duplicate chord rejection | Task 1: `validate_test.go` |
| Hot-reload | Task 3: `watcher_test.go` |
| `orchestratr list` | Task 4: `main_test.go` |
| Defaults on first run | Task 2: `loader_test.go` |
| Invalid config handling | Task 1: `validate_test.go` |

## Session Log

- **2026-03-01**: All 4 tasks implemented and passing. Review completed — fixed `equalFold` Unicode bug (C1), `os.Stat` error handling (C2), watcher error wrapping (N7), replaced `unwrapPathError` with `errors.Is` (N4), `LoadAndValidate` returns nil on validation error (W6).
