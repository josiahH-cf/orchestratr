# Policy Tests

Map policy requirements to validation signals.

## Required Checks

| Requirement | Signal | Source |
|---|---|---|
| Spec has Feature ID and AC IDs | Pattern check in spec file | `/specs/*.md` |
| Task file maps tasks to AC IDs | Pattern check in task file | `/tasks/*.md` |
| Task status counts are coherent | Status/checklist consistency check | `/tasks/*.md` |
| Every spec has matching task file | Filename/feature-id parity check | `/specs/*.md`, `/tasks/*.md` |
| PR includes verification and rollback | PR template checklist complete | `/.github/PULL_REQUEST_TEMPLATE.md` |
| Constitution placeholders resolved after Compass | Fail if `[PROJECT-SPECIFIC]` remains in constitution for phase >= `3-define-features` | `/.specify/constitution.md`, `/workflow/STATE.json` |
| AGENTS placeholders resolved after Scaffold Project | Fail if `[PROJECT-SPECIFIC]` remains in AGENTS for phase >= `5-fine-tune-plan` | `/AGENTS.md`, `/workflow/STATE.json` |
| Build, lint, test commands are defined | No placeholder commands in AGENTS and no default CI command markers by phase >= `5-fine-tune-plan` | `/AGENTS.md`, `/.github/workflows/copilot-setup-steps.yml`, `/workflow/STATE.json` |
| Continue state is valid | `workflow/STATE.json` parses and points to existing task file | `/workflow/STATE.json` |
| Adapter files do not redefine canon | Adapter references AGENTS/workflow docs | `/CLAUDE.md`, `/.github/copilot-instructions.md` |
| AGENTS.md links resolve | Every markdown link in AGENTS.md points to an existing file | `/AGENTS.md` |
| Modular workflow files present | ROUTING.md, COMMANDS.md, BOUNDARIES.md, FILE_CONTRACTS.md exist | `/workflow/` |
| Workflow lint runs without error | `scripts/workflow-lint.sh` exits 0 and produces `workflow/LINT_REPORT.md` | `/workflow/LINT_REPORT.md` |
| Lint contract present | `workflow/LINT_CONTRACT.md` exists with four check categories | `/workflow/LINT_CONTRACT.md` |

## CI Mapping

- Run structural checks before language-specific checks.
- Run build/lint/test checks using AGENTS command values.
- Block merge on any failed policy test.

## Failure Semantics

- Policy test failure means process drift, not optional warning.
- Fix policy or artifact shape before continuing feature delivery.
