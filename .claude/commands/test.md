<!-- generated-from-metaprompt -->
You are writing tests for a feature that does not exist yet. Do not write implementation code.

Read the task file at: $ARGUMENTS
Read the linked spec file referenced in that task file.
Read the project's conventions file (AGENTS.md) for testing patterns.
Read existing test files in the relevant area to match the project's test style, naming, and structure.

For each acceptance criterion in the spec, write at least one test that:
- Asserts the expected behavior described in the criterion.
- Will FAIL right now because the feature has not been implemented.
- Uses a descriptive name that states the expected behavior in plain language.

Rules:
- Every acceptance criterion must have at least one corresponding test. No criterion left untested.
- Tests must fail. If a test passes before implementation, it is not testing new behavior — rewrite it.
- Do not write any implementation code. Not even stubs, helpers, or fixtures that implement feature logic.
- Follow the existing test patterns in this codebase exactly. Match file location, naming, imports, and structure.
- All pre-existing tests must still pass. Your new tests are the only ones that should fail.

After writing tests:
1. Run the full test suite.
2. Confirm: new tests fail, all existing tests pass.
3. Commit the test files with a message referencing the spec (example: "Add failing tests for [feature-name] per /specs/[feature-name].md").

After committing, state: "Failing tests committed. Label the issue status:tests-written. Next phase: Implement (Phase 4). Start in a fresh context window. Run one session per task."
