# Implementation Plan Template

When a task requires more than 3 steps, create an implementation plan following
this structure. The plan is a living document  -  update it as work proceeds.
Treat the reader as having zero prior context.

---

## Purpose

<!-- One paragraph: what this accomplishes and why -->

## Linkage

- Feature ID: [issue-id]-[slug]
- Spec: `/specs/[feature-id]-[slug].md`
- Tasks: `/tasks/[feature-id]-[slug].md`
- Milestones cover tasks: [T-1, T-2, ...]

## Scope

<!-- Explicit boundaries: what IS and IS NOT included -->

## Prerequisites

- [ ] Tests green on the target branch
- [ ] Dependencies installed and up to date
- [ ] Spec exists at `/specs/[feature-name].md`

## Milestones

### Milestone 1: [name]

- Files: [list]
- Steps:
  1. [step]
  2. [step]
- Verification: [how to confirm completion]
- Tests: [how to validate]

### Milestone 2: [name]

<!-- Same structure -->

## Testing

- Approach: [full test suite execution method]
- Expected pass count before: [N]
- Expected pass count after: [N + new]
- Each milestone must leave all tests passing

## Rollback

<!-- How to safely revert if this fails partway -->

## Progress

- [ ] Milestone 1
- [ ] Milestone 2

## Decision Log

<!-- Non-obvious choices made during implementation -->

## Surprises

<!-- Anything unexpected encountered -->
