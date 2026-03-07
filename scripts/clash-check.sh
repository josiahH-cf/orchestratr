#!/usr/bin/env bash
set -euo pipefail

# Clash Check — Pre-write conflict detection for parallel worktrees.
# Simulates three-way merges between all active worktrees to detect conflicts
# before they happen.
#
# Usage: scripts/clash-check.sh [--json]
# Exit code: 0 if no conflicts, 1 if conflicts detected.
#
# Uses `clash` CLI if available, otherwise falls back to git merge-tree simulation.

JSON_OUTPUT=false
if [[ "${1:-}" == "--json" ]]; then
  JSON_OUTPUT=true
fi

# Check if clash CLI is available
if command -v clash &>/dev/null; then
  if $JSON_OUTPUT; then
    clash status --json
  else
    clash status
  fi
  exit $?
fi

# Fallback: git merge-tree simulation
BASE_BRANCH=$(git rev-parse --abbrev-ref HEAD)
BASE_COMMIT=$(git rev-parse HEAD)

conflicts_found=0
conflict_details=()

if [[ ! -d ".trees" ]]; then
  echo "No worktrees found. No conflicts possible."
  exit 0
fi

# Collect all worktree branches
branches=()
for dir in .trees/*/; do
  [[ -d "$dir" ]] || continue
  branch=$(git -C "$dir" rev-parse --abbrev-ref HEAD 2>/dev/null || continue)
  branches+=("$branch")
done

if (( ${#branches[@]} < 2 )); then
  echo "Fewer than 2 active worktrees. No cross-worktree conflicts possible."
  exit 0
fi

# Pairwise conflict check using git merge-tree
for (( i=0; i<${#branches[@]}; i++ )); do
  for (( j=i+1; j<${#branches[@]}; j++ )); do
    branch_a="${branches[$i]}"
    branch_b="${branches[$j]}"

    # Find merge base
    merge_base=$(git merge-base "$branch_a" "$branch_b" 2>/dev/null || echo "$BASE_COMMIT")

    # Simulate merge
    merge_result=$(git merge-tree "$merge_base" "$branch_a" "$branch_b" 2>/dev/null || true)

    if echo "$merge_result" | grep -q "^<<<<<<<"; then
      conflicts_found=1
      conflict_details+=("CONFLICT: $branch_a ↔ $branch_b")
      if ! $JSON_OUTPUT; then
        echo "⚠ CONFLICT detected: $branch_a ↔ $branch_b"
        echo "$merge_result" | grep -A2 "^<<<<<<<"  | head -10
        echo ""
      fi
    fi
  done
done

if $JSON_OUTPUT; then
  echo "{"
  echo "  \"conflicts_found\": $conflicts_found,"
  echo "  \"details\": ["
  for (( k=0; k<${#conflict_details[@]}; k++ )); do
    comma=""
    if (( k < ${#conflict_details[@]} - 1 )); then comma=","; fi
    echo "    \"${conflict_details[$k]}\"$comma"
  done
  echo "  ]"
  echo "}"
fi

if (( conflicts_found )); then
  if ! $JSON_OUTPUT; then
    echo "Conflicts detected. Consider:"
    echo "  1. Rebase one branch onto the other"
    echo "  2. Reassign overlapping tasks to a single agent"
    echo "  3. Define interface contracts to decouple the work"
  fi
  exit 1
else
  if ! $JSON_OUTPUT; then
    echo "No conflicts detected across ${#branches[@]} worktrees."
  fi
  exit 0
fi
