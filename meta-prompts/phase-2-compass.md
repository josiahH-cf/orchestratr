<!-- role: canonical-source -->
<!-- phase: 2 -->
<!-- description: Adaptive interview to establish project identity and boundaries -->
# Phase 2 — Compass

**Objective:** Conduct an adaptive interview to establish the project's identity, goals, success criteria, and boundaries. This produces the constitution — the source of truth that all later phases reference.

**Trigger:** Phase 1 complete (scaffold files in place) or `/compass` invoked manually.

**Entry commands:**
- Claude: `/compass`
- Copilot: `phase-2-compass.prompt.md`
- Codex: see `.codex/AGENTS.md`

---

## What Happens

A dynamic discovery interview that starts broad and narrows only as context justifies specificity. This is not a scripted checklist — depth, ordering, and emphasis adapt to the developer's answers.

### Discovery Model

1. **Start broad** — Ask what the project is about, who it serves, and what success looks like. Let the developer frame the problem in their own terms.
2. **Follow the signal** — Pursue depth where the developer's answers reveal complexity, ambiguity, or risk. Skip areas that are self-evident or premature.
3. **Synthesize boundaries** — Establish what the project IS, what it is NOT, and what remains AMBIGUOUS.

### Guiding Themes

The interview should explore these themes as context warrants — not as mandatory sequential sections:

- **Problem & Context** — What problem does this project solve? Who experiences it?
- **Target User** — Who is the primary audience? What is their context?
- **Success Criteria** — How do we know the project succeeds?
- **Core Capabilities** — What must the project do? (features at the capability level)
- **Out-of-Scope Boundaries** — What will this project explicitly NOT do?
- **Inviolable Principles** — What rules can never be broken?
- **Security Posture** — What security standards apply?
- **Testing Strategy** — What testing standards apply?

Depth per theme varies — a CLI tool may need minimal security discussion while a healthcare app needs extensive coverage. The interview adapts.

### Personas (Optional)

If personas help the developer reason about users or use cases, introduce them. Do not require a fixed persona set — use them only when they clarify the problem.

### Ambiguity Tracking

Not everything can be resolved in one interview. Explicitly document:
- **Known** — what is decided and clear
- **Unknown** — what remains ambiguous and needs future resolution
- **Deferred** — what was intentionally postponed with rationale

## Gate

- `.specify/constitution.md` exists with all relevant themes addressed (no `[PROJECT-SPECIFIC]` placeholders remain in covered sections)
- Known ambiguities are documented rather than papered over with guesses
- `AGENTS.md → Overview` section is filled with project description

## Output

- `.specify/constitution.md` — project identity document with theme-based sections and ambiguity tracking
- `AGENTS.md → Overview` — one-paragraph project description
- `workflow/STATE.json → advisoryProfile` — auto-detected from interview signals (see below)

### Advisory Profile Detection

At the end of the Compass interview, set `advisoryProfile` in `workflow/STATE.json` based on observed signals:

| Signal | Profile |
|--------|---------|
| Simple project, experienced user, few ambiguities | `concise` |
| Moderate complexity, mixed signals | `standard` |
| Complex project, many unknowns, first-time user, detailed questions asked | `detailed` |

If signals are unclear, default to `standard`. The user can override at any time ("switch to concise/detailed").

### Ownership Boundary

Compass owns exactly two write targets:
1. `.specify/constitution.md` — full content
2. `AGENTS.md → Overview` — one-paragraph project description

Compass does **not** write to `workflow/COMMANDS.md` (Core Commands, Code Conventions). Those are owned by Phase 1 (initial values) and Phase 4 Scaffold (finalized values).

## Editing the Constitution

Use `/compass-edit` (Claude) to modify the constitution after initial creation. This requires explicit developer approval and shows a diff of changes.

## See Also

- Constitution template: `.specify/constitution.md`
- Edit gate: `.claude/commands/compass-edit.md`
