<!-- role: derived | canonical-source: meta-prompts/phase-7-test.md -->
<!-- generated-from-metaprompt -->
# Phase 7 — Test & Mark Changes

**Objective:** Verify feature behavior against acceptance criteria in explicit test modes (`pre` and `post`), log failures as bugs, and confirm behavior matches spec.

**Trigger:** `pre` mode runs before Phase 6 implementation. `post` mode runs after Phase 6 implementation.

**Entry commands:**
- Claude: `/test`
- Copilot: `phase-7-test.prompt.md`

---

## What Happens

### Pre-Implementation Mode (`/test pre`)
1. Read spec ACs and existing test patterns
2. Write at least one test per AC using EARS/GWT format
3. Tests must fail before implementation exists
4. Include UI/visual tests where applicable
5. Commit failing tests

### Post-Implementation Mode (`/test post`)
1. Run full test suite
2. Compare results against each AC — mark pass/fail
3. Log failures as bugs via `/bug`
4. If behavior deviated from spec but tests pass, update spec notes
5. Verify no regressions

## Gate

- `pre` mode: failing tests exist for every AC
- `post` mode: all acceptance criteria verified (pass or documented failure)
- Bug log reviewed — no blocking bugs remain
- No regressions in existing tests

## Output

- Test results mapped to ACs
- Bug log entries for any failures
- Updated spec if behavior deviated

## See Also

- Bug logging: `/bug` command
- Bug fixing: `/bugfix` command
