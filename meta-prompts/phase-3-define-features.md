<!-- role: canonical-source -->
<!-- phase: 3 -->
<!-- description: Translate constitution into concrete features with priorities -->
# Phase 3 — Define Features

**Objective:** Translate the Compass constitution into a concrete, prioritized feature set. Each feature maps to a constitution capability.

**Trigger:** Phase 2 complete (constitution populated).

**Entry commands:**
- Claude: `/define-features`
- Copilot: `phase-3-define-features.prompt.md`

---

## What Happens

### Step 1 — Load Context
1. Read `.specify/constitution.md` (all capabilities)

### Step 2 — Interview First (mandatory before any feature proposals)
Do **not** propose features until the developer has answered discovery questions and confirmed intent.

Begin with these critical-thinking questions:
- **Problem fit:** "Which constitution capabilities feel highest-risk or least understood? Where do you expect the hardest trade-offs?"
- **User reality:** "Who will use each capability day-to-day, and what does their current workflow look like without it?"
- **Conflict detection:** "Do any capabilities overlap, contradict, or compete for the same surface area? Which wins if they conflict?"
- **Context gaps:** "Are there areas where you can describe *what* you want but not *how* it should look, feel, or behave? What's blocking that clarity?"
- **Boundary scoping:** "What is explicitly *not* part of this project even if it seems related?"

Then continue with open-ended follow-up:
- Ask the developer what else should be discussed before defining features.

**Drill-down rule:** If any answer reveals ambiguity, conflicting capabilities, missing context for UI/UX placement, unclear data flow, or competing priorities — do not move on. Ask targeted follow-up questions that resolve the specific gap before proceeding. Keep clarifying until the developer confirms the picture is clear enough to define features.

### Step 3 — Define Features
3. For each feature, produce: name, description, Compass capability mapping, priority, scope flag
4. Features that don't map to a constitution capability → explicit deferral or Compass reconsideration
5. No implementation details — this is feature definition, not architecture

### Step 4 — Test Planning Intent
6. For each feature, infer a **Test Planning Intent** section from the feature's acceptance criteria and affected areas:
   - **Test approach:** categories of testing needed (unit, integration, e2e, visual, etc.)
   - **Affected areas:** files, modules, or directories the feature touches
   - **Rough test scenarios:** brief natural-language descriptions of key scenarios to verify (e.g., "test that expired tokens are rejected at login", "verify dashboard loads within 2s with 1000 records")
   - Do **not** specify test file paths, function names, or write executable tests — that belongs to Phase 7

## Gate

- Developer interview completed — discovery questions answered and ambiguities resolved before features were proposed
- At least one feature spec exists in `/specs/` with Compass mapping
- Every constitution capability has at least one feature mapping
- No orphan features (every feature traces to a capability)
- Every feature spec includes a Test Planning Intent section with test approach, affected areas, and rough test scenarios

### Mandatory Completion Checklist

Before concluding this phase, verify all of the following:

- [ ] Developer interview completed with discovery questions answered before any features were proposed
- [ ] At least one `specs/[feature-id]-[slug].md` file has been created (not just discussed)
- [ ] An explicit capability-to-feature coverage list is present (each constitution capability maps to at least one feature)
- [ ] Orphan-feature disposition is documented (any proposed feature that does not trace to a capability is explicitly deferred or triggers a Compass reconsideration)
- [ ] Every feature spec has a populated **Test Planning Intent** section (test approach, affected areas, rough test scenarios)
- [ ] If any of the above are missing, do NOT advance the phase — produce the missing artifacts first

## Output

- Feature specs in `/specs/[feature-id]-[slug].md`
- Feature priority order
- Task planning is deferred to `/tasks/[feature-id]-[slug].md` in Phase 5

## See Also

- Constitution: `.specify/constitution.md`
- Spec template: `.specify/spec-template.md`
