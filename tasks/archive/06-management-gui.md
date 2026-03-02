# Tasks: 06-management-gui

**Spec:** /specs/06-management-gui.md
**Decision:** /decisions/0003-web-based-gui.md

## Status

- Total: 4
- Complete: 4
- Remaining: 0

## Task List

### Task 1: GUI web server, static embed, and browser open

- **Files:** `internal/gui/gui.go`, `internal/gui/gui_test.go`, `internal/gui/static/index.html`
- **Done when:** `gui.Start(cfgPath, apiPort)` starts a localhost HTTP server on an ephemeral port, serves embedded static files at `/`, opens default browser via `xdg-open`/`open`/`start`, `gui.Stop()` shuts it down; headless-friendly (prints URL if no browser)
- **Criteria covered:** GUI opens from tray, separate window (browser tab), closes without affecting daemon
- **Status:** [x] Done

### Task 2: Config JSON API and HTML app table

- **Files:** `internal/gui/gui.go`, `internal/gui/gui_test.go`, `internal/gui/static/index.html`
- **Done when:** `GET /api/config` returns current config as JSON; `PUT /api/config` saves config to YAML and triggers daemon reload; HTML page renders app table with Name, Chord, Command, Env, Status columns; live status fetched from daemon API when running
- **Criteria covered:** app table display, save writes YAML + hot-reload
- **Status:** [x] Done

### Task 3: Add/Edit/Remove app UI with chord validation

- **Files:** `internal/gui/static/index.html` (JS logic)
- **Done when:** Add App form with name, chord, command, environment dropdown, description; Edit pre-populates form; Remove with confirmation; chord field validates uniqueness in real-time (JS-side check against current config)
- **Criteria covered:** add app form, edit/remove actions, chord uniqueness validation
- **Status:** [x] Done

### Task 4: Settings panel and daemon integration

- **Files:** `internal/gui/static/index.html`, `cmd/orchestratr/main.go`
- **Done when:** Leader key and timeout editable in settings section; Save sends PUT /api/config; wire `tray.OnConfigure` and add `orchestratr configure` CLI command that starts GUI server and opens browser
- **Criteria covered:** leader key setting, GUI independent of daemon, tray integration
- **Status:** [x] Done

## Test Strategy

| Criterion | Tested by |
|-----------|-----------|
| GUI opens from tray Configure | Task 4: tray callback wiring |
| App table columns | Task 2: `GET /api/config` handler test + HTML |
| Add/Edit/Remove actions | Task 3: JS validation + `PUT /api/config` test |
| Chord uniqueness validation | Task 3: real-time JS validation |
| Leader key setting | Task 4: settings panel |
| Save writes YAML + triggers reload | Task 2: `TestPutConfig_WritesYAML`, `TestPutConfig_TriggersReload` |
| GUI closes without affecting daemon | Task 1: server lifecycle test |

## Session Log

