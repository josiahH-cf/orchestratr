<!-- role: canonical-source -->
<!-- slash-command: build-session -->
<!-- description: Autonomous build session for test-first development of one feature -->
# Build Session

> This build session maps to Phases 6–7 (Code + Test) of the workflow. The `/continue` command can auto-invoke this flow. For new features, ensure Phase 5 (Fine-tune) is complete first — task files should have model assignments, branch names, and EARS/GWT acceptance criteria. See `AGENTS.md → Workflow Phases`.

**Standing rules for all sessions:**

- Follow `AGENTS.md` and workflow contracts in `/workflow/PLAYBOOK.md`, `/workflow/FILE_CONTRACTS.md`, and `/workflow/FAILURE_ROUTING.md`.
- Reference `.specify/constitution.md` for alignment before every implementation decision.
- Log bugs via `/bug` when discovered — do not silently work around issues.
- Every artifact produced (spec, task file, test, implementation, review report) is committed or written to its canonical location before moving on.
- Fresh context means: no prior conversation carried forward. When indicated, end the current session and begin a new one.

---

## Build

**Covers:** Test phase, Implement phase (looped per task)

**Purpose:** Take a single planned issue and build it end-to-end: write failing tests for all acceptance criteria, then implement one task at a time until all tasks pass. Each task is a commit. This session is autonomous — the agent works through the task file without further input unless a decision or blocker arises.

**Session inputs:** The path to one task file (e.g., `/tasks/[feature-id]-[slug].md`).

---

```text
You are building a planned feature. The task file for this session is provided. Read it now, along with the linked spec.

Work through two phases in order. Commit at every checkpoint described below.

TEST PHASE
1. Read the spec's acceptance criteria.
2. Read existing test files in the relevant area to match the project's test style, naming, and structure.
3. For each acceptance criterion, write at least one test that:
   - Asserts the expected behavior described in the criterion.
   - Will fail because the feature does not exist yet.
   - Uses a descriptive name stating the expected behavior.
4. Do not write any implementation code — not even stubs or helpers that implement feature logic.
5. Run the full test suite. Confirm: new tests fail, all existing tests pass.
6. Commit the test files with a message referencing the spec.

IMPLEMENT PHASE (repeat for each task)
For each task in the task file, in order:
1. Orient before writing:
   - Identify the next task marked "Not started."
   - Read the test file(s) covering this task's criteria.
   - Read the source files this task will modify.
   - Confirm you understand what the tests expect before writing.
2. Implement only this one task.
   - Make the failing tests for this task pass.
   - Follow existing code patterns.
   - Do not modify any existing tests. If a test seems wrong, the implementation is wrong.
   - Do not add functionality beyond what this task specifies.
   - Do not change files outside this task's listed scope.
3. If you encounter a non-obvious decision, write it to /decisions/[NNNN]-[slug].md before proceeding.
4. Run the full test suite. If unrelated tests break, fix the regression without modifying those tests.
5. Commit with a message referencing the task.
6. Update the task file: mark this task's status as `[x] Complete`. Update the Status counts (Complete, Remaining). Append to the Session Log.
7. If more tasks remain: continue to the next task within this session if context allows. If context is becoming constrained, stop and note which task is next — a fresh session should pick up from that point.

When all tasks are complete:
- Confirm the full test suite passes.
- State: "All tasks complete for [feature-id]-[slug]. Label the issue status:implemented. This feature is ready for the Review phase."
```

**Output:** Committed failing tests, committed implementation (one commit per task), updated task file with all tasks marked complete.

**Context note:** If the feature has many tasks and context becomes constrained, it is correct to stop mid-feature and resume in a fresh session pointing at the same task file. The task file's status tracking ensures continuity.

**Next:** Hand off to the Review & Ship meta-prompt (`phase-7d-review-and-ship.md`).
