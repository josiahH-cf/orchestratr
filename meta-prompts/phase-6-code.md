<!-- role: canonical-source -->
<!-- phase: 6 -->
<!-- description: TDD implementation — make failing tests pass, one task at a time -->
# Phase 6 — Code

**Objective:** Implement features following TDD from fine-tuned specs and task files. Tests exist first — make them pass.

`/implement` is **direct single-feature execution**. It does not manage state transitions, select the next action, or advance phases — that is the orchestrator's job (`/continue`). Use `/implement` when you know which feature to build. Use `/continue` when you want the orchestrator to determine the next action (which may invoke `/implement` at Phase 6).

**Trigger:** Phase 5 complete (`/tasks/[feature-id]-[slug].md` exists with ordered tasks, model assignment, and branches).

**Entry commands:**
- Claude: `/implement`
- Copilot: `phase-6-implement.prompt.md`

---

## What Happens

1. Confirm `/test pre` has created failing tests for this feature
2. Orient: read task file, failing tests, source files
3. Verify constitution alignment before implementing
4. Implement one task at a time — make failing tests pass
5. Follow existing code patterns
6. Log bugs via `/bug` — don't silently work around them
7. Update spec before any unplanned decision
8. Commit per task on the assigned branch

## Gate

- Pre-implementation tests exist for every AC (from `/test pre`)
- All tasks in current feature's task file marked Complete
- Full test suite passes
- No unresolved blocking bugs

## Output

- Passing code on feature branch
- Updated task file with completion status
- Decision records for any unplanned decisions
- Bug log entries for any discovered bugs

## Rules

- TDD — tests exist before implementation
- One task per commit
- Do not modify existing tests
- Do not add functionality beyond spec
- Update spec before unplanned decisions

## See Also

- Bug logging: `/bug` command
