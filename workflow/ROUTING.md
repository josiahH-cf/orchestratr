# Agent Routing, Branches & Concurrency

> Referenced from `AGENTS.md`. This is part of the canonical workflow — see `/governance/REGISTRY.md`.

## Capability-Neutral Routing

**Any agent (Claude, Copilot, Codex) may execute any phase or task.** There are no enforced model-to-task bindings. The only hard constraint is **file-level exclusivity** — no two agents may edit the same file simultaneously.

### Advisory Routing Hints

The table below captures model strengths as non-binding suggestions. Use these as tiebreakers when choosing which tool to point at a task, not as enforcement.

| Task Type | Suggested Model | Why |
|-----------|----------------|-----|
| Complex architecture, refactoring | Claude | Deep reasoning, large context |
| UI/frontend, iterative design | Copilot | Fast iteration, inline suggestions |
| Batch operations, migrations | Codex | Unattended execution, structured plans |
| Bug reproduction & fix | Claude | Diagnostic reasoning |
| Automated review + merge | Review Bot (subagent) | Independent automated review; prefer different model than implementer |
| Test writing | Any | Depends on complexity |
| Documentation, boilerplate | Copilot | Fast generation |
| CI/CD, infrastructure | Codex | Deterministic execution |

> These hints are advisory. Any tool can perform any task type.

### Advisory Tier Reference

All advisory language in the workflow follows a three-tier model. Tier selection is context-sensitive — it adapts based on `STATE.json → advisoryProfile` and the current phase. See `workflow/ORCHESTRATOR.md → Context-Sensitive Advisory Guidance` for the full model.

| Tier | Prefix | Meaning |
|------|--------|---------|
| Inform | "Note:" / "FYI:" | Background awareness — no action expected |
| Suggest | "Consider" / "You may want to" | Actionable option — safe to skip |
| Recommend | "Recommended:" / "Strongly consider" | High-value action — skipping has known trade-offs |

All tiers are non-blocking. No advisory suggestion prevents phase advancement.

## Branch Naming

Format: `agent/type-short-description`

- **agent:** Any identifier for the agent session (e.g., `claude`, `copilot`, `codex`, `agent-1`, a user name, or any descriptive label)
- **type:** `feat`, `bug`, `refactor`, `chore`, `docs`
- **short-description:** 2–4 word kebab-case summary

Examples:

- `claude/feat-auth-flow`
- `agent-1/feat-dashboard-layout`
- `codex/bug-login-crash`
- `josiah/refactor-db-queries`

## Concurrency Rules

- Worktree pattern: `.trees/<agent>-<task>/`
- Max parallel agents: `3` (override in AGENTS.md if needed)
- Rule: never two agents editing the same file simultaneously
- Setup: `scripts/setup-worktree.sh <agent> <type> <description>`

## Concurrency-Aware Task Assignment

When assigning tasks for parallel execution:

1. Verify no file overlap between tasks (check `Files:` fields in task files)
2. If overlap exists, reassign to a single agent or redesign the split
3. Run `scripts/clash-check.sh` after worktree creation to detect existing conflicts
4. See `workflow/CONCURRENCY.md` for full concurrency safety reference

## Multi-Agent `/continue` Coordination

When `/continue` is invoked in multiple tools simultaneously, each agent follows the claim protocol:

1. Read `workflow/STATE.json` and the task backlog
2. Identify the next unclaimed, file-disjoint unit of work (pending feature, open bug, maintenance task)
3. Write a claim into `STATE.json → activeClaims` with agent identity, task file, and locked files
4. Execute the claimed work
5. Release the claim on completion (remove from `activeClaims`)

If no unclaimed work exists, the agent reports "nothing to claim" rather than duplicating.

See `workflow/ORCHESTRATOR.md` for the full claim-based loop contract.
