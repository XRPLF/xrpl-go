---
name: address-pr-comments
description: "Fetch PR review comments from GitHub, interactively walk through each one, and choose how to address it — fix the code, reply, ask for clarification, or skip. Use when you have PR feedback to work through."
---

# Address PR Comments

Interactive skill for triaging and addressing GitHub PR review comments one by one.

## Step 1: Identify the PR and repo

### 1a. Resolve the repository

The `gh` CLI default repo may differ from the actual working directory repo. Always derive the repo from git:

```bash
git remote get-url origin
```

Parse `owner/repo` from the URL (handles both `git@github.com:Owner/Repo.git` and `https://github.com/Owner/Repo.git`). Use this as `REPO` for all `gh api` calls via `--repo` flag or explicit URL paths.

### 1b. Resolve the PR number

Use this priority order:

1. **Argument**: if the user passed a number (e.g., `/address-pr-comments 955`), use it directly
2. **Current branch**: check if the current branch has an open PR:
    ```bash
    gh pr view --repo REPO --json number,title,state --jq 'select(.state=="OPEN")'
    ```
3. **Ask the user**: if neither works, use AskUserQuestion to request the PR number

### 1c. Fetch comments

Fetch inline review comments (code-level):

```bash
gh api repos/OWNER/REPO/pulls/PR_NUMBER/comments
```

Fetch general conversation comments (non-inline):

```bash
gh api repos/OWNER/REPO/issues/PR_NUMBER/comments
```

Also fetch PR metadata for context:

```bash
gh pr view PR_NUMBER --repo OWNER/REPO --json number,title,author,headRefName,state
```

## Step 2: Parse and group comments

Group comments by:

1. **File + line** for inline review comments
2. **Conversation thread** — group replies together, show only the root comment for triage
3. **Author** — note who left the comment

Skip:

-   **Self-comments** — match against `git config user.name` and the PR author login
-   **Reply comments** — comments with `in_reply_to_id` set (these are thread replies, shown under their parent)
-   **Unselected bot comments** — see bot filtering below

### Bot comment filtering

Bot authors have a `type` field of `"Bot"` in the GitHub API response. If any bot comments are detected:

1. List the distinct bot names and their comment counts (e.g., `coderabbitai[bot] — 7 comments`)
2. Use AskUserQuestion to ask which bots to include, with options like:
    - **Include all bots**
    - **Include [bot1] only** (one option per bot, if ≤ 3 bots)
    - **Exclude all bots**
3. Filter comments according to the user's choice

If no bot comments exist, skip this step silently. Tag included bot comments with the bot name in the summary table so the user can distinguish human vs. bot feedback.

Sort by file path, then line number, so related comments are addressed together.

## Step 3: Show summary

Before starting the interactive loop, show a summary:

```md
## PR #123 — 12 comments to address

| #   | File                     | Line | Author    | Preview                        |
| --- | ------------------------ | ---- | --------- | ------------------------------ |
| 1   | src/module/service.ts    | L42  | reviewer1 | "This should use BigNumber..." |
| 2   | src/module/controller.ts | L15  | reviewer2 | "Missing validation for..."    |
| ... |                          |      |           |                                |

Ready to walk through each comment.
```

## Step 4: Interactive loop — one comment at a time

For each comment, present it in full, then use AskUserQuestion to let the user choose how to handle it.

### 4a. Present the comment

Show:

-   The **full comment text** (including any code suggestions from the reviewer)
-   The **file and line number**
-   The **surrounding code** (read ~10 lines around the referenced line)
-   Any **thread replies** beneath the root comment

### 4b. Ask the user what to do

> **Constraint:** AskUserQuestion supports a maximum of 4 explicit options (plus an automatic "Other" free-text option). Use a two-round flow to fit all 5 actions.

**Round 1** — present these 4 options:

| Option           | What happens                                                                                                                                                                                                                                        |
| ---------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Explain more** | Provide a detailed explanation of the issue: what the reviewer is pointing out, why it matters, what the current code does vs. what the suggestion would change, and the concrete risks or benefits. After explaining, re-ask with Round 2 options. |
| **Fix it**       | Read the relevant code, understand the reviewer's suggestion, apply the fix, then draft a short reply confirming the fix (e.g., "Fixed in [commit]" or "Good catch, updated.")                                                                      |
| **Reply**        | Proceed to Round 2 to choose the reply type.                                                                                                                                                                                                        |
| **Skip**         | Move to the next comment without action.                                                                                                                                                                                                            |

**Round 2 (Reply)** — if the user picks "Reply", ask a follow-up:

| Option                          | What happens                                                                                                          |
| ------------------------------- | --------------------------------------------------------------------------------------------------------------------- |
| **Current approach is correct** | Draft a reply explaining why the current code is intentional. Show the draft to the user for approval before posting. |
| **Ask for clarification**       | Draft a reply asking the reviewer to elaborate. Show the draft for approval.                                          |

**Round 2 (after Explain more)** — re-present the action options without "Explain more":

| Option     | What happens                    |
| ---------- | ------------------------------- |
| **Fix it** | Same as Round 1                 |
| **Reply**  | Proceed to reply type selection |
| **Skip**   | Move to the next comment        |

#### Explain more

When the user picks "Explain more":

1. Read the referenced file and broader context (~50 lines around the comment)
2. Research the topic — check related files, production code, documentation, or similar patterns in the codebase
3. Explain clearly:
    - **What the reviewer is pointing out** — restate the issue in plain terms
    - **What the current code does** — walk through the relevant logic
    - **What the suggestion would change** — concrete before/after behavior
    - **Risks of not fixing** — what could go wrong in practice
    - **Risks of fixing** — any potential side effects of the suggested change
4. After the explanation, re-present Round 2 action options (Fix / Reply / Skip) without the "Explain more" option

### 4c. Execute the chosen action

#### Fix it

1. Read the file at the referenced line (broader context: ~30 lines around it)
2. Understand what the reviewer is asking for
3. Apply the code fix using Edit
4. Show the diff to the user for confirmation
5. Draft a reply: "Fixed — [brief description of what changed]."
6. Store the reply for later (do NOT post yet)

#### Reply — current approach is correct

1. Read the file to understand the full context
2. Draft a concise, professional reply explaining the rationale
3. Show the draft to the user via AskUserQuestion with options: "Post as-is", "Edit before posting" (Other option covers this), "Skip"
4. Store the approved reply for later

#### Ask for clarification

1. Draft a question that references the specific code and asks what the reviewer would prefer
2. Show the draft for approval (same flow as above)
3. Store the approved reply for later

## Step 5: Summary and batch post

After all comments are processed, show a recap:

```md
## Summary

| #   | Comment                        | Action  | Reply draft                        |
| --- | ------------------------------ | ------- | ---------------------------------- |
| 1   | "This should use BigNumber..." | Fixed   | "Fixed — switched to BigNumber.js" |
| 2   | "Missing validation for..."    | Reply   | "This is validated upstream in..." |
| 3   | "Consider extracting..."       | Skipped | —                                  |
| ... |                                |         |                                    |

Code changes: 3 files modified
Replies ready: 8 drafts
```

Then ask:

| Option                      | What happens                                                          |
| --------------------------- | --------------------------------------------------------------------- |
| **Post all replies**        | Post every drafted reply to the PR using `gh api`                     |
| **Post replies one by one** | Show each reply and confirm before posting                            |
| **Save replies only**       | Don't post — just keep the drafted text for the user to post manually |
| **Discard replies**         | Don't post anything                                                   |

### Posting replies

For inline review comments, reply to the specific comment thread:

```bash
gh api repos/{owner}/{repo}/pulls/<pr>/comments/<comment_id>/replies -f body="<reply>"
```

For general conversation comments:

```bash
gh api repos/{owner}/{repo}/issues/<pr>/comments -f body="<reply>"
```

## Rules

-   **Never post a reply without user approval.** Always show drafts first.
-   **Never modify code without showing the diff.** The user must see what changed.
-   **Keep replies professional and concise.** No emojis, no fluff. Match the tone of the reviewer.
-   **Group related comments.** If two comments are about the same issue in the same file, address them together.
-   **Respect reviewer suggestions.** When GitHub suggestion syntax is used (`suggestion` blocks), extract the exact suggested code for the fix.
-   **Track what was done.** Maintain a running list of actions taken so the final summary is accurate.
