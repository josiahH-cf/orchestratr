# Bug Log

### BUG-001: Stale runtime artifacts remain when daemon is not running
- **Location:** `internal/daemon/lock.go`, `internal/daemon/port.go`, `cmd/orchestratr/main.go`
- **Phase:** 9 - operationalize (investigation during `/continue` session)
- **Severity:** non-blocking
- **Expected:** Runtime artifacts should reflect process state; when daemon is not running, stale PID/port artifacts should be detected and cleaned or ignored consistently.
- **Actual:** `orchestratr status` reports `orchestratr is not running (stale PID 93843)` while `~/.config/orchestratr/orchestratr.pid` and `~/.config/orchestratr/port` still exist from prior runs.
- **Fix-as-you-go:** no
- **Status:** fixed
- **Logged:** 2026-03-08
- **Resolved:** 2026-03-08
