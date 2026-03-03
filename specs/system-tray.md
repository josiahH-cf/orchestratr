# Feature: Native System Tray

**Status:** Implemented
**Project:** orchestratr

## Description

orchestratr's tray is currently a no-op `HeadlessProvider` on every platform. This spec replaces it with a real `SystrayProvider` on Linux and Windows using `fyne.io/systray` — a pure-Go tray library (CGo on Linux for GTK/AppIndicator, syscall-only on Windows). The tray gives users visible daemon status, Pause/Resume/Quit/Configure actions, and error notifications without keeping a terminal open. macOS is out of scope here; it is tied to the macOS platform layer spec.

### Current State

- `internal/tray/tray.go`: `Provider` interface + `HeadlessProvider` (no-op). No platform implementations.
- `cmd/orchestratr/main.go` line 307: hardcodes `&tray.HeadlessProvider{}`. The `OnPause`, `OnResume`, `OnQuit`, and `OnConfigure` callbacks are all registered but never invoked because the headless provider discards them.
- `NotifyError` is called on launch failure in `main.go` (line 396) but produces no visible output.
- `tray.Provider` interface is already correct and complete — no changes needed.

### Library

`fyne.io/systray` (v1.11.0+):
- On Windows: pure Win32 via `golang.org/x/sys/windows` — no CGo
- On Linux: CGo link against `libappindicator3` (or `libayatana-appindicator3-dev`)
- Provides: `systray.Run(onReady, onExit)`, `AddMenuItem`, `SetIcon`, `SetTooltip`, `Quit`
- `systray.Run` is blocking; on Linux/Windows it can run in a dedicated goroutine

### Architecture

```
NewPlatformProvider() Provider
  └─ attempts to build SystrayProvider
  └─ falls back to HeadlessProvider if DISPLAY/WAYLAND_DISPLAY/service env absent (Linux)
  └─ returns HeadlessProvider on Darwin (until macOS spec is implemented)

SystrayProvider (//go:build linux || windows)
  Setup()       → starts systray.Run() in a dedicated goroutine, signals ready
  SetState()    → updates systray.SetTooltip(state)
  NotifyError() → Linux: exec notify-send (best-effort); Windows: momentary tooltip update
  OnPause/OnResume/OnQuit/OnConfigure → stored; called from menu item click goroutine
  Quit()        → calls systray.Quit()
```

### Menu Layout

```
┌─────────────────────────┐
│ [icon] orchestratr      │  ← tooltip shows state: "running" / "paused"
├─────────────────────────┤
│  ● running              │  ← status item (disabled, display only)
├─────────────────────────┤
│  Pause                  │
│  Resume                 │
├─────────────────────────┤
│  Configure…             │
├─────────────────────────┤
│  Quit                   │
└─────────────────────────┘
```

### Build-Tag Strategy

| File | Build tag | Purpose |
|------|-----------|---------|
| `internal/tray/systray.go` | `//go:build (linux \|\| windows) && !notray` | Real SystrayProvider |
| `internal/tray/provider.go` | (none) | `NewPlatformProvider()` — build-tag awareness |
| `internal/tray/provider_stub.go` | `//go:build !(linux \|\| windows) \|\| notray` | `NewPlatformProvider()` stub returning HeadlessProvider |

The `-tags notray` build tag lets CI builds skip the CGo dependency without code changes.

## Acceptance Criteria

- [x] On Linux with a desktop environment (DISPLAY or WAYLAND_DISPLAY set), `orchestratr start` displays a tray icon in the system notification area
- [x] On Windows, `orchestratr start` displays a tray icon in the taskbar notification area
- [x] The tray tooltip reflects the daemon state: "orchestratr: running" or "orchestratr: paused"
- [x] The tray context menu has Pause, Resume, Configure, and Quit items; clicking each fires the registered callback
- [x] Clicking Quit in the tray menu stops the daemon (equivalent to `orchestratr stop`)
- [x] `NotifyError` produces a visible desktop notification on Linux (via `notify-send`) and on Windows (via tooltip flash)
- [x] On Linux without DISPLAY or WAYLAND_DISPLAY (headless/CI), `NewPlatformProvider()` returns `HeadlessProvider` — no crash or missing binary
- [x] `go build ./...` succeeds without CGo when `-tags notray` is passed (Linux CI use case)
- [x] On Darwin, `NewPlatformProvider()` returns `HeadlessProvider` (stub, pending macOS platform layer)

## Affected Areas

| Area | Files |
|------|-------|
| **Create** | `internal/tray/systray.go` — `SystrayProvider` implementation (`linux \|\| windows`) |
| **Create** | `internal/tray/systray_test.go` — provider interface + callback tests (run without real display) |
| **Create** | `internal/tray/provider.go` — `NewPlatformProvider()` |
| **Create** | `internal/tray/provider_stub.go` — `NewPlatformProvider()` stub for darwin/notray |
| **Modify** | `cmd/orchestratr/main.go` — replace `&tray.HeadlessProvider{}` with `tray.NewPlatformProvider()` |
| **Modify** | `go.mod` / `go.sum` — add `fyne.io/systray` |

## Constraints

- `systray.Run()` must be called from the same goroutine as the daemon setup is NOT required on Linux/Windows — a dedicated goroutine is sufficient
- The tray must not block `runStart()` from returning promptly; `SystrayProvider.Setup()` returns after signalling the tray goroutine is running
- If tray initialization fails (e.g. missing display), fall back to `HeadlessProvider` silently with a log warning — never crash
- Linux requires `libappindicator3-dev` (or `libayatana-appindicator3-dev`) at compile time; document in install.sh
- No icons bundled initially — use the default system icon (empty icon bytes is acceptable for first pass)
- `NotifyError` is best-effort: if `notify-send` is not installed or the call fails, log to daemon log and continue

## Out of Scope

- macOS tray (requires Cocoa/main-thread restructuring — separate spec with macOS platform layer)
- Custom tray icons (embedded PNG/ICO) — a follow-up polish item
- Tray animation or badge counts
- Per-app launch/ready notifications (NotifyError is for errors only)

## Dependencies

- `fyne.io/systray` v1.11.0+
- Linux compile-time: `libappindicator3-dev` or `libayatana-appindicator3-dev`
- Linux runtime: a compositor/DE that supports AppIndicator (GNOME, KDE, XFCE, etc.)

## Notes

### Why fyne.io/systray over getlantern/systray

`fyne.io/systray` is the maintained fork, used by the Fyne toolkit. It receives regular updates and bugs are addressed. `getlantern/systray` is effectively unmaintained as of 2024.

### systray.Run goroutine pattern for Linux/Windows

```go
type SystrayProvider struct {
    ready chan struct{}
    // ...
}

func (p *SystrayProvider) Setup() error {
    p.ready = make(chan struct{})
    go systray.Run(p.onReady, p.onExit)
    select {
    case <-p.ready:
        return nil
    case <-time.After(3 * time.Second):
        return fmt.Errorf("tray did not become ready within 3s")
    }
}

func (p *SystrayProvider) onReady() {
    systray.SetTooltip("orchestratr: starting")
    // ... add menu items ...
    close(p.ready)
}
```

### notify-send pattern for Linux

```go
func (p *SystrayProvider) NotifyError(title, message string) {
    if _, err := exec.LookPath("notify-send"); err != nil {
        return // best-effort
    }
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()
    _ = exec.CommandContext(ctx, "notify-send", "--urgency=critical", title, message).Run()
}
```

### Windows notification pattern

Windows does not have a standard CLI notification tool equivalent to `notify-send`. `fyne.io/systray` does not directly expose balloon notifications. For the first pass, `NotifyError` on Windows updates the tray tooltip to show the error for 5 seconds, then restores the previous state. A proper balloon notification via `Shell_NotifyIcon` with `NIF_INFO` can be added as a polish task.
