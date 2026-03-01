# Feature: `orchestratr reload` CLI Command

**Status:** Not started
**Parent:** /specs/03-app-registry.md, /specs/04-http-api.md
**Project:** orchestratr

## Description

The usage string advertises `orchestratr reload` but the command is not implemented. This spec adds a CLI command that sends `POST /reload` to the running daemon's API (mirroring the pattern used by `orchestratr trigger`), completing the CLI ↔ API symmetry.

## Acceptance Criteria

- [ ] `orchestratr reload` sends `POST /reload` to the running daemon via the port discovery file and prints the result
- [ ] On success, outputs a confirmation message including the number of apps in the reloaded config (e.g., "config reloaded (3 apps)")
- [ ] On reload failure (invalid config), prints the validation error message from the API response and exits with a non-zero code
- [ ] When the daemon is not running (port file missing or unreachable), prints "daemon is not running" and exits with a non-zero code
- [ ] The `switch` in `run()` includes `case "reload"` (currently missing — `reload` falls through to "unknown command")
- [ ] Tests cover: success path, daemon-not-running path, and validation-error path

## Affected Areas

| Area | Files |
|------|-------|
| **Modify** | `cmd/orchestratr/main.go` — add `case "reload"` and `runReload()` function |
| **Modify** | `cmd/orchestratr/main_test.go` — add tests for `runReload` via `run()` |

## Constraints

- Follow the same pattern as `runTrigger()`: read port from discovery file, make HTTP request, parse JSON response
- Use the existing `readPort()` helper
- Parse the API's `ErrorResponse` JSON on non-200 responses to extract a meaningful message

## Out of Scope

- Changing the behavior of `POST /reload` on the server side (already working)
- File watcher integration (see `/specs/03a-registry-daemon-integration.md`)

## Dependencies

- `POST /reload` API endpoint (already implemented in `internal/api`)
- Port discovery file written by daemon (already implemented)

## Notes

### Implementation sketch

```go
func runReload(stdout, stderr io.Writer) error {
    port, err := readPort()
    if err != nil {
        return fmt.Errorf("daemon is not running: %w", err)
    }

    url := fmt.Sprintf("http://127.0.0.1:%d/reload", port)
    resp, err := http.Post(url, "application/json", nil)
    if err != nil {
        return fmt.Errorf("sending reload: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusOK {
        var result struct {
            Status string `json:"status"`
            Apps   []any  `json:"apps"`
        }
        if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
            fmt.Fprintf(stdout, "config reloaded (%d apps)\n", len(result.Apps))
        } else {
            fmt.Fprintln(stdout, "config reloaded")
        }
        return nil
    }

    var errResp struct {
        Error string `json:"error"`
    }
    if decErr := json.NewDecoder(resp.Body).Decode(&errResp); decErr == nil && errResp.Error != "" {
        return fmt.Errorf("reload failed: %s", errResp.Error)
    }
    return fmt.Errorf("reload failed: HTTP %d", resp.StatusCode)
}
```

### Test approach

Tests for `runReload` should not require a running daemon. Testing through `run([]string{"reload"}, ...)` with `ORCHESTRATR_PORT_PATH` pointing to a non-existent file covers the not-running case. For success/error cases, start an `api.Server` in the test (port 0), write its port to a temp file, and set `ORCHESTRATR_PORT_PATH`. This mirrors the pattern in `TestRun_TriggerNotRunning`.
