# Feature: Drop-In App Discovery (`apps.d/`)

**Status:** Implemented
**Project:** orchestratr

## Description

orchestratr currently reads app registrations only from its single `config.yml` file. Apps like espansr and templatr generate connector manifests, but orchestratr has no mechanism to discover them. This spec adds a drop-in directory (`apps.d/`) that orchestratr scans at startup and on hot-reload. Each `*.yml` file in the directory is a single app manifest using the same flat schema as a `config.yml` app entry. This is the systemd/nginx `conf.d/` pattern — any app can self-register by writing one file.

### Current State (Context for Implementation)

- **Registry loader** (`internal/registry/loader.go`): `Load()` reads a single YAML path via `DefaultConfigPath()` → `~/.config/orchestratr/config.yml`. No directory scanning, no glob, no multi-source merge.
- **Config struct** (`internal/registry/config.go`): `Config` has a flat `Apps []AppEntry` list. No concept of sources or includes.
- **Registry** (`internal/registry/registry.go`): Thread-safe wrapper around a single `Config`. `Swap()` replaces atomically. No multi-source merging.
- **Watcher** (`internal/registry/watcher.go`): Watches **one file** (`w.path`), not a directory. Uses fsnotify with debounce.
- **Validation** (`internal/registry/validate.go`): Validates required fields (`name`, `chord`, `command`), chord uniqueness, chord format, and environment values. No source-attribution in error messages.
- **Web GUI** (`internal/gui/static/index.html`): Reads/writes `config.yml` via `/api/config`. Shows app table with CRUD. No concept of read-only or externally-managed apps.
- **CLI** (`cmd/orchestratr/main.go`): `orchestratr list` prints apps from the registry. No source column.

### Manifest Schema

Each file in `apps.d/` is a single YAML document with the **exact same flat fields** as a `config.yml` app entry:

```yaml
# apps.d/espansr.yml
name: espansr
chord: "e"
command: "espansr gui"
environment: wsl
description: "Espanso template manager"
ready_cmd: "espansr status --json"
ready_timeout_ms: 3000
```

This is intentionally identical to what you'd write inside `config.yml`'s `apps:` list. No translation layer, no nested objects. A manifest can be copy-pasted into `config.yml` and vice versa.

## Acceptance Criteria

- [ ] `apps.d/*.yml` files are loaded and merged into the registry at startup
- [ ] The `apps.d/` directory is created alongside `config.yml` on first run (via `EnsureDefaults` or equivalent)
- [ ] File watcher detects creation, modification, and deletion of files in `apps.d/` and triggers hot-reload
- [ ] Duplicate chord between `config.yml` and `apps.d/`: `config.yml` entry wins, warning logged naming both sources
- [ ] Duplicate chord between two `apps.d/` files: rejected with a clear error naming both files
- [ ] Duplicate app name between `config.yml` and `apps.d/`: `config.yml` entry wins, drop-in is skipped with warning
- [ ] `orchestratr list` displays a source column (`config` vs `apps.d/<filename>`)
- [ ] Web GUI shows source attribution per app and marks drop-in apps as read-only (no Edit/Remove buttons; tooltip says "Managed by <app name>. Edit via that app's settings or the file directly.")
- [ ] Invalid YAML in an `apps.d/` file is logged as an error and skipped — does not prevent other apps or `config.yml` from loading
- [ ] An empty `apps.d/` directory or missing directory does not cause errors

## Affected Areas

| Area | Files |
|------|-------|
| **Modify** | `internal/registry/loader.go` — add `LoadAppsDir()`, `EnsureAppsDir()`, integrate into `Load()` |
| **Modify** | `internal/registry/config.go` — add `Source string` field to `AppEntry` (not serialized to YAML; internal tracking only) |
| **Modify** | `internal/registry/validate.go` — source-attributed error messages for conflicts, merge-aware validation |
| **Modify** | `internal/registry/watcher.go` — watch `apps.d/` directory in addition to `config.yml` |
| **Modify** | `internal/registry/registry.go` — no structural changes, but `Swap()` callers pass merged config |
| **Modify** | `cmd/orchestratr/main.go` — `list` subcommand displays source column |
| **Modify** | `internal/gui/static/index.html` — source badge per row, read-only for drop-in apps |
| **Modify** | `internal/api/server.go` — `/api/config` GET response includes source attribution |

## Constraints

- Manifest schema is **identical** to `AppEntry` YAML — no special fields, no envelope, no wrapper
- `config.yml` always wins on conflict (user override principle)
- Drop-in files are **not writable** by orchestratr — the owning app manages them
- File watcher debounce interval should match the existing `config.yml` watcher (currently 500ms–1s in `watcher.go`)
- Directory must be created with user-only permissions (`0700`) on Unix
- On Windows: `%APPDATA%\orchestratr\apps.d\`; on Linux/macOS: `~/.config/orchestratr/apps.d/`

## Out of Scope

- Recursive subdirectory scanning (flat `apps.d/` only)
- Non-YAML formats (JSON, TOML)
- App health monitoring (see `ready-cmd-health-polling.md`)
- Remote/network config sources
- Writing or modifying drop-in files from orchestratr's GUI (apps own their files)

## Dependencies

- None — this is a foundational change that other specs depend on

## Notes

### Merge algorithm

```
1. Load config.yml → base_apps[]
2. Glob apps.d/*.yml → for each file:
   a. Parse as single AppEntry
   b. Set entry.Source = "apps.d/<filename>"
   c. If entry.Name conflicts with base_apps → skip, warn
   d. If entry.Chord conflicts with base_apps → skip, warn
   e. If entry.Name or Chord conflicts with another drop-in → reject both, error
   f. Append to merged_apps[]
3. Final registry = base_apps + merged_apps
4. Run existing validation on the combined list
```

### Source tracking

Add an unexported `source` field to `AppEntry`:

```go
type AppEntry struct {
    Name           string `yaml:"name" json:"name"`
    Chord          string `yaml:"chord" json:"chord"`
    Command        string `yaml:"command" json:"command"`
    Environment    string `yaml:"environment,omitempty" json:"environment,omitempty"`
    Description    string `yaml:"description,omitempty" json:"description,omitempty"`
    ReadyCmd       string `yaml:"ready_cmd,omitempty" json:"ready_cmd,omitempty"`
    ReadyTimeoutMs int    `yaml:"ready_timeout_ms,omitempty" json:"ready_timeout_ms,omitempty"`
    Detached       bool   `yaml:"detached,omitempty" json:"detached,omitempty"`
    Source         string `yaml:"-" json:"source,omitempty"` // internal: "config" or "apps.d/foo.yml"
}
```

The `yaml:"-"` tag ensures it's never written to YAML files. The `json:"source"` tag exposes it via the API for the web GUI.

### Platform paths

| Platform | `apps.d/` location |
|----------|-------------------|
| Linux | `~/.config/orchestratr/apps.d/` |
| macOS | `~/Library/Application Support/orchestratr/apps.d/` |
| Windows | `%APPDATA%\orchestratr\apps.d\` |

These must use the same platform detection as `DefaultConfigPath()` in `loader.go`.
