---
name: multi-review
description: "Spawn 4 parallel Claude reviewers on the current branch (2× /review, 2× /code-review), synthesize them into one merged review, then walk findings interactively. Two modes — fix locally before opening a PR (developer mode), or post inline + body comments back on the PR (reviewer mode). Use when /code-review alone feels shallow, when you want a self-review of in-flight work, or when auditing a teammate's PR."
---

# Multi-Review

Orchestrator skill that gets four independent perspectives on a branch's changes and helps you act on the findings. Lives on top of `/review` and `/code-review`, not a replacement for either.

The skill is the brain. The bash workhorse is at `scripts/multi-review.sh` — it spawns the subprocesses and produces a structured merged review. The skill handles intent, scope resolution, and the post-review walkthrough.

## Cost note

Four concurrent Opus 4.7 sessions plus a merger is ~5× the spend of a single `/review` or `/code-review`. Mention this once to the user at the start if they didn't invoke with an explicit PR number (signals an exploratory run).

## Pipeline

### 1. Pre-flight

**Hard-fail with a clear message and stop** if any of these are true:
- Not in a git repo
- `claude` CLI not on `PATH`
- `gh` CLI not on `PATH`
- `/review` or `/code-review` skills not discoverable (check `.claude/skills/` and the user's plugin marketplace)
- Current branch equals the base branch

**Detect — never trust defaults blindly:**
- **Repo**: parse `owner/repo` from `git remote get-url origin`. The `gh` default repo can disagree (fork-vs-upstream); always derive from git. Pass this as `GH_REPO` to the script.
- **Base branch**: `git symbolic-ref refs/remotes/origin/HEAD --short` → strip `origin/`. Fall back to `main`. Pass as `BASE_BRANCH`.
- **PR number** (reviewer mode only — see step 2): priority chain:
  1. Positional argument: `/multi-review 272`
  2. `gh pr view --repo OWNER/REPO --json number,state --jq 'select(.state=="OPEN").number'`
  3. `AskUserQuestion`: "Which PR number?"

**Auto-resolve silently** — don't ask, just decide and log a one-liner in the chat:
- **Dirty tree**: review committed only.
- **No PR exists** (developer mode is fine without one; reviewer mode must have one): only `/code-review` agents will run.
- **Stale base**: run `git fetch origin <base>` once before continuing.

The principle: prompts cost user attention. Only ask on genuine ambiguity where the wrong default would produce a misleading review.

### 2. Mode selection

Ask via `AskUserQuestion` (exactly one question, two options):

> **Are you reviewing your own changes or someone else's?**
>
> - **My own changes** — fix findings locally before opening or updating the PR.
> - **Someone else's PR** — post findings back as inline + body comments.

Save the choice as `MODE` (`developer` | `reviewer`).

If `reviewer` and no PR is detectable after step 1, prompt for the PR number now.
If `developer` and the current branch == base, that already hard-failed in step 1.

### 3. Run the multi-review script

**Compute the run directory name:**
- Reviewer mode: `pr-<N>-<branch-slug>-<timestamp>`
- Developer mode: `<branch-slug>-<timestamp>`

Where:
- `<branch-slug>` = current branch name with `/` → `-` and other shell-unsafe chars stripped (`[^A-Za-z0-9._-]` → `-`).
- `<timestamp>` = `date +%Y%m%d-%H%M%S`.

Then invoke the script (run directory is the first positional arg, the rest are env vars):

```bash
RUN_DIR="reviews/$RUN_NAME"
PR_NUMBER="$PR_NUMBER" \
GH_REPO="$GH_REPO" \
BASE_BRANCH="$BASE_BRANCH" \
MODEL="claude-opus-4-7" \
bash .claude/skills/multi-review/scripts/multi-review.sh "$RUN_DIR"
```

The script handles parallel spawn, wait, merger. It produces:
- `$RUN_DIR/review-skill-{1,2}.md` (when PR_NUMBER is set)
- `$RUN_DIR/code-review-{1,2}.md`
- `$RUN_DIR/merged-review.md` (human-readable)
- `$RUN_DIR/merged-review.json` (machine-readable — the walkthrough's input)
- `$RUN_DIR/*.log`

Stream a one-line "running 4 reviewers + merger, expect a few minutes" to the user. Don't poll; the script blocks until done.

### 4. Parse merged-review.json

Read `$RUN_DIR/merged-review.json`. Schema:

```json
{
  "summary": "one paragraph synthesis",
  "findings": [
    {
      "id": 1,
      "severity": "Blocker | Concern | Nit | Info | Trivial",
      "title": "short title",
      "file": "path/to/file.go",
      "line": 30,
      "body": "the finding text",
      "suggestion": "concrete fix text (may be empty)",
      "reviewer_agreement": "3/4"
    }
  ],
  "positives": ["consensus positive observations, prose"],
  "verdict": "Approve | Approve with nits | Request changes | Block"
}
```

**Filter for the walkthrough**: keep `severity in {Blocker, Concern, Nit}`. Drop `Info` and `Trivial` and the positives.

**Sort**: by severity ascending (Blocker → Concern → Nit), then by reviewer_agreement descending within a tier.

**Summarize the dropped items** as a one-paragraph chat preamble before starting the walkthrough — e.g., *"Skipping 3 Info/Trivial findings (perf observations, allocation note). 2 positives noted: empty-string fix is a real bug, perf win on `isHex`. Walking 5 actionable findings."*

### Display format for findings (used by both walkthrough modes)

Each finding is shown to the user in the chat the same way it appears in `merged-review.md` — so the walkthrough reads like an interactive version of the file. Use this template:

```
─────────────────────────────────────────────
Finding <id>/<total> · <Severity>
<title>

Reviewer agreement: <N/M>  ·  Sources: </review>, </code-review>
File: <file>:<line>      (omit this line if file is null)

<body>

Suggestion:
<suggestion>             (omit this block if suggestion is empty)
─────────────────────────────────────────────
```

For reviewer mode only, **also** show how the comment would be routed before asking the action — append:

```
Would post as: inline comment at <file>:<line>
                    — or —
Would post in: general review body
  Preamble: (<the routing-explanation note>)
```

This makes the routing transparent and gives the user a final sanity check before they answer the action menu.

### 5. Walkthrough — developer mode

Per-finding `AskUserQuestion` (4 options):

| Action | Behavior |
|---|---|
| **Fix it for me** | Read the finding's `suggestion` field. If concrete (code snippet, line ranges), draft the edit via Edit tool, show the diff to the user, mark the finding addressed. If vague, fall back to "Chat about it" inline — present 2–3 possible resolutions, let the user pick. |
| **Fix it later** | Append the finding to `$RUN_DIR/todo.md` (create if missing). Format: `## <severity>: <title>\n\nFile: `<file>:<line>`\n\n<body>\n\nSuggestion: <suggestion>\n\n---`. |
| **Skip** | Mark deferred-without-todo. Don't fix, don't track. |
| **Chat about it** | Discuss with the user. They can change their mind to any of the other 3 actions afterward. |

Do **not** commit or stage. Leave the working tree dirty for the user.

After the loop, summarize in chat: *"Applied N fixes across M files. K findings sent to `$RUN_DIR/todo.md`. L skipped."*

### 6. Walkthrough — reviewer mode

**No upfront mode question.** Always batched — posts go out as one PR review at the end, after a per-comment verification pass (see step 7). The earlier draft of this skill asked "batched or drive-by?" up front; that question is bad UX because the user can't meaningfully decide before they see what's going to be posted. Always batch.

**Comment routing — automatic, not a user choice.** Each finding the user chooses to post is routed to exactly one place; the finding's shape decides where:

- **File-anchored findings** (`file` and `line` are non-null AND `line` falls inside a diff hunk on the PR's HEAD commit) → **inline comment** at that file:line.
- **Everything else** (cross-cutting concerns, findings about files not in the diff, findings without a specific line anchor, PR-description issues, findings whose anchor line falls outside the diff hunks) → **general review body**.

This routing happens silently. The user is not asked "inline or body?" per finding — the data answers it. We hit `Line could not be resolved` 422s during testing precisely because anchors outside hunks were attempted; the routing rule prevents that by construction.

Per-finding `AskUserQuestion` (3 options):

| Action | Behavior |
|---|---|
| **Post** | Route per the rule above. Inline → accumulate `{path, line, body}` in `$RUN_DIR/draft-review.json`'s `comments` array. Body → append to the running general-body buffer with a `### <finding title>` heading. |
| **Skip** | Mark skipped, don't include. |
| **Chat about it** | Workshop the comment text with the user. Afterward they can pick **Post** or **Skip**. They cannot override the routing — that follows the rule. |

**When the routing pushes a finding into the body**, the skill prepends a one-line context note so it's clear *why* it's in the body and not inline. Examples:

- `(File \`xrpl/wallet/wallet.go\` not in this PR's diff but affected by the contract change.)`
- `(Concerns the PR description / changelog wording, not a specific file.)`
- `(Anchor line falls outside the diff hunks for \`<file>\`; surfaced here.)`

### Comment text conventions (what gets posted vs what's shown in chat)

The chat display (the merged-review-style template above) shows internal metadata so the user can make decisions: `Reviewer agreement: 3/4`, `Sources: /review, /code-review`, the finding `id`, etc.

**None of that metadata is included in the comment text posted to GitHub.** The posted comment is just the substance:

- Optional severity prefix: `Nit: `, `Concern: ` (omit for Blockers — the REQUEST_CHANGES event carries that signal).
- The finding `body` from the JSON.
- The `suggestion` block, if present.

Things that must NEVER appear in a posted comment:

- "(2/4 reviewers)" / "Reviewer agreement: N/M" — internal weighting, noise to the author.
- "Multi-review surfaced this" / mention of the skill / "(4 reviewers · /review × 2 + /code-review × 2)" — credit boilerplate. The author doesn't care which tool produced the review.
- Finding IDs or sequence numbers — those are walkthrough nav, not part of the comment.
- The finding `title` if it would be redundant with the first line of `body`. If you do prepend the title for readability, drop any "(2/4 reviewers)"-style suffix from it.

The walkthrough/verification chat shows the metadata. The PR sees only the substance.

### 7. Submit / wrap up

**Reviewer mode — single preview, then submit:**

The walkthrough already approved each comment individually. Don't re-walk them one by one — that's redundant work. Instead, show the *full payload* once as a single preview, and ask one question: ship it, edit a specific item, or cancel.

1. **Render the full preview** — concatenate every staged comment with its routing context. No metadata. Concretely:

   ```
   ════════════════════════════════════════════
   Ready to post to PR #<N>: <M> inline + <K> body comments
   Event: COMMENT (default — change below)
   ════════════════════════════════════════════

   ── Inline comments ──

   [1] <file>:<line>                          [Nit]
       <comment text>

   [2] <file>:<line>                          [Concern]
       <comment text>

   ── General review body ──

   <one-sentence verdict>

   [3] (File `xrpl/wallet/wallet.go` not in this PR's diff …)
       <comment text>

   ════════════════════════════════════════════
   ```

   IDs `[N]` are local to the preview, used by the action menu — not posted.

2. **Single confirmation** — `AskUserQuestion` with 4 options:

   | Action | Behavior |
   |---|---|
   | **Submit as is** *(default)* | Continue to step 3 (event-type prompt). |
   | **Edit comment N** | Ask which `[N]` and what to change. Apply, re-render the preview, ask again. |
   | **Drop comment N** | Ask which `[N]`. Remove from payload, re-render, ask again. |
   | **Cancel review** | Abort. Save `$RUN_DIR/draft-review.json` for inspection. Nothing posted. |

   The Edit/Drop branches loop back to the preview after applying — so the user can make a few targeted changes and then `Submit`. They never re-walk every comment.

3. **Event type**: after the preview is approved, ask:

   > **What event type for the review?**
   > - `COMMENT` *(default)* — neutral feedback
   > - `REQUEST_CHANGES` — required changes before merge
   > - `APPROVE` — sign off

   Recommend `REQUEST_CHANGES` if any verified comment came from a Blocker-severity finding. Don't force.

4. **Compose the final payload**:
   ```json
   {
     "commit_id": "<HEAD of PR>",
     "event": "<chosen event>",
     "body": "<accumulated general body>",
     "comments": [{ "path": "...", "line": ..., "side": "RIGHT", "body": "..." }, ...]
   }
   ```

   **General body format — keep it short.** The body is what the author reads first. It should be:

   - At most one short sentence of verdict context (e.g., *"Approve modulo the inline nits."*). Skip entirely if the inline comments speak for themselves.
   - Then the body-routed findings, each with its routing-preamble line and the comment text.

   Things that must NOT appear in the body:

   - A "Multi-review summary" header or any mention of the skill / number of reviewers / which agents produced it.
   - A restatement of what the PR does. The author already knows. The body summarizes findings, not the change.
   - A "Verdict: Approve with nits" boilerplate line as a separate header — that signal is carried by the `event` field, not the body text.
   - Findings without a real reason to be there. If you have zero body-routed findings, the body should be one short verdict sentence, or empty.

   Example of a good body (zero body-routed findings):
   > Approve modulo the inline nits.

   Example of a good body (one body-routed finding):
   > Approve modulo the inline nits.
   >
   > **(File `xrpl/wallet/wallet.go` not in this PR's diff but affected by the contract change.)** Orphan `maps.Copy` at `wallet.go:144-146` is now dead weight — same workaround as the three you cleaned up in `counterparty_signer.go`. Consider removing for consistency.

   Bad body (the one the user flagged): multi-paragraph PR summary + verdict header + reviewer-process credit + then findings. Cut all of that.

5. **Pre-clean** any *pending* review by the current user on this PR — delete it via `gh api .../pulls/$PR/reviews/<id> -X DELETE`. (`User can only have one pending review per pull request` 422 otherwise.)

6. **Submit**: `gh api repos/$GH_REPO/pulls/$PR/reviews -X POST --input <payload>`. Print the review URL.

**Developer mode:**
- See step 5. No verification pass — file edits are visible in the working tree as they happen, so verification is just "look at the diff I just showed you" per finding.

## Notes

- **REPO is from git, not gh.** Always parse `git remote get-url origin`. The user may be on a fork; `gh`'s notion of "default repo" can diverge from the actual upstream.
- **The merger must verify before propagating.** The bash script's merger prompt instructs the merger to read actual source files before propagating any correctness claim. We hit a false positive in testing (a `Concern`-level "nested mutation risk" that didn't reproduce); the verification step caught it. Don't relax this.
- **Never auto-commit or auto-amend.** Always leave staging to the user. The skill is allowed to edit files (in developer mode "Fix it for me") and to call `gh api` (in reviewer mode), but never to mutate git refs.
- **Cost: 4× Opus is real money.** Mention it once at the start of a run if no PR number was passed (heuristic: explicit PR # = the user knows what they're doing). Don't nag.

## Resuming or re-walking past runs

The skill never overwrites or deletes previous run directories. To walk an existing run again, the user says so in chat — e.g., *"walk through reviews/pr-272-xrpl-fix-uint64-ambiguity-20260512-094530"*. Resume by jumping directly to step 4 with that `$RUN_DIR`. No flag, no auto-detect — just natural language.
