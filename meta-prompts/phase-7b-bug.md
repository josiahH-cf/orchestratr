<!-- role: canonical-source -->
<!-- phase: 7 -->
<!-- description: Log a bug discovered during any workflow phase -->
# Bug Track — Log a Bug (Any Phase)

Log a bug discovered during any workflow phase. This is lightweight and should not interrupt the current phase.

## Collect Bug Details

Gather these fields (infer from context when possible; ask only if ambiguous):

- **Description:** One sentence summary of what is wrong
- **Location:** `file:line` or component name
- **Phase found:** Which workflow phase discovered this
- **Severity:** `blocking` or `non-blocking`
  - **blocking** — prevents the current task from being completed. Must be resolved (via `/bugfix`) before the next task can proceed. `/continue` will prioritize blocking bugs before selecting the next task.
  - **non-blocking** — does not prevent the current task from completing. Logged for a later review cycle. Does not interrupt the current workflow.
- **Expected:** What should happen
- **Actual:** What does happen
- **Fix-as-you-go:** `yes` or `no`

## Log It

Append to `bugs/LOG.md` using:

```
### BUG-[NNN]: [short description]
- **Location:** [file:line]
- **Phase:** [phase number and name]
- **Severity:** [blocking | non-blocking]
- **Expected:** [expected behavior]
- **Actual:** [actual behavior]
- **Fix-as-you-go:** [yes | no]
- **Status:** open
- **Logged:** [date]
```

## Rules

- Do not fix the bug in this command.
- If fix-as-you-go is yes and trivial, note "recommended for immediate fix."
- If severity is blocking, state: "This blocks current task. Run `/bugfix` to resolve."
- If `bugs/LOG.md` does not exist, create it.
- Return to the calling phase immediately after logging.

State: "BUG-NNN logged. Severity: [X]. Returning to [current phase]."
