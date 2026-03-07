# Codex Agent Instructions

> **This file is scoped to Codex. For universal routing, conventions, and workflow phases, see ../AGENTS.md. AGENTS.md takes precedence over this file.**

## Codex-Specific Behavior

- Read `../AGENTS.md` at the start of every task for routing, conventions, and boundaries
- Use `../.specify/constitution.md` as the project identity reference
- Consult advisory routing hints in `../workflow/ROUTING.md` when choosing tasks (any agent can do any task)
- Follow branch naming conventions in `../workflow/ROUTING.md`: `agent/type-short-description` (where agent is any identifier)

## Execution Mode

- Codex operates in unattended mode with structured plans
- Plans are written to `/.codex/PLANS.md` before execution
- Each plan milestone maps to an acceptance criterion from the spec
- Verify each milestone before advancing to the next

## Workflow Commands

Codex participates in the same 8-phase workflow defined in AGENTS.md. Entry points:

- Phase 4 (Scaffold): read feature specs, produce architecture plan
- Phase 6 (Code): execute implementation from fine-tuned spec
- Phase 7 (Test): run test suites, report results
- Bug Track: process bug specs from the backlog

## References

- Universal routing: `../AGENTS.md`
- Constitution: `../.specify/constitution.md`
- Implementation plan template: `PLANS.md` (this directory)
- Sandbox config: `config.toml` (this directory)
