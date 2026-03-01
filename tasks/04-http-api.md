# Tasks: HTTP API & IPC Protocol

**Spec:** /specs/04-http-api.md

## Status

- Total: 4
- Complete: 0
- Remaining: 4

## Task List

### Task 1: Server foundation, localhost middleware, and /health

- **Files:** `internal/api/server.go`, `internal/api/responses.go`, `internal/api/server_test.go`
- **Done when:** Server binds to 127.0.0.1, localhost middleware rejects non-local requests with 403, GET /health returns `{"status":"ok","version":"<version>"}`, JSON error responses use consistent format `{"error":"<message>","code":"<error_code>"}`
- **Criteria covered:** Localhost binding, /health endpoint, error format, non-localhost rejection
- **Status:** [ ] Not started

### Task 2: Registry and reload endpoints

- **Files:** `internal/api/routes.go`, `internal/api/routes_test.go`
- **Done when:** GET /apps returns app registry as JSON array, POST /reload triggers config hot-reload and returns new registry or validation errors as JSON
- **Criteria covered:** /apps endpoint, /reload endpoint
- **Status:** [ ] Not started

### Task 3: App lifecycle endpoints

- **Files:** `internal/api/state.go`, `internal/api/state_test.go`, `internal/api/routes.go` (extend), `internal/api/routes_test.go` (extend)
- **Done when:** POST /apps/{name}/launched records app launch, POST /apps/{name}/ready records app readiness, unknown app names return 404 with error JSON
- **Criteria covered:** /apps/{name}/launched, /apps/{name}/ready
- **Status:** [ ] Not started

### Task 4: CLI integration and port discovery

- **Files:** `cmd/orchestratr/main.go`, `internal/daemon/health.go` (refactor)
- **Done when:** New API server replaces old HealthServer in daemon startup, port discovery file is written/cleaned up, existing tests still pass
- **Criteria covered:** Port discovery file, daemon integration
- **Status:** [ ] Not started
