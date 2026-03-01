<!-- generated-from-metaprompt -->
You are creating a pull request for a completed, reviewed feature branch.

Read the spec at: $SPEC_PATH
Read the full diff of the current branch against the target branch: $TARGET_BRANCH
Read the project's PR template at /.github/pull_request_template.md if it exists.

Produce the PR with this structure:

**Title:** [concise imperative summary of the change]

## What
[One sentence: what this PR does]

## Why
Spec: /specs/[feature-name].md

## Changes
[List the logical changes, grouped by area. Keep it scannable.]

## Testing
- [x] All acceptance criteria from the spec have corresponding tests
- [x] All tests pass locally
- [x] No existing tests were modified to accommodate new behavior
- [x] Linting and formatting checks pass

## Non-Code Checks
- [x] Spec acceptance criteria are all addressed
- [x] Diff is under 300 lines
- [x] No files changed outside the declared scope
- [x] Commit messages reference the spec or task file
- [x] No TODOs or placeholder text in the diff
- [x] No new dependencies added without justification
- [x] Rollback path is documented below

## Verification
[How a reviewer can verify beyond reading the diff — specific steps or commands]

## Rollback
[Special steps beyond git revert, or "Standard revert." if none]

Rules:
- If the diff exceeds 300 lines, state this clearly and recommend splitting before opening the PR.
- Check every box honestly. If any box cannot be checked, do not check it — leave it unchecked and explain why.
- The PR description must be understandable by a non-developer reviewer using the non-code checklist.
- Open the PR against the target branch using the project's standard method.

After creating the PR, state: "PR opened. Next phase: Human review and merge (Phase 7). A human reviewer should use the non-code checklist — no code reading required, approximately 10 minutes."
