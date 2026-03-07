# Workflow Lint Contract

Non-destructive linting and observability for workflow artifacts. Callouts only — no automatic file mutation.

> Referenced from `AGENTS.md`. This is part of the canonical workflow — see `/governance/REGISTRY.md`.

## Lint Categories

| Category | Description | Severity |
|----------|-------------|----------|
| Orphan | File exists in a tracked directory but has no inbound reference | Warning |
| Length | File exceeds 120 lines | Warning |
| EOF Newline | Text file does not end with a single newline | Warning |
| Clarity | Structural or readability heuristic failed | Info |

## Check Definitions

### 1. Orphan Detection

Scan `/specs/`, `/tasks/`, `/decisions/`, and `/workflow/` for markdown files not referenced by any other tracked file (AGENTS.md, LIFECYCLE.md, PLAYBOOK.md, FILE_CONTRACTS.md, STATE.json, or peer markdown files).

- Template files (`_TEMPLATE.md`) are exempt.
- `STATE.json` `currentTaskFile` counts as a reference.

### 2. Length Warning

Flag any markdown file under `/workflow/`, `/governance/`, `/specs/`, `/tasks/`, `/decisions/`, or `.specify/` that exceeds 120 lines.

### 3. EOF Newline

Every tracked text file (`.md`, `.json`, `.sh`, `.yml`, `.yaml`, `.toml`) must end with exactly one newline character. Flag files that:
- End without a newline.
- End with multiple trailing newlines.

### 4. Clarity Heuristics

| Heuristic | Applies To | Trigger |
|-----------|-----------|---------|
| Missing H1 | All `.md` files | No `# ` line in first 5 lines |
| Missing required sections | `/specs/*.md`, `/tasks/*.md` | Spec missing `## Acceptance Criteria`; task missing `## Tasks` |
| TODO/FIXME markers | All tracked files | Contains `TODO` or `FIXME` (informational) |
| Sentence length outlier | `.md` files | Any non-code line exceeds 200 characters |
| Passive voice density | `.md` files | More than 40% of sentences in a file use passive indicators (`is being`, `was`, `were`, `been`, `be done`) |

## Output

The lint script produces a Markdown report written to `/workflow/LINT_REPORT.md` with:
- Timestamp and file count scanned.
- Grouped findings by category.
- Per-finding: file path, line number (where applicable), category, message.
- Summary counts.

## Non-Mutation Guarantee

The lint process **never** modifies, deletes, or creates project source files. Its only write is the report file itself.

## Exit Behavior

The lint script always exits `0`. Findings are advisory, not blocking. CI pipelines should treat lint as a non-blocking informational step.

## Invocation

```bash
scripts/workflow-lint.sh [project-root]
```

Defaults to current directory if no argument is provided.
