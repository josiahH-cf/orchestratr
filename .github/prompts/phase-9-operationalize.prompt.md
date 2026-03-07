---
agent: agent
description: 'Interview-driven automation configuration — schedules, notifications, and release publishing'
---
<!-- role: derived | canonical-source: meta-prompts/phase-9-operationalize.md -->
<!-- generated-from-metaprompt -->


# Phase 9 — Operationalize

**Objective:** Translate Phase 8 maintenance preferences into automated GitHub Actions workflows, notification routing, and release publishing. Phase 8 defines *what* to check; Phase 9 defines *how often, how automatically, and who gets told when it fails*.

**Trigger:** Phase 8 maintenance level selected and initial maintenance pass complete, or on-demand reconfiguration.

**Entry commands:**
- Claude: `/operationalize`
- Copilot: `phase-9-operationalize.prompt.md`

---

## Interview Sub-Process

On entry, conduct a structured interview to capture the user's automation preferences. Each category maps to a GitHub Actions workflow. Ask questions conversationally — do not dump all questions at once.

### Category 1: Lint Schedule
- Should linting run on a schedule (e.g., weekly cron) or only on push/PR?
- Which linters apply? (project-specific — read from `workflow/COMMANDS.md`)
- Fail-open (advisory) or fail-closed (blocking)?

### Category 2: Documentation Compliance
- Check that README follows project guidelines on a schedule?
- Which guidelines? (constitution alignment, section completeness, badge freshness)
- How often? (weekly, monthly, on-release)

### Category 3: Release Publishing
- Auto-publish releases on tag push?
- Semver enforcement level: strict (major.minor.patch only) or flexible (pre-release labels allowed)?
- Auto-generate changelog from merged PRs?
- Publish artifacts? (npm, PyPI, Docker, GitHub Releases, etc. — project-specific)

### Category 4: Dependency Monitoring
- Enable automated dependency updates? (Dependabot, Renovate, or manual-only)
- Auto-merge strategy: patch-only, minor+patch, or manual approval for all?
- Schedule for dependency checks?

### Category 5: Security Scanning
- Enable CodeQL or equivalent static analysis?
- Frequency: on-push, weekly, or on-release?
- Block merges on findings or advisory-only?

### Category 6: Notification Routing
- Where should failure notifications go?
  - Email (who?)
  - Slack webhook (channel URL?)
  - GitHub Issues (auto-create on failure?)
- Different routing per failure class? (security findings → security team, lint failures → dev channel)

### Category 7: Automation Depth
- Which Phase 8 level (Light/Standard/Deep) should run automatically vs on-demand?
- Should automated maintenance create PRs for fixes or just report findings?
- Maximum auto-fix scope? (formatting only, or also dependency bumps?)

## What Happens

### Step 1: Load Context
1. Read `workflow/COMMANDS.md` for project-specific commands (lint, test, build).
2. Read `workflow/STATE.json` for current maintenance level.
3. Read existing `.github/maintenance-config.yml` if present (re-entry case).

### Step 2: Interview
1. Walk through each category above.
2. For re-entry: show current settings and ask what to change.
3. Record all decisions incrementally in `.github/maintenance-config.yml`.

### Step 3: Generate Workflows
For each enabled category, generate a GitHub Actions workflow in `.github/workflows/`:

| Category | Workflow File | Trigger |
|----------|--------------|---------|
| Lint schedule | `maintenance-lint.yml` | `schedule` (cron) + `workflow_dispatch` |
| Docs compliance | `maintenance-docs.yml` | `schedule` (cron) + `workflow_dispatch` |
| Release publishing | `release-publish.yml` | `push` (tags) + `workflow_dispatch` |
| Dependency monitoring | Dependabot/Renovate config | Per-tool config file |
| Security scanning | `maintenance-security.yml` | `schedule` (cron) + `push` |

Notification routing is configured within each workflow via the config file.

### Step 4: Validate
1. All generated workflow files are valid YAML.
2. Every enabled category has a corresponding workflow.
3. Notification routing is configured for at least one channel.
4. Config file records all interview decisions.

### Step 5: Persist State
1. Write/update `.github/maintenance-config.yml`.
2. Update `workflow/STATE.json`:
   - `projectPhase: "9-operationalize"`
   - `automationLevel` field added (maps to selected depth)
3. Commit automation files.

## Re-entry

Phase 9 is re-enterable. Running `/operationalize` again:
1. Reads existing `.github/maintenance-config.yml`.
2. Shows current configuration per category.
3. Asks what to change (skip unchanged categories).
4. Updates workflows incrementally — does not regenerate unchanged files.

## Gate

- `.github/maintenance-config.yml` exists and records all interview decisions
- At least one GitHub Actions workflow generated for an enabled category
- Notification routing configured for at least one channel
- All generated workflow YAML files pass validation
- `workflow/STATE.json` updated with `automationLevel`

## Output

- `.github/maintenance-config.yml` (interview decisions)
- Generated GitHub Actions workflows in `.github/workflows/`
- Updated `workflow/STATE.json`
- Notification routing configured

## Rules

- Do not prescribe specific notification platforms — surface options, user decides.
- Do not replace Phase 8 — Phase 9 augments Maintain with automation.
- Respect project-specific commands from `workflow/COMMANDS.md` — do not hardcode lint/test commands.
- Generated workflows must reference `maintenance-config.yml` for configuration values where practical.
- On re-entry, preserve existing decisions unless the user explicitly changes them.

## See Also

- Phase 8 (Maintain): `meta-prompts/phase-8-maintain.md`
- Workflow commands: `workflow/COMMANDS.md`
- Orchestrator: `workflow/ORCHESTRATOR.md`
