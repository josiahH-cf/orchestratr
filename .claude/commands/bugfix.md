<!-- role: derived | canonical-source: meta-prompts/phase-7c-bugfix.md -->
<!-- generated-from-metaprompt -->
# Bugfix — Reproduce -> Diagnose -> Fix -> Verify -> PR

Fix a logged bug following the structured loop:
reproduce -> diagnose -> fix -> verify -> ship.

Read the bug entry referenced by `$ARGUMENTS` (path to bug log or bug ID like `BUG-NNN`).
Read `/AGENTS.md` (Boundaries, Bug Tracking, Branch Naming).
Read `.specify/constitution.md` and verify the fix aligns with project principles.

## Step 1: Reproduce

- Write a failing test that demonstrates the bug's expected vs actual behavior.
- The test must fail before the fix and pass after.
- If the bug cannot be reproduced, document why and close as "cannot reproduce."

## Step 2: Diagnose

- Read the code at the bug location.
- Identify the root cause (not only the symptom).
- If the root cause suggests a broader design issue, log a follow-up bug for that deeper issue and fix only the current bug here.

## Step 3: Fix

- Fix the root cause with the smallest safe change.
- Do not add functionality beyond restoring intended behavior.
- If a fix requires an unplanned architecture decision, record it in `/decisions/[NNNN]-[slug].md`.

## Step 4: Verify

- Run the failing test and confirm it now passes.
- Run the full test suite and confirm no regressions.
- If regressions appear, fix them without modifying unrelated tests.

## Step 5: Ship

- Branch: `model/bug-short-description` (per AGENTS branch naming).
- Commit message: `Fix BUG-NNN: [short description]`.
- Update bug log status to `fixed` and add fix date + commit reference.
- Create a PR with bug reference and test evidence.

State: "BUG-NNN fixed. Test added. PR ready."
