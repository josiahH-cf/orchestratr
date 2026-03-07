---
agent: agent
description: 'Create ordered specs with ACs, model assignments, and branches'
---
<!-- role: derived | canonical-source: meta-prompts/phase-5-fine-tune-plan.md -->
<!-- generated-from-metaprompt -->

[AGENTS.md](../../AGENTS.md)

# Phase 5 — Fine-tune Plan

**Objective:** Finalize the build plan: order task breakdowns, write acceptance criteria, assign models, and create branch names.

**Trigger:** Phase 4 complete (architecture plan exists, AGENTS.md populated).

**Entry commands:**
- Claude: `/fine-tune`
- Copilot: `phase-5-fine-tune.prompt.md`

---

## Execution Sequence

Follow these steps in order. Produce artifacts as you go — do not revisit earlier steps.

### Step A — Read Inputs

1. Read `workflow/STATE.json`, `AGENTS.md`, and all feature specs in `/specs/`.
2. Read the task template in `template/tasks/_TEMPLATE.md` (or existing task files).
3. Note the advisory routing hints and Branch Naming conventions.

### Step B — Create Task Files

For each feature spec, create or update `/tasks/[feature-id]-[slug].md`:

1. Produce an ordered list of tasks sized for one commit.
2. Write acceptance criteria using EARS + GWT formats:
   - **EARS:** `When [trigger], the system shall [response]`
   - **GWT:** `Given [context], When [action], Then [outcome]`
3. Each criterion is a machine-parseable checkbox: `- [ ] criterion`
4. Fill the Routing Plan section (see below) with suggested model assignments and rationale.
5. Name branches using `AGENTS.md → Branch Naming`: `agent/type-short-description`
6. Note which model will review each task (different from implementer).
7. Populate the Session Log with an initial entry recording the plan creation.

### Step C — Routing Plan

The Routing Plan section in each task file is a **suggestion, not a mandate**. It helps the developer and `/continue` make informed model choices, but assignments may change at execution time.

For each task, record:
- **Suggested model** and short rationale (one line)
- **Parallel suitability** — can this task run concurrently with others? (yes/no + reason)
- **Context needs** — small/medium/large context window expected

These suggestions are evaluated when work begins. The developer or `/continue` may override based on current availability, model capability changes, or task evolution.

### Step D — Summary Table

Present a single summary table for developer approval:

| Task | Description | ACs Covered | Suggested Model | Reviewer | Branch | Parallel? |
|------|-------------|-------------|-----------------|----------|--------|-----------|

This table is the final deliverable of Phase 5. Await approval before advancing.

## Gate

- Every active feature spec has a matching task file in `/tasks/`
- Task breakdowns exist with ordered tasks
- All ACs use EARS/GWT format and are machine-parseable checkboxes
- Routing Plan section populated with suggested model assignments and rationale
- Branch names follow convention
- Second-model review assignments documented
- Session Log has an initial entry for plan creation

## Output

- Updated `/tasks/[feature-id]-[slug].md` files with task status scaffold, AC mappings, routing plan, and branch assignment
- Summary table for developer approval (produced in Step D)
- Session Log initialized with plan-creation entry for `/continue` resumption

## See Also

- AC template: `.specify/acceptance-criteria-template.md`
- Advisory routing hints: `workflow/ROUTING.md → Advisory Routing Hints`
- Branch naming: `AGENTS.md → Branch Naming`
