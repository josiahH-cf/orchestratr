# Canonical Policy Registry

This registry defines the authority and ownership of scaffold policy files.

## Canonical Files

| File | Authority | Owner |
|---|---|---|
| `/AGENTS.md` | Entrypoint and precedence (10 sections including routing matrix) | Human maintainer |
| `/workflow/LIFECYCLE.md` | Lifecycle index (project-level + feature-level phases) | Human maintainer |
| `/workflow/PLAYBOOK.md` | Phase execution contract (project + feature gates) | Human maintainer |
| `/workflow/FILE_CONTRACTS.md` | Artifact schema and linkage | Human maintainer |
| `/workflow/STATE.json` | Machine-readable orchestration state for `/continue` | Orchestrator agent |
| `/workflow/FAILURE_ROUTING.md` | Failure matrix and escalation | Human maintainer |
| `/workflow/ROUTING.md` | Agent routing, branch naming, concurrency | Human maintainer |
| `/workflow/COMMANDS.md` | Build/test/lint commands, code conventions | Human maintainer |
| `/workflow/BOUNDARIES.md` | Behavioral boundaries, bug tracking format | Human maintainer |
| `/workflow/LINT_CONTRACT.md` | Lint check definitions and output contract | Human maintainer |
| `/workflow/CONCURRENCY.md` | Concurrency safety, drift detection, decomposition strategies | Human maintainer |
| `/workflow/ORCHESTRATOR.md` | Autonomous loop contract (session bootstrap, phase transitions) | Human maintainer |
| `/governance/CHANGE_PROTOCOL.md` | Policy mutation rules | Human maintainer |
| `/governance/POLICY_TESTS.md` | Validation requirements | Human maintainer |
| `/.specify/constitution.md` | Project identity (from Compass interview) | Compass / Human (via `/compass-edit`) |
| `/.specify/spec-template.md` | Feature spec template | Human maintainer |
| `/.specify/acceptance-criteria-template.md` | AC format reference (EARS + GWT) | Human maintainer |
| `/.github/REVIEW_RUBRIC.md` | Review scoring rubric (6 categories) | Human maintainer |
| `/.github/PULL_REQUEST_TEMPLATE.md` | PR template (GitHub default location, extended sections) | Human maintainer |
| `/bugs/LOG.md` | Bug tracking log | Any agent (via `/bug`) |

## Adapter Files (Non-Canonical)

- `/CLAUDE.md` — Claude adapter (imports AGENTS.md)
- `/.github/copilot-instructions.md` — Copilot adapter (links to AGENTS.md)
- `/.codex/config.toml` — Codex configuration
- `/.codex/AGENTS.md` — Codex adapter (references ../AGENTS.md)

- `/.codex/PLANS.md` — Codex execution plans
- `/.aiignore` — Files excluded from AI agent context

Adapters may add tool-specific mechanics but must not redefine canonical workflow policy.

## Scripts

- `/scripts/setup-worktree.sh` — Worktree creation, inventory, and cleanup
- `/scripts/clash-check.sh` — Pre-write conflict detection between worktrees
- `/scripts/policy-check.sh` — Policy validation script
- `/scripts/workflow-lint.sh` — Non-destructive workflow artifact linting

## Agent Definition Files

- `/.github/agents/implementer.agent.md` — Implementation specialist (TDD, one task at a time)
- `/.github/agents/reviewer.agent.md` — Review specialist (scores against rubric)
- `/.github/agents/planner.agent.md` — Planning specialist (architecture and task decomposition)

## Issue Templates

- `/.github/ISSUE_TEMPLATE/feature.yml` — Feature request issue template
- `/.github/ISSUE_TEMPLATE/bug.yml` — Bug report issue template
- `/.github/ISSUE_TEMPLATE/agent-task.yml` — Agent task issue template
- `/.github/ISSUE_TEMPLATE/feature.md` — Feature request (markdown fallback)
- `/.github/ISSUE_TEMPLATE/bug.md` — Bug report (markdown fallback)

## CI/CD Files

- `/.github/workflows/copilot-setup-steps.yml` — Environment setup for Copilot Coding Agent
- `/.github/workflows/copilot-agent.yml` — Issue-to-PR automation via @copilot assignment
- `/.github/workflows/claude-review.yml` — Mention-triggered PR review via @claude
- `/.github/workflows/autofix.yml` — CI-failure auto-fix loop
- `/.github/workflows/agentic-triage.yml` — Scheduled issue triage (read-only)


