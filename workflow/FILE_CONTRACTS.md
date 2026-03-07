## Artifact Directory Map

```
.specify/
  constitution.md                    ← project identity (written in Phase 2)
  spec-template.md                   ← copy to /specs/ for each new feature
  acceptance-criteria-template.md    ← AC format reference (EARS + GWT)

/specs/[feature-id]-[slug].md        ← one per feature (created in Phase 3+)
/tasks/[feature-id]-[slug].md        ← one per feature (authoritative execution artifact; created in Phase 5+)
/decisions/[NNNN]-[slug].md          ← architecture choices (any phase)
/bugs/LOG.md                         ← append-only bug log (any phase)
/reviews/[feature-id]-bot-findings.md ← bot review findings (Phase 7a, on FAIL only)
/.codex/PLANS.md                     ← Codex long-run execution plan (Phase 6+)
/workflow/STATE.json                 ← orchestration state for /continue
```

Feature ID format: `[issue-id]-[slug]` (example: `42-user-auth`). Spec and task files share the same feature ID.

# File Contracts

Machine-readable behavior starts with stable artifact contracts.

## Artifact Contracts

| Artifact | Owner | Updated When | Required Fields | Validation Signal |
| -------- | ----- | ------------ | --------------- | ----------------- |
| `.specify/constitution.md` | Compass interview | Phase 2 (Compass) or `/compass-edit` | Theme-based sections covering: Problem & Context, Target User, Success Criteria, Core Capabilities, Out-of-Scope, Principles, Security, Testing (depth varies by project); Ambiguity Tracking section | No `[PROJECT-SPECIFIC]` placeholders remain in covered sections; ambiguities documented |
| `AGENTS.md` (per-section) | See breakdown | See breakdown | See breakdown | See breakdown |
| — `AGENTS.md → Overview` | Compass (Phase 2) | Phase 2 or `/compass-edit` | One-paragraph project description | No `[PROJECT-SPECIFIC]` placeholder |
| — `AGENTS.md → Workflow Phases` | Scaffold template (Phase 1) | Phase 1 (static) | Phase listing with entry commands and gates | Sections present and consistent with LIFECYCLE.md |
| — `AGENTS.md → Quick Reference` | Scaffold template (Phase 1) | Phase 1 (static) | Routing table to workflow/ and governance/ files | All referenced files exist |
| `workflow/COMMANDS.md → Core Commands` | Phase 1 initial; Scaffold (Phase 4) finalizes | Phase 1 (initial values), Phase 4 (refined based on architecture) | Install, Build, Test, Lint, Format, Type-check | No `[PROJECT-SPECIFIC]` after Phase 4 |
| `workflow/COMMANDS.md → Code Conventions` | Phase 1 initial; Scaffold (Phase 4) finalizes | Phase 1 (initial values), Phase 4 (refined based on architecture) | Language, style guide, architecture patterns, naming conventions | No `[PROJECT-SPECIFIC]` after Phase 4 |
| `/specs/[feature-id]-[slug].md` | Builder agent | Scope changes | Feature ID, criteria IDs, affected areas, out-of-scope, Compass mapping | Criteria count 3–7 and IDs present |
| `/tasks/[feature-id]-[slug].md` | Builder agent | Plan + every task completion + session end | Task IDs, criterion mapping, status counts, routing plan, session log (structured), branch name | All criteria mapped, status math consistent, latest Session Log row reflects current state |
| `/decisions/[NNNN]-[slug].md` | Builder/reviewer/human | Non-obvious forks or conflicts | Trigger, options, decision, consequences, rollback impact | Decision linked from task/spec when needed |
| `/bugs/LOG.md` | Any agent via `/bug` | Bug discovered in any phase | BUG-NNN, description, location, phase, severity, expected, actual, status | Sequential BUG-NNN IDs, status field present |
| `/.codex/PLANS.md` (instance copy) | Builder agent | Long-run execution only | Milestones, verification, progress | Milestones map to task IDs |
| `/workflow/STATE.json` | Orchestrator (`/continue`) | Phase transitions and task selection | `projectPhase`, `currentFeatureId`, `currentTaskFile`, `testMode`, `maintenanceLevel`, `advisoryProfile`, `activeClaims`, `updatedAt` (`schemaVersion` optional) | Valid JSON and phase/task references resolve; `maintenanceLevel` is one of `light`, `standard`, `deep`, or empty; `advisoryProfile` is one of `concise`, `standard`, `detailed`, or empty; `activeClaims` is an array of objects with `taskFile`, `agent`, `claimedAt`, `lockedFiles` |
| Review report in PR body | Reviewer agent | Review phase | PASS/FAIL per criterion + scope checks + rubric scores | No unchecked criterion |
| `/reviews/[feature-id]-bot-findings.md` | Review bot agent | Bot review FAIL | Date, feature ID, verdict, rubric scores (6 categories), AC results, issues list with file:line references, re-entry instructions | All issues have category + location + fix description; links to spec, task, and rubric are valid |
| `workflow/ROUTING.md` | Human maintainer | Agent model changes | Advisory routing hints, branch format, concurrency rules, multi-agent claim protocol | Tables present and non-empty; no enforced model-to-task bindings |
| `workflow/COMMANDS.md` | Phase 1 initial; Scaffold (Phase 4) finalizes | Phase 1 (initial values), Phase 4 (architecture-refined) | Command table, conventions | No `[PROJECT-SPECIFIC]` after Phase 4 |
| `workflow/BOUNDARIES.md` | Human maintainer | Policy changes | Best Practices, Recommended Review Points, Avoid sections | All three sections present |
| `workflow/LINT_CONTRACT.md` | Human maintainer | Lint rule changes | Lint categories, check definitions, output format, non-mutation guarantee | All four categories defined |
| `workflow/LINT_REPORT.md` | `scripts/workflow-lint.sh` | Each lint run | Timestamp, findings by category, summary counts | Generated file; not manually edited |

## Linkage Rules

- Feature ID format: `[issue-id]-[slug]` (example: `42-user-auth`).
- Spec filename and task filename share the same feature ID.
- Acceptance criteria IDs use `AC-1..N`.
- Tasks use `T-1..N` and must cite covered `AC-*` IDs.
- Decisions referenced by ID in tasks or PR when they affect behavior.
- `/workflow/STATE.json` `currentTaskFile` must point to an existing `/tasks/[feature-id]-[slug].md`.

## Invalid State Conditions

Treat as blocking failures:

- Spec exists without a matching task file.
- Any criterion has no test mapping.
- Task status counts do not match task checklist states.
- Out-of-scope files changed with no decision record.
- PR missing verification or rollback sections.
- `workflow/STATE.json` points to a missing task file or invalid phase identifier.

## Specification Writing

All work flows from specifications:

1. Read `.specify/constitution.md` before any implementation — it is the project's identity
2. Check `/specs/` for the current feature spec (copy `.specify/spec-template.md` when creating new specs)
3. Create or update `/tasks/[feature-id]-[slug].md` during Phase 5 — this is the authoritative execution artifact
4. Write acceptance criteria using `.specify/acceptance-criteria-template.md` as reference
5. Verify all acceptance criteria pass before creating a PR

### Acceptance Criteria Notation

This project uses **GWT (Given/When/Then)** notation for all acceptance criteria, based on the EARS (Easy Approach to Requirements Syntax) framework.

#### Format

```
GIVEN [precondition or initial state],
WHEN [action or event],
THEN [expected outcome or observable result].
```

#### EARS Extensions

| Pattern | Use When | Example |
|---------|----------|----------|
| `GIVEN...WHEN...THEN` | Standard behavior | GIVEN a logged-in user, WHEN they click logout, THEN the session is destroyed |
| `WHILE...WHEN...THEN` | State-dependent behavior | WHILE the server is offline, WHEN a request arrives, THEN it is queued locally |
| `WHERE...WHEN...THEN` | Conditional features | WHERE dark mode is enabled, WHEN the page loads, THEN the dark theme is applied |
| `IF...THEN` | Unconditional requirement | IF the input is empty, THEN display a validation error |

#### Why GWT?

- **Translatable to tests:** Each GWT criterion maps directly to a test: the GIVEN is the setup, the WHEN is the action, the THEN is the assertion.
- **Machine-verifiable:** Agents can parse GWT format to auto-generate test skeletons.
- **Prevents vague criteria:** Forces specificity about preconditions and expected outcomes.

#### Anti-patterns

- ❌ `AC-1: The system should work correctly` — not testable
- ❌ `AC-1: Users should be able to log in` — no precondition or expected outcome
- ✅ `AC-1: GIVEN valid credentials, WHEN the user submits the login form, THEN the system redirects to the dashboard and sets a session cookie`
