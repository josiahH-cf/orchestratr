# Acceptance Criteria Reference

Use this document as a reference when writing acceptance criteria for feature specs. All criteria should be machine-parseable and verifiable.

## EARS Format (Event-Action-Response-State)

Use for system behavior triggered by events:

```
When [trigger], the system shall [response].
```

Examples:

- When a user submits the login form with valid credentials, the system shall redirect to the dashboard within 2 seconds.
- When the API receives a malformed request, the system shall return HTTP 400 with a descriptive error message.
- When the session token expires, the system shall prompt for re-authentication without losing form state.

## GWT Format (Given-When-Then)

Use for scenario-driven criteria:

```
Given [context/precondition],
when [action/trigger],
then [expected outcome].
```

Examples:

- Given a user with admin role, when they access the user management page, then all CRUD operations are available.
- Given an empty shopping cart, when the user clicks "checkout", then the system displays a message indicating the cart is empty.
- Given a network timeout, when the client retries the request, then the retry uses exponential backoff with a maximum of 3 attempts.

## Verification Pattern

Every criterion must have a concrete verification command:

```
Verification: [INSERT COMMAND] passes
```

Examples:

- Verification: `npm test -- --grep "login redirect"` passes
- Verification: `curl -X POST /api/login -d '{}' | jq .status` returns `400`
- Verification: `pytest tests/test_auth.py::test_session_expiry` passes

## Machine-Parseable Checkbox Format

Use this format in spec files for tracking:

```markdown
- [ ] AC-1: Login redirect — When valid credentials submitted, redirect to dashboard. Verification: `npm test -- --grep "login redirect"` passes
- [ ] AC-2: Invalid request handling — When malformed request received, return 400. Verification: `curl -X POST /api -d '{}' -w '%{http_code}'` returns 400
- [ ] AC-3: Session expiry — When token expires, prompt re-auth. Verification: `pytest tests/test_auth.py::test_session_expiry` passes
```

## Criterion Naming

- Use format: `AC-N: [short name]`
- N is sequential within the feature spec (AC-1, AC-2, AC-3...)
- Short name should be 2–4 words describing the behavior
- Same AC-N is referenced in task breakdown (T-N maps to AC-N)

## Completeness Checklist

Before finalizing criteria for a feature spec:

- [ ] Every user story has at least one AC
- [ ] Every AC has both EARS/GWT description AND verification command
- [ ] Every AC is independently testable
- [ ] Non-functional requirements (performance, security) have measurable thresholds
- [ ] Edge cases and error states are covered, not just happy paths
