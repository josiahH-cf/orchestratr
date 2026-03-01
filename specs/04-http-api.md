# Feature: HTTP API & IPC Protocol

**Status:** Not started  
**Project:** orchestratr

## Description

orchestratr exposes a localhost-only HTTP API that serves two purposes: (1) the daemon's own management interface (health, status, reload), and (2) a language-agnostic connection point for registered apps. Any app in any language that can make HTTP requests can communicate with orchestratr. This is the foundation for future app-to-app messaging, but v1 focuses on registration, heartbeat, and launch notification endpoints.

## Acceptance Criteria

- [ ] HTTP server binds to `127.0.0.1` on a configurable port (default: `9876`) — not accessible from the network
- [ ] `GET /health` returns `{"status": "ok", "version": "<version>"}` (unauthenticated)
- [ ] `GET /apps` returns the current app registry as JSON
- [ ] `POST /apps/{name}/launched` allows an app to notify orchestratr that it has started (for process tracking)
- [ ] `POST /apps/{name}/ready` allows an app to signal it is fully initialized and ready
- [ ] `POST /reload` triggers config hot-reload and returns the new registry or validation errors
- [ ] All endpoints return JSON with consistent error format: `{"error": "<message>", "code": "<error_code>"}`
- [ ] The API port is written to a well-known discovery file (`~/.config/orchestratr/port`) so apps can find it without hardcoding
- [ ] Requests from non-localhost origins are rejected with 403

## Affected Areas

| Area | Files |
|------|-------|
| **Create** | `orchestratr/api/server.py` — HTTP server setup, middleware, routing |
| **Create** | `orchestratr/api/routes.py` — endpoint handlers |
| **Create** | `orchestratr/api/models.py` — request/response JSON schemas |

## Constraints

- **Localhost only** — no TLS, no auth tokens (the OS process boundary is the security perimeter)
- JSON request/response bodies exclusively (no HTML, no form data)
- Response times < 50ms for all endpoints (local only, no I/O-bound operations)
- The server must not block the daemon's hotkey event loop
- No external HTTP framework dependency heavier than the language's standard library (e.g., Go's `net/http`, Python's `http.server` + routing)

## Out of Scope

- App-to-app message routing (v2 — apps communicating through orchestratr as a message bus)
- WebSocket / streaming connections
- Authentication or API keys (localhost-only is sufficient for single-user desktop apps)
- Remote access or tunneling
- Rate limiting

## Dependencies

- `01-core-daemon.md` — API server runs as part of the daemon process
- `03-app-registry.md` — `/apps` endpoint reads the registry

## Notes

### Service discovery

Apps need to find orchestratr's port. The recommended pattern:

1. orchestratr writes its port to `~/.config/orchestratr/port` (a plain text file containing just the number)
2. Apps read this file at startup to discover the API
3. If the file doesn't exist or the port is unreachable, the app knows orchestratr isn't running

This is simpler and more portable than mDNS, Unix sockets (not cross-platform), or named pipes.

### App connection lifecycle

```
App starts → reads port file → POST /apps/{name}/launched
App initializes → POST /apps/{name}/ready
App runs normally (no polling required)
App shuts down → (optional) POST /apps/{name}/stopped, or orchestratr detects via process tracking
```

Apps are not required to connect to orchestratr — the API is opt-in. An app that doesn't call any endpoints still works fine (orchestratr launches it via its command, fire-and-forget). The API enables richer integration for apps that choose to participate.

### Future v2 extension points

The API is designed to be extended without breaking v1 clients:

- `POST /events` — publish an event (e.g., `{"source": "espansr", "type": "sync_complete", "data": {...}}`)
- `GET /events/stream` — SSE stream for subscribing to events
- `POST /apps/{name}/message` — send a message to a specific app

These are **not in scope for v1** but the URL structure and JSON conventions should be designed with them in mind.

### Client SDK (optional, not required)

For convenience, a thin client library could be published for common languages:

```python
# Python example (future)
from orchestratr import client
client.notify_launched("espansr")
client.notify_ready("espansr")
```

But since the API is just HTTP + JSON, any language can use it directly with no SDK. The SDK is a convenience, not a requirement.
