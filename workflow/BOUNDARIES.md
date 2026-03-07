# Boundaries & Bug Tracking

> Referenced from `AGENTS.md`. This is part of the canonical workflow — see `/governance/REGISTRY.md`.

## Boundaries

### Best Practices

- Read the spec before implementation
- Run tests before commit
- Include acceptance criteria evidence in every PR
- Follow branch naming conventions
- Reference the constitution for alignment on design decisions

### Recommended Review Points

These are areas where extra awareness helps — not approval gates. All items below use the **Suggest** advisory tier (see `workflow/ROUTING.md → Advisory Tier Reference`):

- Adding new dependencies
- Modifying CI workflows
- Changing the constitution — consider using `/compass-edit` for traceability. Exception: Phase 2 Compass directly populates `.specify/constitution.md` during initial discovery.
- Modifying `AGENTS.md`
- Architectural decisions not covered by the spec

### Avoid

- Committing secrets or `.env` files
- Modifying files outside the assigned scope
- Skipping tests
- Making decisions not traceable to a spec or constitution principle

## Bug Tracking

Use `/bug` (Claude) or `phase-7b-bug.prompt.md` (Copilot) from any phase:

```
Description: [what's wrong]
Location: [file:line or component]
Phase found: [which phase discovered this]
Severity: blocking | non-blocking
Expected: [what should happen]
Actual: [what does happen]
Fix-as-you-go: yes | no
```

- Small bugs (fix-as-you-go = yes): fix in place, log the fix
- Large bugs: add to backlog as a spec, assign model + branch when picked up
- Backlog review cycle: treat queued bugs as specs with full AC/branch/model assignment
