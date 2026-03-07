#!/usr/bin/env bash
set -euo pipefail

fail() {
  echo "[FAIL] $1" >&2
  exit 1
}

note() {
  echo "[INFO] $1"
}

require_file() {
  local path="$1"
  [[ -f "$path" ]] || fail "Missing required file: $path"
}

require_file "AGENTS.md"
require_file "workflow/LIFECYCLE.md"
require_file "workflow/PLAYBOOK.md"
require_file "workflow/FILE_CONTRACTS.md"
require_file "workflow/STATE.json"
require_file "workflow/FAILURE_ROUTING.md"
require_file "workflow/ROUTING.md"
require_file "workflow/COMMANDS.md"
require_file "workflow/BOUNDARIES.md"
require_file "workflow/FILE_CONTRACTS.md"
require_file "workflow/ORCHESTRATOR.md"
require_file "workflow/CONCURRENCY.md"
require_file "governance/CHANGE_PROTOCOL.md"
require_file "governance/POLICY_TESTS.md"
require_file "governance/REGISTRY.md"
require_file ".specify/constitution.md"
require_file ".specify/spec-template.md"
require_file ".specify/acceptance-criteria-template.md"
require_file ".github/workflows/copilot-setup-steps.yml"

# Validate that AGENTS.md links resolve to existing files
ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
while IFS= read -r link; do
    target=$(echo "$link" | sed -n 's/.*](\([^)]*\)).*/\1/p')
    if [[ -n "$target" && ! -f "$target" && ! -f "$ROOT/$target" ]]; then
        fail "AGENTS.md contains broken link: $target"
    fi
done < <(grep -oE '\[[^]]*\]\([^)]*\)' "$ROOT/AGENTS.md" 2>/dev/null || true)

# Ensure policy mapping document still declares checks this script enforces.
for requirement in \
  "Every spec has matching task file" \
  "Continue state is valid" \
  "Build, lint, test commands are defined"; do
  grep -Fq "$requirement" governance/POLICY_TESTS.md || fail "POLICY_TESTS missing mapping: $requirement"
done

# Validate state JSON, phase-aware placeholder rules, and currentTaskFile reference.
python3 - <<'PY'
import json
from pathlib import Path

state_path = Path('workflow/STATE.json')
state = json.loads(state_path.read_text())

required = {"projectPhase", "currentFeatureId", "currentTaskFile", "testMode", "updatedAt"}
missing = required.difference(state)
if missing:
    raise SystemExit(f"workflow/STATE.json missing keys: {sorted(missing)}")

valid_phases = [
    "2-compass",
    "3-define-features",
    "4-scaffold-project",
    "5-fine-tune-plan",
    "6-code",
    "7-test",
    "7b-review-ship",
    "8-maintain",
    "done",
]
phase = state.get("projectPhase", "")
if phase not in valid_phases:
    raise SystemExit(f"workflow/STATE.json has invalid projectPhase: {phase}")

phase_idx = valid_phases.index(phase)

# Constitution placeholders are allowed only before Define Features.
constitution = Path(".specify/constitution.md").read_text()
if phase_idx >= valid_phases.index("3-define-features") and "[PROJECT-SPECIFIC]" in constitution:
    raise SystemExit("constitution.md still has [PROJECT-SPECIFIC] placeholders for phase >= 3-define-features")

# AGENTS placeholders are allowed until Fine-tune Plan starts.
agents = Path("AGENTS.md").read_text()
if phase_idx >= valid_phases.index("5-fine-tune-plan") and "[PROJECT-SPECIFIC]" in agents:
    raise SystemExit("AGENTS.md still has [PROJECT-SPECIFIC] placeholders for phase >= 5-fine-tune-plan")

# CI command defaults must be replaced by fine-tune/code phases.
workflow = Path(".github/workflows/copilot-setup-steps.yml").read_text()
if phase_idx >= valid_phases.index("5-fine-tune-plan"):
    defaults = [
        "Set INSTALL_CMD in workflow env",
        "Set BUILD_CMD in workflow env",
        "Set LINT_CMD in workflow env",
        "Set TEST_CMD in workflow env",
    ]
    for marker in defaults:
        if marker in workflow:
            raise SystemExit(f"copilot-setup-steps.yml still has default command marker: {marker}")

task_file = (state.get("currentTaskFile") or "").strip()
if task_file:
    p = Path(task_file.lstrip('/'))
    if not p.exists():
        raise SystemExit(f"workflow/STATE.json currentTaskFile does not exist: {task_file}")

# Validate advisoryProfile if present
advisory = state.get("advisoryProfile", "")
if advisory and advisory not in ("concise", "standard", "detailed"):
    raise SystemExit(f"workflow/STATE.json has invalid advisoryProfile: {advisory}")
PY

# Spec/task parity checks.
shopt -s nullglob
specs=(specs/*.md)
tasks=(tasks/*.md)

for spec in "${specs[@]}"; do
  base="$(basename "$spec")"
  [[ "$base" == "_TEMPLATE.md" ]] && continue
  task="tasks/$base"
  [[ -f "$task" ]] || fail "Spec has no matching task file: $spec -> $task"

done

for task in "${tasks[@]}"; do
  base="$(basename "$task")"
  [[ "$base" == "_TEMPLATE.md" ]] && continue
  grep -q 'AC-' "$task" || fail "Task file has no AC mapping: $task"
done

note "Policy checks passed"
