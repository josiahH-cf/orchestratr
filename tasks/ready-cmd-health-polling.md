# Tasks: `ready_cmd` Health Polling

**Spec:** `/specs/ready-cmd-health-polling.md`
**Branch:** `feat/ready-cmd-health-polling`

## Task 1: Readiness Poller (internal/launcher/readiness.go)

- Create `PollReadiness(ctx, entry, tracker, executor, logger)` function
- If `ready_cmd` is empty, call `tracker.SetReady(entry.Name)` immediately and return
- If `ready_cmd` is set, poll at 500ms intervals using `bash -c <ready_cmd>`
- On exit code 0, transition to `ready` and stop polling
- On timeout (`ready_timeout_ms`, default 5000ms), log warning and stop (leave in `launched`)
- If process exits before ready (checked via `executor.IsRunning(name)`), stop polling
- Respect context cancellation (daemon shutdown)
- Each poll attempt has its own 5s execution timeout
- Write failing tests first

## Task 2: Integration with Launch Flow (cmd/orchestratr/main.go)

- After `SetLaunched(name)`, spawn `go PollReadiness(ctx, entry, tracker, executor, logger)`
- Wire in both `launchApp()` (hotkey path) and the launch API endpoint
- Pass daemon context so polling cancels on shutdown
- Write failing tests first

## Progress

- [x] Task 1
- [x] Task 2
