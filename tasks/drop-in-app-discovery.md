# Tasks: Drop-In App Discovery (`apps.d/`)

**Spec:** `/specs/drop-in-app-discovery.md`
**Branch:** `feat/drop-in-app-discovery`

## Task 1: Schema + Loader (config.go, loader.go)

- Add `Source string` field to `AppEntry` with `yaml:"-" json:"source,omitempty"` tags
- Add `AppsDirPath()` helper (platform-appropriate `apps.d/` directory)
- Add `EnsureAppsDir()` to create the directory on first run
- Add `LoadAppsDir(dir string) ([]AppEntry, []error)` — globs `*.yml`, parses each as a single `AppEntry`
- Add `LoadWithDropins(configPath string) (*Config, []error)` — merge algorithm per spec
- Write failing tests first, then implement

## Task 2: Validation + Merge Conflicts (validate.go)

- Update `ValidateConfig` to include source-attributed error messages
- Add `ValidateMerge(base []AppEntry, dropins []AppEntry) ([]AppEntry, []error)` — implements the merge conflict rules:
  - Duplicate name/chord between config.yml and apps.d → config wins, warning
  - Duplicate name/chord between two apps.d files → reject both, error
- Write failing tests first

## Task 3: Watcher + Hot-Reload (watcher.go, main.go)

- Extend `Watcher` to watch the `apps.d/` directory in addition to `config.yml`
- Trigger reload on file create/modify/delete in `apps.d/`
- Update `runStart()` reload function to use `LoadWithDropins`
- Update `EnsureDefaults` to also call `EnsureAppsDir`
- Write failing tests first

## Task 4: CLI + API + GUI (main.go, server.go, index.html)

- `orchestratr list` displays a SOURCE column (`config` vs `apps.d/filename.yml`)
- API `/apps` response includes `source` field (already via JSON tag)
- Web GUI: source badge per row, read-only for drop-in apps (no Edit/Remove)
- Tooltip: "Managed by <source>. Edit via that app's settings or the file directly."

## Progress

- [x] Task 1
- [x] Task 2
- [x] Task 3
- [x] Task 4
