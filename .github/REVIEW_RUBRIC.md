# Review Rubric

Standard scoring rubric for all code reviews. Used by reviewer agents and human reviewers.

## Categories

### 1. Correctness

Does the code do what the spec says it should do?

- **PASS**: All acceptance criteria met, edge cases handled, no logic errors
- **FAIL**: Any AC unmet, incorrect behavior, unhandled edge cases

Evidence required: Test results mapped to ACs.

### 2. Test Coverage

Are acceptance criteria adequately tested?

- **PASS**: Every AC has at least one test, tests are meaningful (not trivially passing), edge cases tested
- **FAIL**: Missing tests for any AC, tests that pass without the feature, no edge case coverage

Evidence required: Test names mapped to AC IDs.

### 3. Security

Is the code free from common vulnerabilities?

- **PASS**: No secrets in code, inputs validated at boundaries, no injection vectors, dependencies audited
- **FAIL**: Hardcoded secrets, unvalidated user input, SQL/XSS/command injection possible, known vulnerable dependencies

Evidence required: Security checklist completed.

### 4. Performance

Does the code avoid unnecessary performance problems?

- **PASS**: No obvious N+1 queries, no unbounded loops on user data, appropriate caching, reasonable memory usage
- **FAIL**: Quadratic or worse algorithms on user data, memory leaks, missing pagination, synchronous blocking on I/O

Evidence required: Note any perf-sensitive paths and their approach.

### 5. Style

Does the code follow project conventions?

- **PASS**: Matches existing patterns, naming conventions followed, file organization consistent, linting passes
- **FAIL**: Inconsistent naming, wrong file location, linting failures, dead code or unused imports

Evidence required: Linter output clean. Note: style issues are flagged but do not block approval alone.

### 6. Documentation

Are changes documented appropriately?

- **PASS**: Spec updated if behavior changed, decisions logged, PR description complete, non-obvious code commented
- **FAIL**: Spec drift (behavior doesn't match spec), missing decision records, empty PR description

Evidence required: Spec diff if behavior changed.

## Scoring

| Category | Weight | PASS/FAIL |
|----------|--------|-----------|
| Correctness | Required | |
| Test Coverage | Required | |
| Security | Required | |
| Performance | Advisory | |
| Style | Advisory | |
| Documentation | Required | |

**Required** categories must PASS for approval. **Advisory** categories are flagged but don't block.

## Final Verdict

- **APPROVE**: All Required categories PASS
- **REQUEST CHANGES**: Any Required category FAILS — list specific items to fix
- **COMMENT**: Only Advisory categories flagged — approve with notes
