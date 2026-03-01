<!-- generated-from-metaprompt -->
You are reviewing a completed feature branch. You did not write this code. Approach it with fresh eyes.

Read the spec at: $ARGUMENTS
Read the project's conventions file (AGENTS.md).
Read the full diff of the current branch against the target branch.
Read the task file linked from the spec's feature name at /tasks/[feature-name].md.

Evaluate every item below. For each, report PASS or FAIL with specific evidence.

**Acceptance criteria coverage:**
For each criterion in the spec:
- Does a test exist that specifically verifies this criterion?
- PASS: [criterion] — test at [file:line] verifies this.
- FAIL: [criterion] — no test found, or test does not verify the stated behavior.

**Code quality checks:**
1. No existing tests were modified to accommodate the new code.
   - PASS/FAIL + evidence.
2. No files outside the spec's "Affected Areas" were changed.
   - PASS/FAIL + list any out-of-scope files.
3. No hardcoded values, secrets, or environment-specific strings.
   - PASS/FAIL + locations if found.
4. No functions over 50 lines.
   - PASS/FAIL + function names and line counts if found.
5. No dead code or unused imports.
   - PASS/FAIL + locations if found.
6. Error handling is explicit, not silent (no bare catches, no swallowed errors).
   - PASS/FAIL + locations if found.
7. Naming is consistent with existing codebase conventions.
   - PASS/FAIL + inconsistencies if found.

**Summary:**
- Total criteria: [N]
- Passed: [N]
- Failed: [N]

If ALL checks pass:
  State: "Review PASSED. Branch is ready for cross-review. Next: Phase 5b (Cross-Review) in a fresh context window with a DIFFERENT agent or model. If cross-review is not available, proceed to Phase 6 (PR Create)."

If ANY check fails:
  List each failure with the specific fix needed.
  State: "Review FAILED. Return to Phase 4 (Implement) in a fresh context window to address the failures listed above. Do not proceed to PR."
