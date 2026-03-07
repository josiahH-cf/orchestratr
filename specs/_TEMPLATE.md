# Feature: [feature-id]-[slug]

**Feature ID:** [issue-id]-[slug]

## Description

<!-- 2–3 sentences: what this does and why -->

## Acceptance Criteria

<!-- 3–7 testable criteria using GIVEN/WHEN/THEN (GWT) format.
     Each criterion must be independently verifiable by an automated test.

     Format:
       GIVEN [precondition or initial state],
       WHEN [action or event],
       THEN [expected outcome or observable result].

     EARS extensions (optional):
       WHILE [ongoing state], WHEN [trigger], THEN [outcome]  — for state-dependent behavior
       WHERE [feature is supported], WHEN [action], THEN [outcome]  — for conditional features
       IF [condition], THEN [outcome]  — for unconditional requirements
-->

- [ ] **AC-1:** GIVEN [precondition], WHEN [action], THEN [expected outcome]
- [ ] **AC-2:** GIVEN [precondition], WHEN [action], THEN [expected outcome]
- [ ] **AC-3:** GIVEN [precondition], WHEN [action], THEN [expected outcome]

## Affected Areas

<!-- Files, modules, or directories this touches -->

- [path or module]

## Constraints

<!-- Performance targets, backward compatibility, security requirements -->

## Out of Scope

<!-- Explicitly excluded to prevent scope creep -->

## Dependencies

<!-- Other features, services, or data this depends on -->

## Test Planning Intent

<!-- Populated during Phase 3 (define-features). Inferred from acceptance criteria and affected areas.
     Do NOT add test file paths or function names here — that belongs in the Verification Map during Phase 7. -->

**Test approach:** <!-- e.g., unit, integration, e2e, visual -->

**Rough test scenarios:**
<!-- Brief natural-language descriptions of key scenarios to verify -->
- [ ] [Scenario description, e.g., "verify login rejects expired tokens with a 401"]
- [ ] [Scenario description]

## Verification Map

<!-- Map each criterion to its exact test location. Every AC must have a test. -->

| AC | Test File | Test Function | Assertion |
|----|-----------|---------------|-----------|
| AC-1 | [path/to/test_file.py] | [test_function_name] | [what the test asserts] |
| AC-2 | [path/to/test_file.py] | [test_function_name] | [what the test asserts] |
| AC-3 | [path/to/test_file.py] | [test_function_name] | [what the test asserts] |

## Contracts

<!-- Interfaces, types, or API contracts this feature exposes or consumes.
     Define these BEFORE implementation to enable parallel agent work.
     Leave empty if this feature has no public interfaces. -->

- Exposes: [interface/type/endpoint]
- Consumes: [interface/type/endpoint]

## Execution Linkage

<!-- Execution planning is authoritative in /tasks/[feature-id]-[slug].md -->

- Task file: /tasks/[feature-id]-[slug].md
- Task ordering, model assignment, and branch naming live in the task file

## Notes

<!-- Non-obvious details the implementer should know -->
