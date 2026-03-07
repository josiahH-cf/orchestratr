# Failure Routing

Use this matrix when autonomous execution encounters failure.

> Responses in this matrix are **recommended actions** (Suggest tier per `workflow/ORCHESTRATOR.md → Advisory Tiers`). Agents may adapt based on context.

## Failure Matrix

| Failure Type | First Action | Second Action | Escalate When |
| ------------ | ------------ | ------------- | ------------- |
| Test failure caused by current task | Fix implementation within task scope | Add decision record if tradeoff is required | Failure persists after two focused attempts |
| CI/build tooling failure | Reproduce locally with exact command from AGENTS | Apply minimal env or config fix in scope | Root cause depends on external system access |
| Ambiguous requirement | Re-read spec and out-of-scope section | Propose a decision with options and impacts | Spec is contradictory or incomplete |
| Policy conflict across files | Follow policy precedence in AGENTS | Update task log with chosen precedence | Canonical docs appear internally inconsistent |
| Reviewer FAIL | Address listed failures and rerun checks | Update task evidence and PR verification | FAIL repeats with no new signal |
| Review bot FAIL | Read findings at `/reviews/[feature-id]-bot-findings.md`, address each issue | Run `/continue` to re-enter Phase 6, fix issues, re-run tests and bot review | Bot FAIL repeats after two fix cycles — escalate to manual `/review-session` |
| Compass gap (Phase 2) | Re-read constitution, identify which themes lack coverage | Propose targeted follow-up or constitution amendment via `/compass-edit` | Developer unavailable for interview |
| Feature-constitution mismatch (Phase 3) | Check if feature maps to any capability | Propose adding capability to constitution | Feature is valuable but outside current constitution |
| Architecture uncertainty (Phase 4) | Document as `[DECISION NEEDED]` in plan | Present tradeoff options with pros/cons | Developer input required for infrastructure choice |
| Blocking bug during implementation | Log via `/bug` with severity: blocking | Run `/bugfix` to attempt resolution. `/continue` will prioritize blocking bugs before the next task. | Bug requires design change or external fix |
| Non-blocking bug during implementation | Log via `/bug` with severity: non-blocking | Continue current task. Bug is deferred to a later review cycle. | Non-blocking bug is promoted to blocking (prevents task completion) |
| AC verification failure (Phase 7) | Re-read spec ACs, check test correctness | Log as bug, return to `/implement` | AC cannot be met with current architecture |

## Escalation Packet

When escalation is required, provide:

- Feature ID and branch
- Current phase
- Exact failing check/output
- What was tried
- Decision requested (single clear question)

## Stop Conditions

Stop autonomous execution and request human input when:

- A security/privacy decision is required.
- A destructive data migration is needed.
- Repo policy requires human approval for merge.
