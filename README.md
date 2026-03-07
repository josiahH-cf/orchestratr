# orchestratr

A cross-platform global hotkey launcher. Maps keyboard chords to app launches, runs as a background daemon with a web-based management GUI, and discovers apps via drop-in config files. Single binary, zero runtime dependencies.

## Features

- **Global hotkey engine** — Ctrl+Space + chord launches or focuses your app in under 200ms
- **Drop-in app discovery** — Place a YAML manifest in `apps.d/` and orchestratr picks it up automatically
- **Web-based management GUI** — Localhost UI for adding/editing apps, viewing hotkeys, and checking status
- **System tray integration** — Status, pause/resume, open GUI, quit
- **Health polling** — `ready_cmd` checks verify your app is responsive after launch
- **Startup diagnostics** — `orchestratr doctor` validates config, binary paths, hotkey conflicts, and platform capabilities
- **Autostart** — Registers as a background daemon on boot (systemd, launchd, Windows registry)

## Platforms

| Platform | Hotkey Backend | Status |
|----------|---------------|--------|
| Linux (X11) | XGrabKey | Implemented |
| Linux (Wayland) | GlobalShortcuts portal + XWayland fallback | Planned |
| macOS | CGEventTap | Implemented |
| Windows | RegisterHotKey + WH_KEYBOARD_LL | Implemented |
| WSL2 | Launched from Windows-side daemon | Implemented |

## Quick Start

### Install

```bash
# Linux / macOS
curl -sSL https://raw.githubusercontent.com/josiahH-cf/orchestratr/main/install.sh | sh

# Windows (PowerShell)
irm https://raw.githubusercontent.com/josiahH-cf/orchestratr/main/install.ps1 | iex
```

### Build from Source

```bash
git clone https://github.com/josiahH-cf/orchestratr.git
cd orchestratr
make build        # produces ./orchestratr binary
make install      # installs to $GOPATH/bin
```

### Run

```bash
orchestratr                # start daemon (foreground)
orchestratr doctor         # run startup diagnostics
orchestratr install        # register autostart
```

### Configure

Edit `configs/example.yml` or drop app manifests into `apps.d/`:

```yaml
# apps.d/myapp.yml
name: myapp
chord: "m"
command: "myapp gui"
environment: native
ready_cmd: "myapp status --json"
ready_timeout_ms: 5000
```

See [docs/CONNECTOR.md](docs/CONNECTOR.md) for the full manifest schema, health checks, and integration guide.

## Architecture

```
cmd/orchestratr/       CLI entrypoint (main, doctor, install)
internal/
  api/                 Localhost HTTP API (port 9876)
  autostart/           Platform-specific autostart registration
  daemon/              Lifecycle, lock files, logging, port management
  gui/                 Web GUI with embedded static assets
  hotkey/              Global hotkey binding (X11, Windows, Darwin)
  launcher/            Process spawning, PID tracking, readiness checks
  platform/            Platform abstraction (accessibility, WSL detection)
  registry/            App config loading, validation, file watching
  tray/                System tray integration
```

## Development

```bash
make test             # go test ./...
make lint             # golangci-lint run ./...
make fmt              # gofmt -w .
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for branch naming, commit format, and PR requirements.

## Related

- [espansr](https://github.com/josiahH-cf/espansr) — Espanso template manager
- [templatr](https://github.com/josiahH-cf/templatr) — Local-model prompt optimizer

## License

[Choose license]
