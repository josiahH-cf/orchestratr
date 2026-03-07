# Tasks: [feature-id]-[slug]

**Feature ID:** [issue-id]-[slug]
**Spec:** /specs/[feature-id]-[slug].md

## Status

- Total: [N]
- Complete: [N]
- Remaining: [N]
- Blocked: [N]

## Pre-Implementation Tests

<!-- List test files to create BEFORE coding. These are written during `/test pre` mode. -->

| AC | Test File | Status |
|----|-----------|--------|
| AC-1 | [path/to/test_file.py::test_name] | [ ] Not written |
| AC-2 | [path/to/test_file.py::test_name] | [ ] Not written |
| AC-3 | [path/to/test_file.py::test_name] | [ ] Not written |

## Task List

### T-1: [name]

- **Files:** [list]
- **Test File:** [test file covering this task's criteria]
- **Done when:** [one sentence]
- **Criteria covered:** [AC-1, AC-2]
- **Branch:** [agent/type-short-description]
- **Status:** [ ] Not started

### T-2: [name]

- **Files:** [list]
- **Test File:** [test file covering this task's criteria]
- **Done when:** [one sentence]
- **Criteria covered:** [AC-*]
- **Branch:** [agent/type-short-description]
- **Status:** [ ] Not started

### T-3: [name]

- **Files:** [list]
- **Test File:** [test file covering this task's criteria]
- **Done when:** [one sentence]
- **Criteria covered:** [AC-*]
- **Branch:** [agent/type-short-description]
- **Status:** [ ] Not started

## Routing Plan

<!-- Suggested model assignments. These are recommendations — the developer or /continue may override at execution time based on current conditions. -->

| Task | Suggested Model | Rationale | Reviewer | Parallel? | Context Needs |
|------|-----------------|-----------|----------|-----------|---------------|
| T-1 | [claude\|copilot\|codex] | [one-line reason] | [different model] | [yes/no + reason] | [small/medium/large] |
| T-2 | [claude\|copilot\|codex] | [one-line reason] | [different model] | [yes/no + reason] | [small/medium/large] |
| T-3 | [claude\|copilot\|codex] | [one-line reason] | [different model] | [yes/no + reason] | [small/medium/large] |

## Test Strategy

<!-- Which criterion is tested by which test file or suite -->

- AC-1: [test name/path]
- AC-2: [test name/path]
- AC-3: [test name/path]

## Evidence Log

<!-- Append after each task completion -->

- [YYYY-MM-DD] T-[N]  -  commands run: [build/lint/test], result: [pass/fail], notes: [short note]

## Session Log

<!-- Append after each session. /continue reads the latest entry + workflow/STATE.json to resume. -->
<!-- Format: date, last completed task, next action, blockers, link to state file -->

| Date | Last Completed | Next Action | Blockers | State Link |
|------|---------------|-------------|----------|------------|
| [YYYY-MM-DD] | [T-N or "plan created"] | [next task ID or phase action] | [none or description] | [workflow/STATE.json](../workflow/STATE.json) |
