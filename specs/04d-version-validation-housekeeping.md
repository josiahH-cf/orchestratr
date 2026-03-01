# Feature: Version & Validation Housekeeping

**Status:** Not started
**Parent:** /specs/03-app-registry.md, /specs/04-http-api.md
**Project:** orchestratr

## Description

Several small consistency and future-proofing issues across the registry and API: the version string is hardcoded in multiple places, the `environment` field has inconsistent empty-string handling between CLI and API, and `api_port: 0` in a user config silently creates an unpredictable port. This spec bundles these small fixes.

## Acceptance Criteria

- [ ] The version string is defined in exactly **one** place (a `var Version` in `cmd/orchestratr/main.go` or a shared package) and used by both the `version` command and `api.NewServer`
- [ ] The `Version` variable is overridable via `go build -ldflags "-X main.Version=v1.0.0"` (or equivalent package path)
- [ ] When `environment` is empty in a loaded config, it is normalized to `"native"` by the `Load()` function so all consumers (CLI, API, launcher) see the same value
- [ ] `ValidateConfig` emits a warning (not a hard error) when `api_port` is `0`, logged at startup — the daemon still starts but the user is informed this means a random port
- [ ] Existing tests continue to pass; new behaviors have tests

## Affected Areas

| Area | Files |
|------|-------|
| **Modify** | `cmd/orchestratr/main.go` — extract version to `var`, use in `version` command and `api.NewServer` call |
| **Modify** | `internal/registry/loader.go` — normalize empty `environment` to `"native"` after `yaml.Unmarshal` in `Load()` |
| **Modify** | `internal/registry/loader_test.go` — test that empty environment is normalized |
| **Modify** | `internal/registry/validate.go` — add `api_port: 0` warning to `ValidateConfig` return |
| **Modify** | `internal/registry/validate_test.go` — test for port 0 warning |

## Constraints

- Environment normalization happens in `Load()`, **not** in `ValidateConfig()` — validation should see the canonical value
- The `api_port: 0` check returns a warning, not a blocking error. Consider using a separate return type or adding a `Warnings` field. Alternatively, the check can return a non-fatal error that `runStart()` logs but does not treat as a failure. Keep the approach simple — logging the warning at daemon startup from `runStart()` is sufficient.
- Do not change the default value of `environment` in the YAML struct tag (that would add `environment: native` to every serialized config, even for apps that don't set it)

## Out of Scope

- Automated release builds or CI `-ldflags` configuration
- Versioned API paths (e.g., `/v1/health`)

## Dependencies

- None — all changes are to existing code

## Notes

### Version variable pattern

```go
// cmd/orchestratr/main.go
// Version is set at build time via -ldflags.
var Version = "v0.0.0-dev"

// In run():
case "version":
    fmt.Fprintf(stdout, "orchestratr %s\n", Version)

// In runStart():
apiSrv := api.NewServer(apiPort, Version, reg, reloadFn)
```

### Environment normalization in `Load()`

After `yaml.Unmarshal`, iterate over `cfg.Apps` and normalize:

```go
for i := range cfg.Apps {
    if cfg.Apps[i].Environment == "" {
        cfg.Apps[i].Environment = "native"
    }
}
```

This ensures the API's `GET /apps` response, the CLI's `orchestratr list` table, and the future launcher all see `"native"` instead of `""`.

### Port 0 warning

Keep it simple — check in `runStart()` after loading config:

```go
if cfg.APIPort == 0 {
    logger.Println("warning: api_port is 0; a random port will be assigned each start")
}
```

Alternatively, add it to `ValidateConfig` as a non-fatal diagnostic. The caller (`LoadAndValidate`) would need to distinguish warnings from errors. Simplest approach: just check in `runStart()`.
