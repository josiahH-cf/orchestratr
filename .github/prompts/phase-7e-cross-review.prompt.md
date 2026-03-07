---
agent: agent
description: 'Independent review by a different agent or model (OPTIONAL)'
---
<!-- role: derived | canonical-source: meta-prompts/phase-7e-cross-review.md -->
<!-- generated-from-metaprompt -->

[AGENTS.md](../../AGENTS.md)
[workflow/PLAYBOOK.md](../../workflow/PLAYBOOK.md)
[workflow/FILE_CONTRACTS.md](../../workflow/FILE_CONTRACTS.md)

You are performing an independent review of a completed feature branch. You did not implement this code and you did not perform the initial review. Your judgment is independent.

Read the spec at: ${input:specOrFeature:Provide the spec path or feature description}
Read `/AGENTS.md`, `/workflow/PLAYBOOK.md`, and `/workflow/FILE_CONTRACTS.md`.
Read the full diff of the current branch against the target branch.

Perform the same evaluation as a standard review:

**For each acceptance criterion in the spec:**
- Does a test exist that verifies this criterion?
- Does the implementation actually satisfy the criterion (not just the test)?
- PASS: [criterion] — [evidence]
- FAIL: [criterion] — [what is wrong and where]

**Code quality (flag any issues found):**
- Functions over 50 lines
- Missing error handling
- Hardcoded secrets or environment-specific values
- Files changed outside the declared scope
- Dead code or unused imports
- Naming inconsistencies with the existing codebase
- Tests that were modified rather than implementation being fixed

**Your independent assessment:**
- APPROVE: All criteria met, code quality acceptable. Ready for PR.
- REQUEST CHANGES: [list specific changes needed]

If APPROVE:
  State: "Cross-review passed. Label the issue status:reviewed. Next phase: PR Create."

If REQUEST CHANGES:
  State: "Cross-review found issues. Return to the Implement phase in a fresh context window to address the items listed above."
