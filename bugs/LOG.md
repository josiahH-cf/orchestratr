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

### BUG-002: Windows installer path (`irm ... | iex`) fails with parser/empty-path behavior
- **Location:** `install.ps1` execution mode and script assumptions when invoked via pipeline
- **Phase:** manual troubleshooting (post-install verification)
- **Severity:** blocking
- **Expected:** `irm https://raw.githubusercontent.com/josiahH-cf/orchestratr/main/install.ps1 | iex` installs orchestratr on Windows reliably.
- **Actual:** Installer invocation fails (empty-path/parse behavior), and installation does not complete.
- **Fix-as-you-go:** no
- **Status:** open
- **Logged:** 2026-03-13

### BUG-003: Windows `go install ...@latest` fails due case-colliding repo files
- **Location:** repository tree includes `.github/PULL_REQUEST_TEMPLATE.md` and `.github/pull_request_template.md`
- **Phase:** manual troubleshooting (Windows fallback install)
- **Severity:** blocking
- **Expected:** `go install github.com/josiahH-cf/orchestratr/cmd/orchestratr@latest` succeeds on Windows.
- **Actual:** Go module zip creation fails on Windows with case-insensitive filename collision, preventing binary creation.
- **Fix-as-you-go:** no
- **Status:** open
- **Logged:** 2026-03-13

### BUG-004: Windows binary reports WSL2 and refuses `start` when launched via WSL-invoked PowerShell
- **Location:** platform detection path used by `orchestratr start` (runtime environment detection)
- **Phase:** manual troubleshooting (Windows install/start validation)
- **Severity:** blocking
- **Expected:** Windows-built `orchestratr.exe` launched from PowerShell should run as Windows host runtime and allow `start`.
- **Actual:** CLI prints WSL2 warning and refuses start (`refusing to start inside WSL2`) even when executed from Windows PowerShell in this setup.
- **Fix-as-you-go:** no
- **Status:** fixed
- **Logged:** 2026-03-13
- **Resolved:** 2026-03-13

### BUG-005: Scripted Windows start path exits quickly and leaves stale PID
- **Location:** daemon lifecycle when launched via WSL-invoked PowerShell automation (`start` + immediate follow-up status)
- **Phase:** manual troubleshooting (Windows install/start validation)
- **Severity:** non-blocking
- **Expected:** `orchestratr start` keeps daemon running and `orchestratr status` reports running state.
- **Actual:** Start reports daemon PID, then status reports stale PID. Logs show daemon receives terminated signal shortly after scripted launch in this environment.
- **Fix-as-you-go:** no
- **Status:** fixed
- **Logged:** 2026-03-13
- **Resolved:** 2026-03-13
