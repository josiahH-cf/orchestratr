# Feature: App Lifecycle API Endpoints

**Status:** Archived
**Parent:** /specs/04-http-api.md
**Project:** orchestratr

## Description

The `StateTracker` records `launched` and `ready` state per app, but this data is not queryable via the API, and there is no way for an app to signal that it has stopped. This spec adds the missing lifecycle endpoints needed by the management GUI (spec 06) and connected apps.

## Acceptance Criteria

- [ ] `GET /apps/{name}/state` returns the app's lifecycle state as JSON: `{"name":"...","launched":bool,"ready":bool,"launched_at":"...","ready_at":"..."}`
- [ ] `GET /apps/{name}/state` returns 404 with standard error JSON if the app is not in the registry
- [ ] `GET /apps/{name}/state` returns 200 with `{"name":"...","launched":false,"ready":false}` if the app is registered but has no lifecycle state yet
- [ ] `POST /apps/{name}/stopped` marks the app as not launched and not ready, clearing timestamps
- [ ] `POST /apps/{name}/stopped` returns 200 with `{"status":"ok","app":"...","state":"stopped"}`
- [ ] `POST /apps/{name}/stopped` returns 404 if the app is not in the registry
- [ ] The `StateTracker` has a `SetStopped(name string)` method that resets launched/ready state
- [ ] All new endpoints follow the existing error format (`{"error":"...","code":"..."}`) and method enforcement (405 with Allow header)

## Affected Areas

| Area | Files |
|------|-------|
| **Modify** | `internal/api/state.go` — add `SetStopped()` method |
| **Modify** | `internal/api/state_test.go` — tests for `SetStopped` |
| **Modify** | `internal/api/server.go` — add route handling for `state` and `stopped` actions in `handleAppAction` |
| **Modify** | `internal/api/server_test.go` or `internal/api/state_test.go` — HTTP handler tests |

## Constraints

- `GET /apps/{name}/state` must work for registered apps that have never called `launched` or `ready` — return a zero-value state object, not 404
- `SetStopped` should be idempotent — calling it on an already-stopped or never-launched app is a no-op (no error)
- Maintain the pattern in `handleAppAction` where the action string is switched: add `case "state"` (GET) and `case "stopped"` (POST)

## Out of Scope

- Automatic stop detection via PID tracking (spec 05 concern)
- WebSocket/SSE streaming of state changes (v2)
- Aggregate state endpoint (e.g., `GET /state` for all apps) — can be added later if needed

## Dependencies

- `internal/api/state.go` — existing `StateTracker`, `AppState` structs
- `internal/api/server.go` — existing `handleAppAction` routing

## Notes

### Method routing in `handleAppAction`

Currently, `handleAppAction` rejects all non-POST requests. With the addition of `GET /apps/{name}/state`, the method enforcement needs to become action-specific:

```go
switch action {
case "launched":
    // POST only (existing)
case "ready":
    // POST only (existing)
case "stopped":
    // POST only (new)
case "state":
    if r.Method != http.MethodGet {
        methodNotAllowed(w, http.MethodGet)
        return
    }
    // handle GET
default:
    writeError(w, http.StatusNotFound, "not_found", ...)
}
```

The initial `if r.Method != http.MethodPost` guard at the top of `handleAppAction` must be replaced with per-action method checks.

### `SetStopped` design

```go
func (st *StateTracker) SetStopped(name string) {
    st.mu.Lock()
    defer st.mu.Unlock()

    state, ok := st.apps[name]
    if !ok {
        return // no-op for unknown apps
    }
    state.Launched = false
    state.Ready = false
    state.LaunchedAt = nil
    state.ReadyAt = nil
}
```

### State response for untracked apps

When `GET /apps/{name}/state` is requested for an app that is in the registry but has no `StateTracker` entry, construct a default response:

```go
appState := s.state.Get(name)
if appState == nil {
    appState = &AppState{Name: name}
}
writeJSON(w, http.StatusOK, appState)
```
