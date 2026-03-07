#!/usr/bin/env bash
set -euo pipefail

# Setup a git worktree for parallel agent work.
# Usage: scripts/setup-worktree.sh <agent> <type> <description>
# Example: scripts/setup-worktree.sh claude feat auth-flow
#          scripts/setup-worktree.sh agent-1 feat auth-flow
#
# Creates: .trees/<agent>-<type>-<description>/
# Branch: <agent>/<type>-<description>
# Agent: any identifier (claude, copilot, codex, agent-1, user name, etc.)

# --list: inventory of all active worktrees
if [[ "${1:-}" == "--list" ]]; then
  echo "Active Worktrees:"
  if [[ -d ".trees" ]]; then
    for dir in .trees/*/; do
      [[ -d "$dir" ]] || continue
      name=$(basename "$dir")
      branch=$(git -C "$dir" rev-parse --abbrev-ref HEAD 2>/dev/null || echo "detached")
      # Age calculation from directory creation time
      age_seconds=$(( $(date +%s) - $(stat -c %Y "$dir" 2>/dev/null || stat -f %m "$dir" 2>/dev/null) ))
      if (( age_seconds < 3600 )); then
        age="$((age_seconds / 60))m"
      else
        age="$((age_seconds / 3600))h"
      fi
      # Modified files count
      modified=$(git -C "$dir" status --porcelain 2>/dev/null | wc -l | tr -d ' ')
      if [[ "$modified" == "0" ]]; then
        status="clean"
      else
        status="$modified modified files"
      fi
      printf "  %-30s branch: %-30s age: %-6s status: %s\n" "$name" "$branch" "$age" "$status"
    done
  else
    echo "  (no worktrees)"
  fi
  exit 0
fi

# --cleanup: remove worktrees whose branches have been merged
if [[ "${1:-}" == "--cleanup" ]]; then
  if [[ ! -d ".trees" ]]; then
    echo "No worktrees to clean up."
    exit 0
  fi
  base_branch=$(git rev-parse --abbrev-ref HEAD)
  merged_branches=$(git branch --merged "$base_branch" | sed 's/^[ *]*//')
  cleaned=0
  for dir in .trees/*/; do
    [[ -d "$dir" ]] || continue
    branch=$(git -C "$dir" rev-parse --abbrev-ref HEAD 2>/dev/null || continue)
    if echo "$merged_branches" | grep -qx "$branch"; then
      echo "Removing merged worktree: $(basename "$dir") (branch: $branch)"
      git worktree remove "$dir" --force
      ((cleaned++))
    fi
  done
  echo "Cleaned up $cleaned worktree(s)."
  exit 0
fi

if [[ $# -lt 3 ]]; then
  echo "Usage: $0 <agent> <type> <description>"
  echo "  agent:       any identifier (e.g., claude, copilot, codex, agent-1, your name)"
  echo "  type:        feat | bug | refactor | chore | docs"
  echo "  description: 2-4 word kebab-case summary"
  echo ""
  echo "Flags:"
  echo "  --list      Show active worktrees"
  echo "  --cleanup   Remove worktrees with merged branches"
  echo ""
  echo "Example: $0 claude feat auth-flow"
  echo "         $0 agent-1 feat auth-flow"
  exit 1
fi

AGENT="$1"
TYPE="$2"
DESC="$3"

# Validate agent identifier (alphanumeric, hyphens, underscores)
if [[ ! "$AGENT" =~ ^[a-zA-Z0-9_-]+$ ]]; then
  echo "Error: agent must contain only alphanumeric characters, hyphens, or underscores (got: $AGENT)"
  exit 1
fi

# Validate type
if [[ ! "$TYPE" =~ ^(feat|bug|refactor|chore|docs)$ ]]; then
  echo "Error: type must be feat, bug, refactor, chore, or docs (got: $TYPE)"
  exit 1
fi

BRANCH="${AGENT}/${TYPE}-${DESC}"
WORKTREE_DIR=".trees/${AGENT}-${TYPE}-${DESC}"

# Check if worktree already exists
if [[ -d "$WORKTREE_DIR" ]]; then
  echo "Error: worktree already exists at $WORKTREE_DIR"
  exit 1
fi

# Create the worktree directory parent
mkdir -p .trees

# Check disk space (warn if < 2GB available)
available_kb=$(df -k . | awk 'NR==2 {print $4}')
if (( available_kb < 2097152 )); then
  echo "Warning: Less than 2GB disk space available ($(( available_kb / 1024 ))MB)."
  echo "Worktrees consume significant space. Proceed? (y/N)"
  read -r confirm
  [[ "$confirm" =~ ^[Yy]$ ]] || exit 1
fi

# Check for file overlap with existing worktrees
if [[ -d ".trees" ]]; then
  for existing_dir in .trees/*/; do
    [[ -d "$existing_dir" ]] || continue
    existing_modified=$(git -C "$existing_dir" diff --name-only 2>/dev/null)
    if [[ -n "$existing_modified" ]]; then
      echo "Note: Worktree $(basename "$existing_dir") has modified files:"
      echo "$existing_modified" | head -5
      echo "Ensure your new task doesn't overlap these files."
    fi
  done
fi

# Get the current branch as base
BASE_BRANCH=$(git rev-parse --abbrev-ref HEAD)

# Create worktree with new branch
echo "Creating worktree: $WORKTREE_DIR"
echo "Branch: $BRANCH (from $BASE_BRANCH)"
git worktree add -b "$BRANCH" "$WORKTREE_DIR" "$BASE_BRANCH"

# Copy local config files that aren't tracked
for config_file in .claude/settings.local.json .vscode/settings.json; do
  if [[ -f "$config_file" ]]; then
    target_dir="$WORKTREE_DIR/$(dirname "$config_file")"
    mkdir -p "$target_dir"
    cp "$config_file" "$WORKTREE_DIR/$config_file"
    echo "Copied: $config_file"
  fi
done

echo ""
echo "Worktree ready:"
echo "  Directory: $WORKTREE_DIR"
echo "  Branch:    $BRANCH"
echo "  Base:      $BASE_BRANCH"
echo ""
echo "To work in this worktree: cd $WORKTREE_DIR"
echo "To remove when done: git worktree remove $WORKTREE_DIR"
