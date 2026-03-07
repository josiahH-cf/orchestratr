# Feature Lifecycle

This file is the lifecycle index. For execution and validation rules, use:

- `/workflow/PLAYBOOK.md`
- `/workflow/FILE_CONTRACTS.md`
- `/workflow/FAILURE_ROUTING.md`

## Project-Level Phases

The project lifecycle follows 9 phases plus a parallel Bug Track. These are one-time or periodic phases that establish the project foundation.

| Phase | Input | Action | Output |
| ----- | ----- | ------ | ------ |
| 1. Scaffold Import | Empty/new repo | Run `initialization.md` | Scaffold files placed |
| 2. Compass | Scaffold in place | Dynamic discovery interview (`/compass`) | `.specify/constitution.md`, AGENTS.md Overview |
| 3. Define Features | Constitution | Feature interview (`/define-features`) | Feature specs with Compass mapping |
| 4. Scaffold Project | Feature specs | Architecture reasoning (`/scaffold`) | `workflow/COMMANDS.md` Code Conventions + Core Commands |
| 5. Fine-tune Plan | Architecture plan | Spec + task finalization (`/fine-tune`) | Updated specs + `/tasks/[feature-id]-[slug].md` with task/model/branch mappings |
| 6. Code | Fine-tuned specs + task files + pre-tests | Direct TDD implementation (`/implement`) | Passing code on feature branches |
| 7. Test | Feature branch state | `pre` mode: failing tests, `post` mode: AC verification (`/test`) | Test results, bug log entries |
| 7a. Review Bot | All ACs pass | Automated full-rubric review + auto commit/push/merge (`/review-bot`) | Auto-merged PR → next feature or Phase 8; or findings file → back to Phase 6 |
| 7b. Review & Ship | Manual review requested | Feature review + PR creation (`/review-session`), optional cross-review (`/cross-review`) | Approved PR merged → next feature or Phase 8 |
| 8. Maintain | Shipped features | Ongoing maintenance (`/maintain`) — select level: Light / Standard / Deep | Maintenance mode active (level recorded in STATE.json) |
| 9. Operationalize | Maintenance level selected | Interview-driven automation config (`/operationalize`) — schedules, notifications, release publishing | `.github/maintenance-config.yml` + generated GitHub Actions workflows |
| Bug Track | Any phase | Bug logging (`/bug`) + fixing (`/bugfix`) | Bug log entries, fix PRs |

`/continue` is the **orchestrator** that auto-advances through phases 2–9 based on exit criteria and persisted state in `/workflow/STATE.json`. It selects the next right action — including routing to `/bugfix` for blocking bugs before continuing task work. At Phase 6, `/continue` delegates to `/implement`. Use `/implement` directly when you know which feature to build; use `/continue` when you want the orchestrator to decide.

## Feature-Level Phases (Per-Feature Lifecycle)

> **Note on phase numbering:** Project-Level Phases (above) run once per project setup. Feature-Level Phases (below) run once per feature, nested within Project Phases 6–7. When `continue.md` refers to "Phase 6", it means the Project-Level Code phase. Inside Phase 6, the Feature-Level cycle (Pre-test → Implement → Post-test → Review → Ship) applies to each feature.


Every phase produces a named artifact; the next phase consumes it.

| Phase | Input | Action | Output | Who |
| ----- | ----- | ------ | ------ | --- |
| 1. Pre-test | Task file | Write failing tests for ACs | Committed failing tests | Any agent |
| 2. Implement | Tests + Tasks | TDD implementation (per task) | Passing code on feature branch | Any agent |
| 3. Post-test | Implementation | Verify all ACs pass | Test results, bug log | Any agent |
| 4. Bot Review | Post-test pass + diff + spec | Automated full-rubric review | AUTO-MERGE or findings file | Review bot (different model preferred) |
| 5. Review (manual) | Bot review fail or manual request | Human-triggered review + optional cross-review | PASS/FAIL report with criterion evidence | Different agent |
| 6. Ship | Review pass (bot or manual) | Create PR, merge (auto by bot, or manual) | Merged + branch deleted | Bot auto-merge or human approval |

## Label Conventions

- `status:planned`  -  tasks written, ready to build
- `status:tests-written`  -  failing tests committed
- `status:implemented`  -  all tasks complete, tests pass
- `status:reviewed`  -  cross-agent review passed
- `status:done`  -  merged

- `size:S` / `size:M` / `size:L`  -  effort estimate
- `agent:claude` / `agent:codex` / `agent:copilot`  -  assigned agent (optional)

## Bulk Issue Creation

When creating multiple features, write all issues first (labeled `status:idea`), then scope them one at a time. Each issue must be independently actionable; note explicit dependencies in the issue body.

## Feature Identity

- Feature ID format: `[issue-id]-[slug]` (example: `42-user-auth`)
- Spec and task files must share the same feature ID
- Criteria IDs in spec: `AC-1..N`
- Task IDs in tasks file: `T-1..N`
- Orchestration state file: `/workflow/STATE.json`
