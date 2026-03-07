# Project Initialization  -  Meta-Prompt

Paste this into a coding agent session at the root of any project where a scaffold zip has been placed. Works with any of the three ZIP variants: `scaffold-template.zip`, `scaffold-metaprompts.zip`, or `scaffold-full.zip`.

> **Updating an existing project?** Use the [Update Meta-Prompt](update.md) instead  -  it preserves your customizations while updating generated files.

> **Workflow:** This initialization covers Phase 1 (Scaffold Import) of the 8-phase agentic workflow. After initialization completes, Phase 2 (Compass) is auto-triggered to begin the project identity interview. The full lifecycle: Scaffold Import → Compass → Define Features → Scaffold Project → Fine-tune Plan → Code → Test → Maintain (+ parallel Bug Track). See `AGENTS.md → Workflow Phases` for details.

---

```text
You are initializing a project with a standard development scaffolding. A zip archive has been placed somewhere in this project directory. Find it.

Work through the following steps in order.

---

STEP 1  -  LOCATE, DETECT VARIANT, AND EXTRACT

Find the scaffold zip file in the project directory. The zip may be named scaffold-template.zip, scaffold-metaprompts.zip, scaffold-full.zip, or a custom name.

Extract its contents to a temporary location so you can inspect them before placing anything.

Determine which variant this is by inspecting the contents:
- If it contains AGENTS.md → template content is present.
- If it contains a prompts/ directory with .prompt.md files → metaprompt slash commands are present.
- If it contains both → this is the full variant.

Report: "Detected ZIP variant: [template-only / metaprompts-only / full]. Contents: [N] files."

List every file the zip contains with its intended destination path relative to the project root. Present this list and confirm: "These files will be placed in your project. Proceeding to check for conflicts."

---

STEP 2  -  PLACE TEMPLATE FILES, RESOLVE CONFLICTS

If the ZIP contains template content (AGENTS.md, .claude/, .github/, etc.):

For each template file in the extracted archive (excluding the prompts/ directory):
1. Check whether a file already exists at the target path in the project.
2. If no conflict: place the file silently.
3. If a conflict exists: show a comparison of what differs between the existing file and the template file. Then ask one of the following based on the file type:
   - For additive files (like .gitignore additions): "This file contains additions meant to be appended. Should I append these entries to your existing file, replace it entirely, or skip this file?"
   - For configuration files with placeholder values (like settings.json, config.toml, copilot-setup-steps.yml): "A version of this file already exists. Should I replace it with the template version, keep your existing version, or show a side-by-side so you can decide what to merge?"
   - For all other files: "This file already exists and differs from the template. Replace with template version, keep existing, or skip?"
   Wait for a response before moving to the next file.

After all template files are placed:
- Delete the zip archive and any temporary extraction directory.
- Commit all placed files: "Initialize project scaffolding"

If the ZIP does NOT contain template content (metaprompts-only variant), skip this step.

---

STEP 2b  -  INSTALL COPILOT PROMPT FILES

If the ZIP contains a prompts/ directory with .prompt.md files:

Copy all .prompt.md files from the prompts/ directory into the project's `.github/prompts/` directory (create it if it does not exist). VS Code discovers slash commands from `.github/prompts/` in each workspace.

List all installed prompt files and their slash command names (filename without .prompt.md extension).

State: "Copilot prompt files installed to .github/prompts/. Type / in Copilot chat to verify they appear."

Note: The Claude slash commands (.claude/commands/) are included in the template files and were already placed in Step 2 if template content was present.

---

STEP 3  -  CUSTOMIZE PROJECT CONVENTIONS

If template content was NOT placed (metaprompts-only variant), skip Steps 3 and 4 entirely  -  jump to Step 5.

Open `workflow/COMMANDS.md`. It contains `[PROJECT-SPECIFIC]` placeholder values for build commands and code conventions. These are initial values — Phase 4 (Scaffold Project) will refine them based on architecture reasoning from feature specs. Collect what the developer knows now:

**Project section:**
Ask: "What is the project name and a one-line description? What is the primary language or framework?"
(Store answers for use in Phase 2 Compass — do not fill AGENTS.md Overview yet; that is owned by the Compass phase.)

**Build section (→ workflow/COMMANDS.md → Core Commands):**
Ask: "What are your project's build steps? I need each of the following  -  provide the actual values or say 'not applicable' for any that don't apply:"
- Install
- Build
- Test (all)
- Test (single file or case)
- Lint
- Format
- Type-check

**Architecture section:**
Ask: "Describe the key directories in this project and what each one is responsible for. Aim for 5–15 lines mapping directories to responsibilities."

**Conventions section (→ workflow/COMMANDS.md → Code Conventions):**
Ask: "What naming conventions does this project use?"
- Functions and variables: (e.g., camelCase, snake_case)
- Files and directories: (e.g., kebab-case, PascalCase)

**Workflow/Governance routing check:**
Ask: "AGENTS.md routes to `/workflow/*.md` and `/governance/*.md`. Do you want to keep the default control-plane docs as-is, or should we tailor any sections now?"

After receiving answers for each section, update `workflow/COMMANDS.md` with the provided build and convention values. Present the completed file for review.

> **Note:** These are initial values. Phase 4 (Scaffold Project) refines commands and conventions based on architecture decisions derived from feature specs.

Ask: "Does this look correct? Anything to adjust?"
Iterate until confirmed.

Commit: "Customize AGENTS.md for this project"

---

STEP 4  -  CUSTOMIZE REMAINING CONFIGURATION

Walk through each of the following files, but only the ones that were placed (skip any that were not placed or were kept as existing versions during conflict resolution):

**.github/workflows/copilot-setup-steps.yml:**
This file now validates scaffold contracts and runs project commands via environment variables. Using the build steps from AGENTS.md, set:
- `INSTALL_CMD`
- `BUILD_CMD`
- `LINT_CMD`
- `TEST_CMD`
Present the updated file. Ask: "Does this match your CI environment? Adjust anything?"

**.claude/settings.json:**
Review the permissions list. Ask: "These tool permissions will be pre-approved for agent sessions. The current list covers common operations for git, npm, pip, python, node, and file management. Do these match your project's toolchain, or should any be added or removed?"
Also review the hooks section. Ask: "The post-edit hook runs a formatter on save. What is the format command for this project?" (Use the values from AGENTS.md if already provided.)
Ask: "The session-end Stop hook reminds you to run tests. The current reminder message is generic — would you like to customize it with your specific test command (e.g., `npm test`, `pytest`, `bun test`)?"
If yes, update the Stop hook command to: `echo 'Remember: run [their test command] before ending this session.'`
Update and present for confirmation.

**Optional agent permission mode (local-only):**
Ask exactly once: "Agent auto-approval (YOLO mode) can be configured for this project. The default is **option B  -  default interactive mode**. Choose a different option only if you intentionally want auto-approval:"

Offer these options:
- A) YOLO mode for both Copilot and Claude Code
- **B) Default interactive mode (keep prompts) (default)**
- C) Copilot YOLO only
- D) Claude Code YOLO only

If the user confirms or does not object, apply option B.

After applying the selected option:
- If option A, C, or D is selected, display this warning:
  > ⚠️ YOLO mode is enabled. The agent will execute file edits, terminal commands, and tool calls without asking for confirmation. You can switch to Plan Mode at any time for careful review. To disable, re-run initialization and select option B (Default interactive).
- If option B is selected, state:
  > Default interactive mode is active. The agent will continue asking for confirmation before privileged actions.

Apply by option:
- For Copilot YOLO (A or C): ensure `.vscode/settings.json` exists and merge `"chat.agent.autoApprove": true` while preserving all other keys.
- For Copilot default (B or D): if `.vscode/settings.json` contains `chat.agent.autoApprove`, remove only that key and preserve remaining settings.
- For Claude YOLO (A or D):
  1. Merge into `.vscode/settings.json`:
     - `"claudeCode.allowDangerouslySkipPermissions": true`
     - `"claudeCode.initialPermissionMode": "bypassPermissions"`
  2. Create or merge `.claude/settings.local.json` with:
     - `{ "permissions": { "defaultMode": "bypassPermissions" } }`
  3. Ensure `.claude/settings.local.json` is gitignored.
  4. Note: if CLI behavior differs from VS Code, the user may need to check their user-level `~/.claude/settings.json`.
- For Claude default (B or C): remove only these keys from `.vscode/settings.json` if present:
  - `claudeCode.allowDangerouslySkipPermissions`
  - `claudeCode.initialPermissionMode`
  And remove `.claude/settings.local.json` (or set its `permissions.defaultMode` to `default`).

Git hygiene for all options:
- Ensure `.gitignore` includes `.vscode/` and `.claude/settings.local.json`.
- Do not remove those ignore entries when reverting to default mode.

**.codex/config.toml:**
Ask: "Are you using Codex? If yes, review these settings  -  otherwise I'll leave the defaults and we can move on."
If yes, present the file and ask about model preference and sandbox configuration.
If no, move on.

Commit each file as it is confirmed: "Configure [filename] for this project"

---

STEP 5  -  VERIFY

Run checks appropriate to what was installed:

If template content was placed:
1. Confirm all expected directories exist: specs/, tasks/, decisions/, workflow/, governance/
2. Confirm AGENTS.md has no remaining placeholder brackets (no `[install step]`, `[project name]`, etc.)
3. Confirm .gitignore includes the scaffolding entries (CLAUDE.local.md, .claude/plans/, .trees/, .vscode/, .claude/settings.local.json)
4. Confirm all template files (specs/_TEMPLATE.md, tasks/_TEMPLATE.md, decisions/_TEMPLATE.md) are in place
5. Confirm workflow control-plane files are in place:
  - `workflow/LIFECYCLE.md`
  - `workflow/PLAYBOOK.md`
  - `workflow/FILE_CONTRACTS.md`
  - `workflow/STATE.json`
  - `workflow/FAILURE_ROUTING.md`
6. Confirm governance files are in place:
  - `governance/CHANGE_PROTOCOL.md`
  - `governance/POLICY_TESTS.md`
  - `governance/REGISTRY.md`
7. Confirm `.github/workflows/copilot-setup-steps.yml` has command env values populated for this project.
8. List any files that still contain placeholder values and ask: "These files still have placeholder values. Would you like to fill them in now, or leave them as templates?"
9. If optional agent permission mode was configured, confirm:
  - `.vscode/settings.json` is valid JSON and preserved existing keys
  - `.claude/settings.local.json` is valid JSON when present
  - Applied mode is accurately reported: Full YOLO / Copilot YOLO / Claude YOLO / Default interactive

If Copilot prompt files were installed:
10. List the installed .prompt.md files and confirm they are in the correct VS Code prompts directory.
11. List the available slash commands by name.

If Claude commands were placed (template content included .claude/commands/):
12. List the Claude slash commands available in .claude/commands/.

Report what was set up based on the detected variant:

For template-only:
  State: "Project scaffolding is in place. All conventions are configured and committed. Claude slash commands are available. The project is ready for development."

For metaprompts-only:
  State: "Copilot prompt files are installed. Type / in Copilot chat to access workflow commands. No template scaffolding was installed."

For full:
  State: "Project scaffolding is in place with all conventions configured. Claude slash commands and Copilot prompt files are both installed. The project is ready for development."

---

STEP 6  -  AUTO-INITIATE COMPASS (Phase 2)

If template content was placed (scaffold includes AGENTS.md and .specify/):

Check whether `.specify/constitution.md` exists and has its themes populated (no `[PROJECT-SPECIFIC]` placeholders remaining in covered sections).

If constitution is NOT populated:
  State: "Scaffolding complete. Now starting the Compass interview to establish your project's identity, goals, and boundaries."
  Auto-trigger: execute the Compass command (`.claude/commands/compass.md` or equivalent). This begins a dynamic discovery interview — not a scripted checklist. The compass will:
  1. Start broad — ask about the problem being solved, target audience, and what success looks like
  2. Follow the signal — pursue depth where answers reveal complexity or ambiguity
  3. Synthesize boundaries — establish what the project IS, is NOT, and what remains AMBIGUOUS
  4. Write outputs to `.specify/constitution.md` and `AGENTS.md → Overview`

  After Compass completes, state: "Constitution established. Run `/define-features` to translate it into a feature set, or run `/continue` to let the agent advance through remaining phases automatically."

If constitution IS already populated:
  State: "Constitution already exists. Run `/continue` to resume the agentic workflow from the current phase."
```
