# Workflow Playbook

This document defines how an agent executes work from spec to merged PR.

## Global Rules

> These rules are **Recommend** tier (see `workflow/ORCHESTRATOR.md → Advisory Tiers`). Agents should follow them by default but may adapt when context justifies deviation.

- Work from a single feature ID at a time.
- Keep branch scope aligned to the spec's Affected Areas.
- Recommended: Record non-obvious decisions in `/decisions/` before continuing.
- Move forward only when the current phase gate is satisfied (verify via PLAYBOOK gates below).
- Reference `.specify/constitution.md` for alignment on all design decisions.

## Project-Level Phase Gates

| Phase | Required Input | Required Output | Gate to Advance |
| ----- | -------------- | --------------- | --------------- |
| Compass | Scaffold files placed | `.specify/constitution.md` populated | All relevant themes addressed (no `[PROJECT-SPECIFIC]` placeholders in covered sections); ambiguities documented |
| Define Features | Constitution | Feature specs with Compass mapping | At least one feature spec exists in `/specs/` with Compass capability mapping; every constitution capability has at least one feature mapping; no orphan features (every feature traces to a capability) |
| Scaffold Project | Feature specs | `workflow/COMMANDS.md` Code Conventions + Core Commands | Neither section contains `[PROJECT-SPECIFIC]` |
| Fine-tune Plan | Architecture plan | `/tasks/[feature-id]-[slug].md` files + ordered AC/task/model/branch mappings | Every active spec has a matching task file; all ACs mapped to tasks |
| Code | Fine-tuned specs + task files + pre-tests | Passing code on feature branch | All tasks marked Complete, tests pass |
| Test | Implementation | Verified ACs, bug log reviewed | No blocking bugs, all ACs pass in `/test post` mode; `scripts/workflow-lint.sh` run (advisory — non-blocking; uses Suggest tier) |
| Review Bot | Post-test pass | Auto-merged PR or findings file | All rubric categories PASS, tests PASS, lint PASS → auto-merge; any FAIL → findings file written, route back to Code |
| Maintain | Shipped features | Maintenance mode active (level selected) | Maintenance level recorded in STATE.json; all items for selected level completed |
| Operationalize | Maintenance level selected + interview answers | `.github/maintenance-config.yml` + generated GitHub Actions workflows | Config file records all interview decisions; at least one workflow generated per enabled category; notification routing configured; all workflow YAML valid |

## Feature-Level Phase Contract

| Phase | Required Input | Required Output | Gate to Advance |
| ----- | -------------- | --------------- | --------------- |
| Scope | Issue or request | `/specs/[feature-id]-[slug].md` | 3–7 testable acceptance criteria with IDs |
| Plan | Spec | `/tasks/[feature-id]-[slug].md` | Every criterion mapped to one or more tasks |
| Test (`pre`) | Task file + spec | Failing tests committed | At least one failing test per criterion |
| Implement | Failing tests + task file | Passing code commits | Task statuses updated with evidence |
| Test (`post`) | Implemented feature + task file + spec | AC verification report + bug entries | All ACs verified or logged as bugs |
| Bot Review | Post-test pass + spec + task file + diff | Rubric review report | All 6 rubric categories PASS, tests PASS, lint PASS |
| Auto-Merge (bot) | Bot review PASS | Committed, pushed, merged PR | PR squash-merged, branch deleted, feature labeled `status:done` |
| Review (manual) | Bot review FAIL or manual override | PASS/FAIL review report | All criteria have passing test evidence |
| PR | Review PASS (bot or manual) | Open PR with required checklist | CI and policy checks green; lint report reviewed (advisory) |
| Merge | Approved PR | Merged branch + cleanup | Bot auto-merge or human merge approval |

## Definition of Done

A feature is done only when all are true:

- Task file status counts show zero remaining tasks.
- Full test suite passes.
- Review report is PASS with criterion-level evidence.
- PR template is complete with verification and rollback.

## Context Discipline

- Start each phase in a fresh session when context quality drops.
- Prefer file artifacts over chat memory for continuity.
- If compacting repeatedly, split the feature and continue in a new branch.
- Persist orchestration state transitions in `/workflow/STATE.json` when using `/continue`.
