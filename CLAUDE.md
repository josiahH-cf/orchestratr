@AGENTS.md

# Claude Code — Session Rules

## Context Discipline

- Start every task in a fresh session
- Compact at 60 percent context usage — do not wait until 95 percent
- If you compact more than twice, stop — the task is too large
- Include only files the current task touches via @ references
- Never carry planning context into implementation — write to a file, clear, restart

## Planning

- Begin complex work in Plan Mode
- Write plans to `/specs/` or `/tasks/` — never keep plans only in chat

## Testing

- Commit tests separately before implementation
- Use a subagent for test verification when the test suite is large

## Implementation

- One task per session
- Orient first: read the task file, relevant source files, and test file before writing
- Commit after each task passes its tests
- Follow existing patterns — read before writing
- When uncertain, write the decision to `/decisions/` before proceeding

## Scope Discipline

- Do not add code style rules here — linters enforce style
- Do not generate entire modules in a single pass
- Do not refactor code outside the current task scope

## Personal Overrides

- Use `/CLAUDE.local.md` for personal preferences (gitignored)
- Do not put project rules in CLAUDE.local.md

## File References

- Specs: `/specs/[feature-name].md`
- Tasks: `/tasks/[feature-name].md`
- Decisions: `/decisions/[NNNN]-[slug].md`
- ExecPlan template: `/.codex/PLANS.md`
