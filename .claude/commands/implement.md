<!-- generated-from-metaprompt -->
You are implementing one task from a planned feature. Only one.

Read the task file at: $ARGUMENTS
Read the project's conventions file (AGENTS.md).

Orient before writing:
1. Identify the next task in the task file that is marked "Not started."
2. Read the test file(s) that cover this task's acceptance criteria.
3. Read the source files this task will modify.
4. Confirm you understand what the tests expect before writing any code.

Implement ONLY this one task. Not the next one. Not a partial start on another.

Rules:
- Make the failing tests for this task pass.
- Follow existing code patterns. Read the surrounding code before writing.
- Do not modify any existing tests. If a test seems wrong, the implementation is wrong — not the test.
- Do not add functionality beyond what this task specifies. No bonus features, no preemptive refactors.
- Do not change files outside the scope listed in this task's "Files" field.
- If you encounter a non-obvious decision, write it to /decisions/[NNNN]-[slug].md before proceeding. Use the next available number.

After implementation:
1. Run the full test suite — not just this task's tests.
2. If unrelated tests break, fix the regression without modifying those tests.
3. Commit with a message referencing the task (example: "Implement [task name] for [feature-name] — task 2/4").
4. Update the task file: mark this task's status as [x] Complete. Update the Status counts (Complete, Remaining).
5. Append to the Session Log: date, what was completed, any blockers or decisions made.

After committing, check the task file:

- If MORE tasks remain with status "Not started":
  State: "Task [N] complete. [M] tasks remaining. End this session. Start a fresh context window and run Phase 4 again with the same task file path."

- If ALL tasks are now complete:
  State: "All tasks complete. Label the issue status:implemented. Next phase: Review (Phase 5). Start in a fresh context window. For best results, use a different agent or model for review."
