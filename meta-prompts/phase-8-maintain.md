<!-- role: canonical-source -->
<!-- phase: 8 -->
<!-- description: Ongoing maintenance — documentation, compliance, and standards enforcement -->
# Phase 8 — Maintain

**Objective:** Maintain project health through documentation, compliance, and standards enforcement. Maintenance is an **ongoing mode**, not a terminal "done" state.

**Trigger:** Feature shipped (post-Phase 7) or periodic maintenance trigger.

**Entry commands:**
- Claude: `/maintain`
- Copilot: `phase-8-maintain.prompt.md`

---

## Maintenance Level Selection

On entry, select a maintenance level and record it in `/workflow/STATE.json` (`maintenanceLevel` field). The level determines scope for this maintenance pass.

| Level | Scope | When to Use |
|-------|-------|-------------|
| **Light** | Docs + lint + stale TODO sweep | Low-touch periodic check-in |
| **Standard** | Light + compliance check + dependency audit + bug log review | Default recommended level |
| **Deep** | Standard + security audit + architecture review + performance regression check | Pre-release, scheduled audit, or after major changes |

If no level is specified, default to **Standard**.

## What Happens

### Initial Setup (first run after shipping)
1. Select maintenance level (prompt if not provided)
2. Generate README from constitution + feature specs
3. Generate CONTRIBUTING from AGENTS.md conventions
4. Produce release notes from implemented features
5. Run security baseline check

### Ongoing (periodic)
Execute items based on selected level:

**Light:**
1. Documentation drift — README/CONTRIBUTING vs current state
2. Auto-corrections — lint, format, stale TODOs

**Standard** (includes Light):
3. Compliance check — all specs have tests, all tests pass
4. Dependency audit — outdated or vulnerable
5. Bug log review — stale entries flagged or closed

**Deep** (includes Standard):
6. Security audit — vulnerability scan + policy review
7. Architecture review — drift from constitution principles
8. Performance regression check

## Feature Re-entry

Maintenance mode does not block new feature work. A project in maintenance can start new `/compass-edit` → `/define-features` → `/implement` cycles at any time. Feature-level phases run within ongoing maintenance — the project does not need to "exit" maintenance first. After the feature ships, return to the current maintenance level.

## Gate

- Maintenance level selected and recorded in `STATE.json`
- README exists and reflects current project state
- CONTRIBUTING exists with branch naming, commit format, PR requirements
- All items for the selected level completed
- No stale compliance issues (Standard and Deep)
- Security baseline checked (Deep)

## Output

- Updated documentation
- Compliance report (scope depends on level)
- Bug log entries for any findings
- `STATE.json` updated: `projectPhase: "8-maintain"`, `maintenanceLevel` set

## Rules

- Do not change application logic — maintenance only
- Commit maintenance changes separately from feature work
- Maintenance is ongoing — never set project status to a terminal "completed" state

## See Also

- Bug tracking parallel workflow: `AGENTS.md → Bug Track`
