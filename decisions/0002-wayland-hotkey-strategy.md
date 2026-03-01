# Decision 0002: Wayland Global Hotkey Strategy

**Date:** 2025-07-13
**Status:** Accepted
**Feature:** /specs/02-hotkey-engine.md

## Context

Wayland's security model intentionally prevents applications from grabbing global hotkeys. Unlike X11 (`XGrabKey`), Windows (`RegisterHotKey`), and macOS (`CGEventTap`), there is no universal Wayland API for global keyboard shortcuts. Each compositor has its own mechanism, if any. The hotkey engine must define a strategy for Wayland that is pragmatic without requiring per-compositor plugin development (which is out of scope).

## Options

1. **Compositor-specific D-Bus protocols** — Support KDE (`org.kde.kglobalaccel`), GNOME (Shell extension or `org.freedesktop.portal.GlobalShortcuts`), and Sway (IPC). Good coverage but high implementation cost and maintenance burden per compositor.

2. **XWayland fallback** — Use the X11 `XGrabKey` path when `DISPLAY` is set, which works under XWayland compatibility. Simple to implement but doesn't work on pure Wayland sessions and is increasingly deprecated by distros.

3. **`org.freedesktop.portal.GlobalShortcuts` portal (freedesktop standard)** — A compositor-neutral D-Bus portal adopted by GNOME 44+, KDE 5.27+, and other portal-aware compositors. Single implementation covers the majority of Wayland desktops. Falls back to XWayland or CLI trigger on unsupported compositors.

4. **CLI trigger only** — No native Wayland grab. User configures their compositor to run `orchestratr trigger` as a custom keybinding. Zero platform code but poor user experience.

## Decision

**Option 3 (freedesktop GlobalShortcuts portal) as primary, with fallback chain.** The detection order is:

1. If `WAYLAND_DISPLAY` is set and `org.freedesktop.portal.GlobalShortcuts` is available → use the portal.
2. Else if `DISPLAY` is set (XWayland) → fall back to X11 `XGrabKey`.
3. Else → log a warning and document manual `orchestratr trigger` as the workaround.

The portal is the emerging standard and avoids per-compositor maintenance. KDE and GNOME both support it in recent versions. This gives good coverage today and improves over time.

## Consequences

- **Positive:** Single Wayland implementation covers GNOME 44+, KDE 5.27+, and future portal-aware compositors. Avoids maintaining compositor-specific code.
- **Positive:** XWayland fallback provides coverage for older sessions without extra work (reuses X11 code).
- **Negative:** The portal requires a D-Bus round-trip and user confirmation on first use (portal grant dialog). The user experience is slightly worse than X11's synchronous grab.
- **Negative:** Compositors that don't implement the portal (older GNOME, Sway, Hyprland) require manual keybinding configuration or XWayland.
- **Negative:** The portal API requires `libdbus` or a Go D-Bus client, adding a dependency.
