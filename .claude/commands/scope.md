<!-- generated-from-metaprompt -->
You are scoping a new feature. Do not write code. Do not create implementation files.

The feature is: $ARGUMENTS

Read the project's conventions file (AGENTS.md) before starting.

Explore the codebase to understand:
1. What files, modules, and areas are relevant to this feature.
2. What existing patterns and conventions apply to this area.
3. What dependencies, constraints, or risks exist.

Then produce a spec file at /specs/[feature-name].md with exactly this structure:

# Feature: [name]

## Description
[2–3 sentences: what this does and why]

## Acceptance Criteria
[3–7 testable statements. Each must be verifiable by an automated test. Write them as checkboxes.]

## Affected Areas
[Files, modules, or directories this will touch]

## Constraints
[Performance targets, backward compatibility, security requirements — or "None"]

## Out of Scope
[What is explicitly excluded to prevent scope creep]

## Dependencies
[Other features, services, or data this depends on — or "None"]

## Notes
[Non-obvious details the implementer should know — or "None"]

Rules:
- Acceptance criteria must be between 3 and 7. No more.
- If you cannot get below 8 criteria, this feature is too large. Instead of writing the spec, recommend how to split it into smaller features and stop. Each sub-feature will need its own Phase 0 issue and its own Phase 1 scope session.
- Every criterion must be a concrete, testable statement — not vague ("works correctly") or procedural ("run the tests").
- Do not include implementation details in the criteria. Describe what, not how.
- The spec file is the only output. Do not produce task files, test files, or code.

After writing the spec, state: "Spec complete. Label the issue status:scoped. Next phase: Plan."
