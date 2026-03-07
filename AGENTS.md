# AGENTS

Canonical entrypoint for all coding agents. Read this first, then follow links to detailed references.

## Overview

Orchestratr is a cross-platform global hotkey launcher. It maps keyboard chords (Ctrl+Space + key) to app launches, runs as a background daemon with a web-based management GUI, and discovers apps via drop-in config files. Written in Go, distributed as a single binary with zero runtime dependencies. Targets Linux, macOS, and Windows.

## Workflow Phases

The project lifecycle follows 9 phases plus a parallel Bug Track.

### Phase 1 — Scaffold Import
- **Entry:** Run `initialization.md` meta-prompt
- **Gate:** Empty or new project repository → **Output:** Scaffold files placed → **Next:** Phase 2

### Phase 2 — Compass
- **Entry:** Claude: `/compass` · Copilot: `phase-2-compass.prompt.md` · Codex: `.codex/AGENTS.md`
- **Gate:** Scaffold present → **Output:** `.specify/constitution.md` populated (themes addressed, ambiguities documented) → **Next:** Phase 3

### Phase 3 — Define Features
- **Entry:** Claude: `/define-features` · Copilot: `phase-3-define-features.prompt.md`
- **Gate:** Constitution complete → **Output:** Feature specs in `/specs/` → **Next:** Phase 4

### Phase 4 — Scaffold Project
- **Entry:** Claude: `/scaffold` · Copilot: `phase-4-scaffold.prompt.md`
- **Gate:** Feature specs exist → **Output:** Architecture plan, `workflow/COMMANDS.md` finalized (no code) → **Next:** Phase 5

### Phase 5 — Fine-tune Plan
- **Entry:** Claude: `/fine-tune` · Copilot: `phase-5-fine-tune.prompt.md`
- **Gate:** Scaffold plan exists → **Output:** `/tasks/` files with AC, model, branch → **Next:** Phase 6

### Phase 6 — Code
- **Entry:** Claude: `/implement` · Copilot: `phase-6-implement.prompt.md`
- `/implement` is **direct single-feature execution** — use when you know which feature to build
- **Session mode:** `/build-session` — sustained multi-feature implementation session
- **Gate:** Task file + pre-impl tests exist → **Output:** Passing code on feature branch → **Next:** Phase 7

### Phase 7 — Test
- **Entry:** Claude: `/test` · Copilot: `phase-7-test.prompt.md`
- **Gate:** Implementation on feature branch → **Output:** Test results, bug log → **Next:** Phase 7a

### Phase 7a — Review Bot (Default Merge Path)
- **Entry:** Claude: `/review-bot` · Copilot: `phase-7a-review-bot.prompt.md`
- **Automatic:** `/continue` dispatches here after tests pass — no manual trigger needed
- **On-demand:** `/review-bot` to run manually at any time
- **Gate:** All ACs pass → **Output:** Auto-merged PR (on PASS) or findings file at `/reviews/[feature-id]-bot-findings.md` (on FAIL) → **Next:** Phase 8 or next feature (on PASS); back to Phase 6 (on FAIL)
- **Agent:** `.github/agents/review-bot.agent.md` — prefer a different model than the implementer (advisory)

### Phase 7b — Review & Ship (Manual Fallback)
- **Entry:** Claude: `/review-session` · Copilot: `phase-7d-review-session.prompt.md`
- **Optional:** `/cross-review` — second-opinion review from a different agent
- **Use when:** Manual human review is desired (security-critical, architectural changes)
- **Gate:** All ACs pass → **Output:** Approved PR merged → **Next:** Phase 8 or next feature

### Phase 8 — Maintain
- **Entry:** Claude: `/maintain` · Copilot: `phase-8-maintain.prompt.md`
- **Gate:** Feature shipped → **Output:** Updated docs, compliance report → **Next:** Phase 9 or next cycle

### Phase 9 — Operationalize
- **Entry:** Claude: `/operationalize` · Copilot: `phase-9-operationalize.prompt.md`
- **Gate:** Maintenance level selected → **Output:** `.github/maintenance-config.yml` + generated GitHub Actions workflows → **Next:** Ongoing (re-enterable)
- **Interview:** Covers lint schedule, docs compliance, release publishing, dependency monitoring, security scanning, notification routing, automation depth
- **Re-entry:** Run `/operationalize` again to update existing config — no duplication

### Bug Track (Parallel)
- **Entry:** Claude: `/bug` · Copilot: `phase-7b-bug.prompt.md` — invoke from any phase
- **Fix flow:** `/bugfix` — reproduce → diagnose → fix → verify → PR

### Orchestrator
- **Entry:** Claude: `/continue` · Copilot: `phase-10-continue.prompt.md`
- `/continue` is the **orchestrator**, not a direct implementation command. It reads `workflow/STATE.json`, determines the next action (including bug-routing), dispatches to the appropriate phase command, and auto-advances through phases 2–9. At Phase 6 it delegates to `/implement`.
- See `workflow/ORCHESTRATOR.md` for the loop contract

## Quick Reference

| Section | Reference |
|---------|-----------|
| Advisory routing hints, branches, concurrency | `workflow/ROUTING.md` |
| Advisory tier model and context-sensitive guidance | `workflow/ORCHESTRATOR.md → Context-Sensitive Advisory Guidance` |
| Concurrency safety, drift detection | `workflow/CONCURRENCY.md` |
| Build/test/lint commands, code conventions | `workflow/COMMANDS.md` |
| Boundaries (best practices, review points, avoid patterns), bug tracking | `workflow/BOUNDARIES.md` |
| Lifecycle phases (detailed) | `workflow/LIFECYCLE.md` |
| Phase execution gates | `workflow/PLAYBOOK.md` |
| Artifact ownership & contracts | `workflow/FILE_CONTRACTS.md` |
| Failure routing & escalation | `workflow/FAILURE_ROUTING.md` |
| Autonomous loop contract | `workflow/ORCHESTRATOR.md` |
| Policy changes | `governance/CHANGE_PROTOCOL.md` |
| Policy validation | `governance/POLICY_TESTS.md` |
| File registry | `governance/REGISTRY.md` |
| Orchestrator state | `workflow/STATE.json` |
