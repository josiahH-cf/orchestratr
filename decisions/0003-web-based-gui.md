# 0003: Web-based management GUI

**Date:** 2026-03-01
**Status:** Accepted
**Context:** Spec 06 — Management GUI

## Decision

Use a web-based GUI (embedded HTML/JS served by Go's `net/http` + `embed`) instead of Fyne or another native desktop toolkit.

## Rationale

- Fyne requires heavy C dependencies (OpenGL, GLFW, X11 headers) that complicate CI builds and don't work on headless/remote machines or WSL setups without an X server
- The management GUI is used **rarely** (initial setup, adding apps) — it is a config editor, not an always-on dashboard
- A web UI served on localhost requires zero external dependencies, works on any platform with a browser (including WSL where the browser opens on the Windows host), and is trivial to test
- Go's `embed` package allows bundling HTML/CSS/JS into the single binary — no external assets
- The daemon already runs a localhost HTTP API on port 9876; the GUI server uses a separate ephemeral port to avoid coupling

## Consequences

- "Open in browser" UX instead of a native window — acceptable for a config editor
- Requires `xdg-open` / `open` / `start` to launch the browser — falls back to printing the URL if unavailable
- No complex native key capture widget — chord entry is a text field with validation
- Lighter binary than Fyne (~30 MB savings)
