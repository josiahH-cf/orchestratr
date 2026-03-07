---
agent: agent
description: 'Batch review and PR creation for completed features'
---
<!-- role: derived | canonical-source: meta-prompts/phase-7d-review-and-ship.md -->
<!-- generated-from-metaprompt -->

[workflow/PLAYBOOK.md](../../workflow/PLAYBOOK.md)
[workflow/FILE_CONTRACTS.md](../../workflow/FILE_CONTRACTS.md)

You are reviewing completed feature branches and preparing them for merge. You will process one feature at a time through review, and if it passes, create its pull request.

For each feature, work through the following:

REVIEW PHASE
Read the spec at /specs/[feature-id]-[slug].md, the task file at /tasks/[feature-id]-[slug].md, and the full diff of the feature branch against the target branch.

Evaluate every item below. For each, report PASS or FAIL with specific evidence.

Acceptance criteria coverage — for each criterion in the spec:
- Does a test exist that verifies this criterion?
- Does the implementation satisfy the criterion?
- Report: PASS with test location, or FAIL with what is missing.

Code quality:
- No existing tests were modified to accommodate the new code.
- No files outside the spec's Affected Areas were changed.
- No hardcoded values, secrets, or environment-specific strings.
- No functions over 50 lines.
- No dead code or unused imports.
- Error handling is explicit, not silent.
- Naming is consistent with existing codebase conventions.

Produce a summary:
- Total criteria: [N], Passed: [N], Failed: [N]
- Code quality issues: [list or "None"]

If ALL checks pass:
- State: "Review PASSED for [feature-id]-[slug]."
- Proceed to PR Create below.

If ANY check fails:
- List each failure with the specific fix needed.
- State: "Review FAILED for [feature-id]-[slug]. These issues must be addressed in a Build session before this feature can proceed."
- Do not create a PR. Move to the next feature if there are more to review.

CROSS-REVIEW (when applicable)
If you are a different agent or model than the one that built AND initially reviewed this feature, your review evaluation above serves as the cross-review. State this explicitly in the report.

If you are the same agent that built the feature, note: "Self-reviewed only. Cross-review by a different agent is recommended before merge."

PR CREATE (only for features that passed review)
Read the PR template at /.github/pull_request_template.md if it exists.
Read the full diff of the feature branch against the target branch.

Produce the pull request:
- Title: concise imperative summary
- What: one sentence
- Why: link to /specs/[feature-id]-[slug].md
- Changes: logical changes grouped by area
- Testing checklist: check each box honestly — if any cannot be checked, leave it unchecked and explain
- Non-code checklist: same — honest checks only
- Verification: how a reviewer can verify beyond reading the diff
- Rollback: special steps or "Standard revert."
- If the diff exceeds 300 lines, state this and recommend splitting before opening.

Open the PR against the target branch.
State: "PR created for [feature-id]-[slug]. Ready for human review."

After completing review (and PR if applicable) for one feature:
- Ask: "Would you like to review another feature, or is this review session complete?"

If another feature: ask for the next feature name and repeat from the Review phase.
If complete: produce a session summary:
  - Features reviewed
  - Status of each: PR created / failed review (with reason)
  - Any features needing a Build session to address failures

State: "All features that passed review have open PRs. Remaining items need Build sessions to address review failures. Human review can proceed on the open PRs — estimated 10 minutes per PR using the non-code checklist."
