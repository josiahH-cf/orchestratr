# Feature: Remove Dead HealthServer Code

**Status:** Not started
**Parent:** /specs/04-http-api.md
**Project:** orchestratr

## Description

The `daemon.HealthServer` was the original minimal health endpoint created in spec 01. Spec 04 replaced it with the full `api.Server` which includes `/health` and all other endpoints. `HealthServer` is no longer referenced from `cmd/orchestratr/main.go` or any other production code — only its own test file still exercises it. Per AGENTS.md: "No dead code."

## Acceptance Criteria

- [ ] `internal/daemon/health.go` is deleted
- [ ] `internal/daemon/health_test.go` is deleted
- [ ] `DefaultPortFilePath()`, `WritePortFile()`, and `RemovePortFile()` are **preserved** (moved to their own file if currently only in `health.go`) — these are actively used by `runStart()` and `readPort()`
- [ ] `go build ./...` succeeds with no references to `HealthServer`
- [ ] `go test ./...` passes (no import or reference errors)

## Affected Areas

| Area | Files |
|------|-------|
| **Delete** | `internal/daemon/health.go` — `HealthServer` struct and methods |
| **Delete** | `internal/daemon/health_test.go` — tests for `HealthServer` |
| **Create or Modify** | `internal/daemon/port.go` (new) — move `DefaultPortFilePath`, `WritePortFile`, `RemovePortFile` if they're in `health.go` |

## Constraints

- Do **not** remove or alter the port file functions — they are used in `cmd/orchestratr/main.go` (`runStart`, `readPort`)
- Do **not** remove the test for port file functions — either keep them in a renamed test file or write new tests in `port_test.go`

## Out of Scope

- Changing the `api.Server` implementation
- Modifying the daemon startup flow

## Dependencies

- `/specs/04-http-api.md` — `api.Server` already handles `/health`

## Notes

### What's in `health.go`

The file contains two groups of code:

1. **Dead code** (remove): `HealthServer` struct, `NewHealthServer`, `ListenAddr`, `Start`, `Stop`, `Port`, `handleHealth`
2. **Live code** (preserve): `DefaultPortFilePath`, `WritePortFile`, `RemovePortFile`

The live functions should be moved to a new `internal/daemon/port.go` file. Create a corresponding `port_test.go` if there are existing tests for the port functions in `health_test.go` (there aren't currently — the health tests only test the HTTP server).

### Verification

After deletion, run:
```bash
grep -r "HealthServer" internal/ cmd/
go build ./...
go test ./...
```
All should show zero references and pass cleanly.
