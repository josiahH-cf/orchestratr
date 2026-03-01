<!-- generated-from-metaprompt -->
You are planning implementation tasks. Do not write code. Do not write tests.

Read the spec at: $ARGUMENTS
Read the project's conventions file (AGENTS.md) for relevant patterns.

Decompose the spec into 2–5 implementation tasks, ordered by dependency (tasks that others depend on come first).

For each task, provide:
- **Name:** [short descriptive name]
- **Files:** [specific files this task will create or modify]
- **Done when:** [one sentence describing the verifiable end state]
- **Criteria covered:** [which acceptance criteria from the spec this task addresses]
- **Status:** [ ] Not started

Write the output to /tasks/[feature-name].md with this structure:

# Tasks: [feature-name]

**Spec:** /specs/[feature-name].md

## Status
- Total: [N]
- Complete: 0
- Remaining: [N]

## Task List

### Task 1: [name]
- **Files:** [list]
- **Done when:** [one sentence]
- **Criteria covered:** [which from spec]
- **Status:** [ ] Not started

[repeat for each task]

## Test Strategy
[Map each acceptance criterion to the task that will test it. Every criterion must appear exactly once.]

## Session Log
[empty — will be filled during implementation]

Rules:
- Maximum 5 tasks. If you need more, the spec should have been split during Phase 1. Do not proceed — recommend splitting and stop.
- If any single task touches more than 8 files, split that task into smaller tasks. The total must still not exceed 5. If it would, recommend splitting the spec instead.
- Every acceptance criterion must be covered by at least one task. No orphaned criteria.
- Tasks must be completable independently in sequence — no circular dependencies.
- Do not write code, tests, or modify any source files.

After writing the task file, state: "Plan complete. Label the issue status:planned. Next phase: Test (Phase 3)."

If this feature will use a long-running execution plan (milestone-based, multi-step), also state: "This feature is a candidate for an ExecPlan. Run Phase 2b before proceeding to Phase 3."
