# Governance Change Protocol

This protocol controls changes to instruction and policy files.

## Principle

Core policy files are not edited ad hoc during feature delivery. Changes require a proposal artifact and independent review.

## Trigger

Open a policy change when recurring failures indicate missing or unclear instructions.

## Required Proposal

Create a decision record that includes:

- Triggering failure pattern
- Files to change
- Proposed policy diff summary
- Compatibility and portability impact
- Rollback plan
- Validation updates required in `/governance/POLICY_TESTS.md`

## Approval Flow

1. Builder agent drafts proposal.
2. Reviewer agent validates necessity and risk.
3. Human approves merge for governance changes.

## Merge Rules

- Governance changes must be isolated from feature code changes.
- PR must include before/after behavior statement.
- PR must include rollback instructions.
