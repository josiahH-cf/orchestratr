# Feature Lifecycle

This file defines the end-to-end path from idea to merged code.
Every phase produces a named artifact; the next phase consumes it.

| Phase | Input | Action | Output | Who |
|-------|-------|--------|--------|-----|
| 0. Ideate | Raw idea | Human writes GitHub Issue | Issue (labeled `status:idea`) | Human |
| 1. Scope | Issue | Scoping process | `/specs/[name].md` + label → `status:scoped` | Any agent |
| 2. Plan | Spec | Planning process | `/tasks/[name].md` + label → `status:planned` | Any agent |
| 3. Test | Tasks | Test authoring process | Committed failing tests + label → `status:tests-written` | Any agent |
| 4. Implement | Tests + Tasks | Implementation process (per task) | Passing code + label → `status:implemented` | Any agent |
| 5. Review | Branch diff + Spec | Review process + cross-agent review | PASS/FAIL report + label → `status:reviewed` | Different agent |
| 6. PR | Review pass | PR creation process | Open PR | Any agent |
| 7. Merge | Approved PR | Human merges | Merged + branch deleted + label → `status:done` | Human |

## Label Conventions

- `status:idea` — raw, unscoped
- `status:scoped` — spec written, ready to plan
- `status:planned` — tasks written, ready to build
- `status:tests-written` — failing tests committed
- `status:implemented` — all tasks complete, tests pass
- `status:reviewed` — cross-agent review passed
- `status:done` — merged

- `size:S` / `size:M` / `size:L` — effort estimate
- `agent:claude` / `agent:codex` / `agent:copilot` — assigned agent (optional)

## Bulk Issue Creation

When creating multiple features at once:
1. Write all issues first using the Issue Template — do not scope yet
2. Label each `status:idea` and assign a `size:` label
3. Scope them one at a time in separate agent sessions
4. Each issue must be independently actionable — no implicit ordering between issues
5. If issues depend on each other, note the dependency in the issue body and scope the dependency first
