#!/bin/bash
set -e

# Gather git information for PR template generation.
# Usage: gather-git-info.sh [feature-branch] [base-branch]
#   feature-branch — branch with changes (default: current branch)
#   base-branch    — branch to compare against (default: auto-detect main/master/develop)
# Outputs JSON with branch, commits, diff stats, and full diff.

REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"

FEATURE_BRANCH="${1:-}"
BASE_BRANCH="${2:-}"

# Auto-detect base branch
if [ -z "$BASE_BRANCH" ]; then
  for candidate in main master develop; do
    if git rev-parse --verify "$candidate" >/dev/null 2>&1; then
      BASE_BRANCH="$candidate"
      break
    fi
  done
  if [ -z "$BASE_BRANCH" ]; then
    echo "Error: could not detect base branch (tried main, master, develop)." >&2
    exit 1
  fi
fi

# Resolve feature branch
if [ -n "$FEATURE_BRANCH" ]; then
  if ! git rev-parse --verify "$FEATURE_BRANCH" >/dev/null 2>&1; then
    echo "Error: branch '$FEATURE_BRANCH' does not exist." >&2
    exit 1
  fi
  RESOLVED_BRANCH="$FEATURE_BRANCH"
else
  RESOLVED_BRANCH=$(git rev-parse --abbrev-ref HEAD)
fi

if [ "$RESOLVED_BRANCH" = "$BASE_BRANCH" ]; then
  echo "Error: feature branch is the same as base branch ($BASE_BRANCH)." >&2
  exit 1
fi

MERGE_BASE=$(git merge-base "$BASE_BRANCH" "$RESOLVED_BRANCH" 2>/dev/null || echo "")
if [ -z "$MERGE_BASE" ]; then
  echo "Error: no common ancestor between $RESOLVED_BRANCH and $BASE_BRANCH." >&2
  exit 1
fi

REF="$RESOLVED_BRANCH"

# Collect commits as proper JSON via jq
COMMITS=$(git log --reverse --format='%H|||%s' "$MERGE_BASE".."$REF" | \
  jq -R -s -c '[split("\n") | .[] | select(. != "") | split("|||") | {hash: .[0], subject: .[1]}]')

# Diff stats
DIFF_STAT=$(git diff --stat "$MERGE_BASE".."$REF")

# Full diff (capped at 50000 chars to avoid excessive output)
FULL_DIFF=$(git diff "$MERGE_BASE".."$REF" | head -c 50000)

# Files changed
FILES_CHANGED=$(git diff --name-only "$MERGE_BASE".."$REF" | sort)

# Detect GitHub PR template
PR_TEMPLATE_PATH=""
PR_TEMPLATE_CONTENT=""
TEMPLATE_CANDIDATES=(
  "$REPO_ROOT/.github/PULL_REQUEST_TEMPLATE.md"
  "$REPO_ROOT/.github/pull_request_template.md"
  "$REPO_ROOT/PULL_REQUEST_TEMPLATE.md"
  "$REPO_ROOT/pull_request_template.md"
  "$REPO_ROOT/docs/pull_request_template.md"
  "$REPO_ROOT/docs/PULL_REQUEST_TEMPLATE.md"
)

for candidate in "${TEMPLATE_CANDIDATES[@]}"; do
  if [ -f "$candidate" ]; then
    PR_TEMPLATE_PATH="${candidate#$REPO_ROOT/}"
    PR_TEMPLATE_CONTENT=$(cat "$candidate")
    break
  fi
done

# If no single file found, check for template directory (use default.md or first file)
if [ -z "$PR_TEMPLATE_PATH" ]; then
  TEMPLATE_DIR="$REPO_ROOT/.github/PULL_REQUEST_TEMPLATE"
  if [ -d "$TEMPLATE_DIR" ]; then
    if [ -f "$TEMPLATE_DIR/default.md" ]; then
      PR_TEMPLATE_PATH=".github/PULL_REQUEST_TEMPLATE/default.md"
      PR_TEMPLATE_CONTENT=$(cat "$TEMPLATE_DIR/default.md")
    else
      FIRST_TEMPLATE=$(find "$TEMPLATE_DIR" -name '*.md' -type f 2>/dev/null | sort | head -1)
      if [ -n "$FIRST_TEMPLATE" ]; then
        PR_TEMPLATE_PATH="${FIRST_TEMPLATE#$REPO_ROOT/}"
        PR_TEMPLATE_CONTENT=$(cat "$FIRST_TEMPLATE")
      fi
    fi
  fi
fi

# Output JSON
cat <<ENDJSON
{
  "branch": "$RESOLVED_BRANCH",
  "base_branch": "$BASE_BRANCH",
  "merge_base": "$MERGE_BASE",
  "commits": $COMMITS,
  "files_changed": $(echo "$FILES_CHANGED" | jq -R -s 'split("\n") | map(select(. != ""))'),
  "diff_stat": $(echo "$DIFF_STAT" | jq -R -s .),
  "diff": $(echo "$FULL_DIFF" | jq -R -s .),
  "pr_template_path": $(echo "$PR_TEMPLATE_PATH" | jq -R -s .),
  "pr_template": $(echo "$PR_TEMPLATE_CONTENT" | jq -R -s .)
}
ENDJSON
