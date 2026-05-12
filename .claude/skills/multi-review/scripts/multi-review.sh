#!/usr/bin/env bash
#
# multi-review.sh — workhorse for the multi-review skill.
#
# Spawns 4 Claude Code reviewers in parallel (2× /review when a PR exists,
# 2× /code-review always), then runs a merger that synthesizes them into
# both a human-readable markdown review AND a machine-readable JSON file
# that the skill's walkthrough consumes.
#
# Usage:
#   bash multi-review.sh <run_dir>
#
# Args:
#   <run_dir>     output directory relative to repo root (required)
#                 typical: reviews/pr-272-fix-uint64-20260512-094530
#
# Env:
#   BASE_BRANCH   base branch (default: main)
#   PR_NUMBER     PR number — enables /review agents; required for reviewer
#                 mode but optional for developer mode
#   GH_REPO       repo for /review (e.g. XRPLF/xrpl-go) — should be derived
#                 from `git remote get-url origin`, never trust gh default
#   CLAUDE_BIN    path to claude CLI (default: claude)
#   MODEL         model for every agent (default: claude-opus-4-7)
#
# Output (under <run_dir>/):
#   review-skill-{1,2}.md    /review agents (only when PR_NUMBER is set)
#   code-review-{1,2}.md     /code-review agents
#   merged-review.md         human-readable synthesis
#   merged-review.json       machine-readable findings (skill consumes this)
#   <name>.log               stdout+stderr per agent + merger.log
#
# Cost: 4 concurrent Opus 4.7 sessions + 1 merger ≈ 5× a single review.

set -euo pipefail

if [[ $# -lt 1 || -z "${1:-}" ]]; then
    echo "ERROR: missing required <run_dir> argument" >&2
    echo "Usage: bash multi-review.sh <run_dir>" >&2
    exit 1
fi

RUN_DIR="$1"
BASE_BRANCH="${BASE_BRANCH:-main}"
CLAUDE_BIN="${CLAUDE_BIN:-claude}"
PR_NUMBER="${PR_NUMBER:-}"
GH_REPO="${GH_REPO:-}"
MODEL="${MODEL:-claude-opus-4-7}"

REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || true)"
if [[ -z "$REPO_ROOT" ]]; then
    echo "ERROR: not inside a git repo" >&2
    exit 1
fi

BRANCH="$(git -C "$REPO_ROOT" rev-parse --abbrev-ref HEAD)"
if [[ "$BRANCH" == "$BASE_BRANCH" ]]; then
    echo "ERROR: current branch ($BRANCH) is the base branch; nothing to review" >&2
    exit 1
fi

if ! command -v "$CLAUDE_BIN" >/dev/null 2>&1; then
    echo "ERROR: '$CLAUDE_BIN' not found on PATH" >&2
    exit 1
fi

mkdir -p "$REPO_ROOT/$RUN_DIR"
ABS_OUT="$REPO_ROOT/$RUN_DIR"

echo "Repo:     $REPO_ROOT"
echo "Branch:   $BRANCH"
echo "Base:     $BASE_BRANCH"
echo "PR:       ${PR_NUMBER:-<none — /review agents will be skipped>}"
[[ -n "$GH_REPO" ]] && echo "GH repo:  $GH_REPO"
echo "Model:    $MODEL"
echo "Output:   $ABS_OUT"
echo

# ----- Helpers --------------------------------------------------------------

# Build the /review invocation. If GH_REPO is set, tell the agent to use it
# explicitly so it doesn't pick up the wrong default origin.
build_review_prompt() {
    local pr="$1"
    if [[ -n "$GH_REPO" ]]; then
        cat <<EOF
Use the /review skill on PR #$pr in the GitHub repo $GH_REPO. When running gh
commands, always pass --repo $GH_REPO. Produce the review as your final
assistant message in markdown.
EOF
    else
        echo "/review $pr"
    fi
}

run_agent() {
    local name="$1" prompt="$2"
    cd "$REPO_ROOT"
    "$CLAUDE_BIN" -p "$prompt" \
        --model "$MODEL" \
        --permission-mode acceptEdits \
        --output-format text \
        > "$ABS_OUT/$name.md" 2> "$ABS_OUT/$name.log"
}

# ----- Spawn agents ---------------------------------------------------------

declare -a PIDS NAMES

if [[ -n "$PR_NUMBER" ]]; then
    REVIEW_PROMPT="$(build_review_prompt "$PR_NUMBER")"
    echo "Spawning 2× /review agents on PR #$PR_NUMBER..."
    run_agent "review-skill-1" "$REVIEW_PROMPT" & PIDS+=($!); NAMES+=("review-skill-1")
    run_agent "review-skill-2" "$REVIEW_PROMPT" & PIDS+=($!); NAMES+=("review-skill-2")
else
    echo "Skipping /review agents (no PR number)."
fi

echo "Spawning 2× /code-review agents..."
run_agent "code-review-1" "/code-review" & PIDS+=($!); NAMES+=("code-review-1")
run_agent "code-review-2" "/code-review" & PIDS+=($!); NAMES+=("code-review-2")

echo "  PIDs: ${PIDS[*]}"
echo

# ----- Wait -----------------------------------------------------------------

FAILED=0
for i in "${!PIDS[@]}"; do
    name="${NAMES[$i]}"
    if wait "${PIDS[$i]}"; then
        if [[ -s "$ABS_OUT/$name.md" ]]; then
            echo "  $name: ✓ $ABS_OUT/$name.md"
        else
            echo "  $name: ✗ exited 0 but produced empty output (see $name.log)" >&2
            FAILED=1
        fi
    else
        echo "  $name: ✗ failed (see $name.log)" >&2
        FAILED=1
    fi
done

if [[ $FAILED -eq 1 ]]; then
    echo
    echo "One or more agents failed. Inspect logs in $ABS_OUT/." >&2
    exit 1
fi

# Collect produced review files for the merger.
REVIEW_FILES=()
for f in "$ABS_OUT"/review-skill-1.md \
         "$ABS_OUT"/review-skill-2.md \
         "$ABS_OUT"/code-review-1.md \
         "$ABS_OUT"/code-review-2.md; do
    [[ -s "$f" ]] && REVIEW_FILES+=("$f")
done

if [[ ${#REVIEW_FILES[@]} -lt 2 ]]; then
    echo "ERROR: fewer than 2 non-empty review files; nothing to merge." >&2
    exit 1
fi

# ----- Merger ---------------------------------------------------------------

echo
echo "Merging ${#REVIEW_FILES[@]} reviews..."

# Build a bullet list of source files for the merger.
SOURCE_LIST=""
for f in "${REVIEW_FILES[@]}"; do
    SOURCE_LIST+="- $f"$'\n'
done

read -r -d '' MERGE_PROMPT <<EOF || true
Multiple independent code reviews of the current branch's diff vs $BASE_BRANCH
have been written. Up to two were produced by the /review skill, up to two by
the /code-review skill. Read all of them:

$SOURCE_LIST

You must produce TWO files in $ABS_OUT/:

============================================================================
FILE 1: $ABS_OUT/merged-review.md  (human-readable)
============================================================================

A single synthesized review with:

1. Deduplicated findings across the source reviews.
2. A "Reviewer agreement" count per finding (how many of the source reviews
   surfaced it — a proxy for signal-to-noise). Note which skill surfaced it
   (/review or /code-review) when relevant.
3. A severity per finding (Blocker / Concern / Nit / Info / Trivial).
4. A "Reasoning" paragraph per finding explaining why it matters and the
   concrete impact.
5. A leading summary table.
6. A "What's Good" section consolidating positive observations agreed on
   across reviewers.
7. A Verdict at the end (Approve / Approve with nits / Request changes / Block)
   and a short list of recommended in-PR fixes.

============================================================================
FILE 2: $ABS_OUT/merged-review.json  (machine-readable — the walkthrough's input)
============================================================================

Strict schema:

{
  "summary": "one paragraph synthesis prose",
  "findings": [
    {
      "id": 1,
      "severity": "Blocker" | "Concern" | "Nit" | "Info" | "Trivial",
      "title": "short title",
      "file": "path/to/file.go" | null,
      "line": 30 | null,
      "body": "the finding text in markdown",
      "suggestion": "concrete fix text in markdown — may be empty string",
      "reviewer_agreement": "N/M (e.g. 3/4)",
      "source_skills": ["/review", "/code-review"]
    }
  ],
  "positives": [
    "consensus positive observation 1",
    "consensus positive observation 2"
  ],
  "verdict": "Approve" | "Approve with nits" | "Request changes" | "Block"
}

Rules for the JSON:
- file/line are null when the finding doesn't have a single anchor (e.g.,
  a cross-cutting concern, a PR-description issue, a finding about a file
  not in the diff).
- "Info / positive" entries that are notes (not actions) go in "positives"
  as plain strings, not in "findings".
- Severity is mandatory and matches between the .md and .json.
- IDs are sequential starting at 1, in the same order as the .md.
- Output valid JSON. Use the Write tool to create the file directly.

============================================================================
CRITICAL: verify before propagating
============================================================================

Before propagating any finding that claims a correctness risk, verify it
against the actual code by reading the relevant files. If a concern doesn't
reproduce, downgrade it to a Nit (in both .md and .json) and add a one-line
note to the body explaining what you checked. Do not propagate unverified
claims — rigor in presentation does not equal rigor in investigation.

Do NOT edit any file other than the two output files above. Do NOT post to
GitHub. Do NOT modify the source review files.
EOF

cd "$REPO_ROOT"
"$CLAUDE_BIN" -p "$MERGE_PROMPT" \
    --model "$MODEL" \
    --permission-mode acceptEdits \
    > "$ABS_OUT/merger.log" 2>&1

OK_MD=0
OK_JSON=0
[[ -s "$ABS_OUT/merged-review.md"   ]] && OK_MD=1
[[ -s "$ABS_OUT/merged-review.json" ]] && OK_JSON=1

if [[ $OK_MD -eq 1 && $OK_JSON -eq 1 ]]; then
    # Validate JSON parses.
    if jq -e . "$ABS_OUT/merged-review.json" >/dev/null 2>&1; then
        echo "  merger: ✓ merged-review.md + merged-review.json"
        echo
        echo "Done."
    else
        echo "  merger: ✗ merged-review.json is not valid JSON (see merger.log)" >&2
        exit 1
    fi
else
    [[ $OK_MD   -eq 0 ]] && echo "  merger: ✗ merged-review.md not produced" >&2
    [[ $OK_JSON -eq 0 ]] && echo "  merger: ✗ merged-review.json not produced" >&2
    echo "  See $ABS_OUT/merger.log for details." >&2
    exit 1
fi
