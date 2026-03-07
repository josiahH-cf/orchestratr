#!/usr/bin/env bash
set -euo pipefail

# workflow-lint.sh — Non-destructive linting for workflow artifacts.
# Produces a Markdown report at workflow/LINT_REPORT.md.
# Always exits 0 (advisory only). See workflow/LINT_CONTRACT.md for spec.

PROJECT_ROOT="${1:-.}"
cd "$PROJECT_ROOT"

REPORT="workflow/LINT_REPORT.md"
TIMESTAMP="$(date -u '+%Y-%m-%dT%H:%M:%SZ')"

# Counters
ORPHAN_COUNT=0
LENGTH_COUNT=0
EOF_COUNT=0
CLARITY_COUNT=0
FILES_SCANNED=0

# Accumulate findings
ORPHAN_FINDINGS=""
LENGTH_FINDINGS=""
EOF_FINDINGS=""
CLARITY_FINDINGS=""

# ─── Helpers ───

add_finding() {
  local category="$1" file="$2" line="$3" message="$4"
  local entry="| \`$file\` | $line | $message |"
  case "$category" in
    orphan)  ORPHAN_FINDINGS="${ORPHAN_FINDINGS}${entry}"$'\n'; ORPHAN_COUNT=$((ORPHAN_COUNT + 1)) ;;
    length)  LENGTH_FINDINGS="${LENGTH_FINDINGS}${entry}"$'\n'; LENGTH_COUNT=$((LENGTH_COUNT + 1)) ;;
    eof)     EOF_FINDINGS="${EOF_FINDINGS}${entry}"$'\n'; EOF_COUNT=$((EOF_COUNT + 1)) ;;
    clarity) CLARITY_FINDINGS="${CLARITY_FINDINGS}${entry}"$'\n'; CLARITY_COUNT=$((CLARITY_COUNT + 1)) ;;
  esac
}

# Collect all tracked text files
collect_files() {
  find . \( -name '*.md' -o -name '*.json' -o -name '*.sh' -o -name '*.yml' -o -name '*.yaml' -o -name '*.toml' \) \
    -not -path './.git/*' \
    -not -path './node_modules/*' \
    -not -path './.trees/*' \
    -not -name 'LINT_REPORT.md' \
    | sort
}

# Collect markdown files in lint-target directories
collect_lint_md() {
  local dirs=("specs" "tasks" "decisions" "workflow" "governance" ".specify")
  for d in "${dirs[@]}"; do
    [[ -d "$d" ]] && find "$d" -name '*.md' -not -name '_TEMPLATE.md' | sort
  done
}

# Build a reference corpus: all text from tracked markdown files + STATE.json
build_reference_corpus() {
  local corpus=""
  while IFS= read -r f; do
    [[ -f "$f" ]] && corpus+="$(cat "$f")"$'\n'
  done < <(find . -name '*.md' -not -path './.git/*' -not -path './node_modules/*' -not -path './.trees/*' 2>/dev/null | sort)
  # Include STATE.json task references
  [[ -f "workflow/STATE.json" ]] && corpus+="$(cat "workflow/STATE.json")"$'\n'
  printf '%s' "$corpus"
}

# ─── Check 1: Orphan Detection ───

run_orphan_check() {
  local corpus
  corpus="$(build_reference_corpus)"

  local dirs=("specs" "tasks" "decisions")
  for d in "${dirs[@]}"; do
    [[ -d "$d" ]] || continue
    while IFS= read -r f; do
      local basename
      basename="$(basename "$f")"
      [[ "$basename" == "_TEMPLATE.md" ]] && continue
      # Check if this file is referenced anywhere in the corpus (by filename or relative path)
      if ! echo "$corpus" | grep -qF "$basename" && ! echo "$corpus" | grep -qF "$f"; then
        add_finding "orphan" "$f" "—" "No inbound reference found"
      fi
    done < <(find "$d" -name '*.md' | sort)
  done
}

# ─── Check 2: Length Warning ───

run_length_check() {
  while IFS= read -r f; do
    local lines
    lines="$(wc -l < "$f")"
    if [[ "$lines" -gt 120 ]]; then
      add_finding "length" "$f" "—" "**$lines** lines (threshold: 120)"
    fi
  done < <(collect_lint_md)
}

# ─── Check 3: EOF Newline ───

run_eof_check() {
  while IFS= read -r f; do
    [[ -s "$f" ]] || continue
    FILES_SCANNED=$((FILES_SCANNED + 1))

    # Check last byte
    local last_byte
    last_byte="$(tail -c 1 "$f" | xxd -p)"
    if [[ "$last_byte" != "0a" && "$last_byte" != "" ]]; then
      add_finding "eof" "$f" "EOF" "Missing trailing newline"
      continue
    fi

    # Check for multiple trailing newlines
    local last_two
    last_two="$(tail -c 2 "$f" | xxd -p)"
    if [[ "$last_two" == "0a0a" ]]; then
      add_finding "eof" "$f" "EOF" "Multiple trailing newlines"
    fi
  done < <(collect_files)
}

# ─── Check 4: Clarity Heuristics ───

run_clarity_check() {
  while IFS= read -r f; do
    [[ "$f" == *.md ]] || continue

    # 4a. Missing H1 in first 5 lines
    if ! head -5 "$f" | grep -qE '^# '; then
      add_finding "clarity" "$f" "1–5" "No H1 heading in first 5 lines"
    fi

    # 4b. Missing required sections for specs/tasks
    local basename
    basename="$(basename "$f")"
    if [[ "$f" == specs/* && "$basename" != "_TEMPLATE.md" ]]; then
      if ! grep -qE '^## Acceptance Criteria' "$f"; then
        add_finding "clarity" "$f" "—" "Spec missing \`## Acceptance Criteria\` section"
      fi
    fi
    if [[ "$f" == tasks/* && "$basename" != "_TEMPLATE.md" ]]; then
      if ! grep -qE '^## Tasks' "$f"; then
        add_finding "clarity" "$f" "—" "Task missing \`## Tasks\` section"
      fi
    fi

    # 4c. TODO/FIXME markers (skip table rows that document the pattern)
    local todo_lines
    todo_lines="$(grep -nE '\bTODO\b|\bFIXME\b' "$f" 2>/dev/null | grep -vE '^[0-9]+:\|' || true)"
    if [[ -n "$todo_lines" ]]; then
      while IFS= read -r match; do
        local lineno msg
        lineno="$(echo "$match" | cut -d: -f1)"
        msg="$(echo "$match" | cut -d: -f2- | sed 's/|/\\|/g' | head -c 80)"
        add_finding "clarity" "$f" "$lineno" "TODO/FIXME: ${msg}"
      done <<< "$todo_lines"
    fi

    # 4d. Sentence length outlier (non-code lines > 200 chars)
    local lineno=0
    while IFS= read -r line; do
      lineno=$((lineno + 1))
      # Skip code blocks
      [[ "$line" =~ ^\`\`\` ]] && continue
      # Skip table rows
      [[ "$line" =~ ^\| ]] && continue
      # Skip links-heavy lines (URLs inflate length)
      [[ "$line" =~ https?:// ]] && continue
      if [[ "${#line}" -gt 200 ]]; then
        add_finding "clarity" "$f" "$lineno" "Line exceeds 200 characters (${#line} chars)"
      fi
    done < "$f"

    # 4e. Passive voice density
    local total_sentences=0
    local passive_sentences=0
    while IFS= read -r line; do
      [[ "$line" =~ ^\`\`\` ]] && continue
      [[ "$line" =~ ^\| ]] && continue
      [[ -z "$line" ]] && continue
      [[ "$line" =~ ^#  ]] && continue
      [[ "$line" =~ ^-\  ]] && continue
      total_sentences=$((total_sentences + 1))
      if echo "$line" | grep -qiE '\bis being\b|\bwas\b|\bwere\b|\bbeen\b|\bbe done\b|\bis performed\b|\bis executed\b'; then
        passive_sentences=$((passive_sentences + 1))
      fi
    done < "$f"
    if [[ "$total_sentences" -gt 5 ]]; then
      local pct=$((passive_sentences * 100 / total_sentences))
      if [[ "$pct" -gt 40 ]]; then
        add_finding "clarity" "$f" "—" "Passive voice density: ${pct}% ($passive_sentences/$total_sentences lines)"
      fi
    fi

  done < <(collect_lint_md)
}

# ─── Run All Checks ───

run_orphan_check
run_length_check
run_eof_check
run_clarity_check

TOTAL=$((ORPHAN_COUNT + LENGTH_COUNT + EOF_COUNT + CLARITY_COUNT))

# ─── Generate Report ───

mkdir -p "$(dirname "$REPORT")"

cat > "$REPORT" <<REPORT_HEADER
# Workflow Lint Report

- **Generated:** $TIMESTAMP
- **Project root:** \`$(pwd)\`
- **Files scanned:** $FILES_SCANNED
- **Total findings:** $TOTAL

## Summary

| Category | Count |
|----------|-------|
| Orphan | $ORPHAN_COUNT |
| Length | $LENGTH_COUNT |
| EOF Newline | $EOF_COUNT |
| Clarity | $CLARITY_COUNT |

REPORT_HEADER

if [[ "$ORPHAN_COUNT" -gt 0 ]]; then
  cat >> "$REPORT" <<'SECTION'
## Orphan Detection

| File | Line | Message |
|------|------|---------|
SECTION
  printf '%s' "$ORPHAN_FINDINGS" >> "$REPORT"
  echo "" >> "$REPORT"
fi

if [[ "$LENGTH_COUNT" -gt 0 ]]; then
  cat >> "$REPORT" <<'SECTION'
## Length Warnings

| File | Line | Message |
|------|------|---------|
SECTION
  printf '%s' "$LENGTH_FINDINGS" >> "$REPORT"
  echo "" >> "$REPORT"
fi

if [[ "$EOF_COUNT" -gt 0 ]]; then
  cat >> "$REPORT" <<'SECTION'
## EOF Newline

| File | Line | Message |
|------|------|---------|
SECTION
  printf '%s' "$EOF_FINDINGS" >> "$REPORT"
  echo "" >> "$REPORT"
fi

if [[ "$CLARITY_COUNT" -gt 0 ]]; then
  cat >> "$REPORT" <<'SECTION'
## Clarity Heuristics

| File | Line | Message |
|------|------|---------|
SECTION
  printf '%s' "$CLARITY_FINDINGS" >> "$REPORT"
  echo "" >> "$REPORT"
fi

if [[ "$TOTAL" -eq 0 ]]; then
  echo "**No findings. All checks passed.**" >> "$REPORT"
fi

echo "---" >> "$REPORT"
echo "_Generated by \`scripts/workflow-lint.sh\`. See \`workflow/LINT_CONTRACT.md\` for check definitions._" >> "$REPORT"

# ─── Console Summary ───

echo "Workflow lint complete: $TOTAL findings ($ORPHAN_COUNT orphan, $LENGTH_COUNT length, $EOF_COUNT eof, $CLARITY_COUNT clarity)"
echo "Report written to: $REPORT"

exit 0
