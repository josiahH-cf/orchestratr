<!-- role: canonical-source -->
<!-- phase: 1 -->
<!-- description: Import scaffold files into a project and establish baseline -->
# Phase 1 — Scaffold Import

**Objective:** Place workflow scaffold files into a new or existing project and establish the baseline structure.

**Trigger:** New project needs the agentic workflow, or existing project needs scaffold update.

**Context:** This phase is executed by the `admin/initialization.md` meta-prompt. It is listed here for lifecycle completeness.

---

## What Happens

1. Scaffold ZIP is detected and extracted
2. Template files placed with conflict resolution
3. Copilot prompt files installed to editor
4. Project conventions customized in AGENTS.md
5. Configuration files tailored (settings.json, CI, Codex)
6. Verification pass confirms all files in place

## Gate

- All scaffold files placed
- AGENTS.md has no unresolved placeholder brackets (except `[PROJECT-SPECIFIC]` in Overview — filled by Compass)
- `workflow/COMMANDS.md` has `[PROJECT-SPECIFIC]` placeholders in Core Commands and Code Conventions — filled by Phase 1 (initial) and Phase 4 (finalized)
- `.specify/` directory exists with 3 template files
- `/workflow/STATE.json` exists and is valid JSON

## Output

Project ready for Phase 2 (Compass).

## Auto-Advance

Phase 1 completion auto-triggers Phase 2 (Compass) — the project identity interview begins immediately.

## See Also

- Full initialization process: `meta-prompts/admin/initialization.md`
- Update existing scaffolds: `meta-prompts/admin/update.md`
