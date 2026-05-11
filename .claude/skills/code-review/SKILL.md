---
name: code-review
description: "Review the current branch's diff against go-* anti-pattern rules and xrpl-standards specs. Spawns focused subagents per changed package, plus a single XRPL-domain reviewer when protocol files are touched, and synthesizes findings cross-cutting. Use when the user types /code-review, finishes a task and wants a self-review, or fetches a teammate's branch and wants to audit it before merging."
---

# Code Review

Multi-reviewer audit of the current branch's changes. The orchestrator computes the diff, classifies changed files by Go top-level package, and spawns subagents that apply the team's `go-*` rules and `xrpl-standards` specs. Findings come back as JSON, are deduped + capped, and rendered grouped by file.

This skill **does not** replace `/review` or `/security-review`. It's a **rule-based audit** keyed to this repo's codified standards.

## Pipeline

### 1. Resolve scope

Default: committed changes on the current branch vs `main`.

**Prompt the user (`AskUserQuestion`) only on ambiguous cases:**

| Situation                               | Ask                                                    |
| --------------------------------------- | ------------------------------------------------------ |
| Dirty tree, no `--include-*` flag       | committed only / + uncommitted / + untracked           |
| Current branch == base (e.g. on `main`) | last commit / last N commits / pick a package / cancel |
| Base does not exist                     | pick a base from `git branch -a`                       |

If the user passed flags or a positional package, those resolve the ambiguity — don't re-prompt.

### 2. Classify changed files

A package = top-level dir under repo root: `address-codec/`, `binary-codec/`, `keypairs/`, `xrpl/`, `pkg/`. Anything else (`examples/`, root-level files) → `other`.

Build `{package: [files]}`.

### 3. Decide single-pass vs fan-out

- **Single-pass** when added+removed lines < 200 AND only one package touched: run the go-mistakes rubric inline (no subagent), skip to step 6.
- **Fan-out** otherwise.

### 4. Spawn subagents (in parallel — single message, multiple `Agent` calls, `subagent_type: "general-purpose"`)

**Go-mistakes — one per changed package.** Use `prompts/go-mistakes.md` as the prompt with substitutions:

- `{{PACKAGE}}` → package name
- `{{CHANGED_FILES}}` → newline-separated file list
- `{{BASE}}`, `{{INCLUDE_UNCOMMITTED}}`, `{{INCLUDE_UNTRACKED}}`

**XRPL-domain — at most one, conditional.** Spawn if any changed file matches:

- `xrpl/transaction/**`, `xrpl/ledger-entry-types/**`, `xrpl/queries/**`
- `binary-codec/types/**`, `binary-codec/definitions/**`

Use `prompts/xrpl-domain.md` with the same substitutions but `{{CHANGED_FILES}}` is the XRPL-relevant slice across packages.

If `--only=go` → skip XRPL subagent. If `--only=xrpl` → skip go-mistakes subagents.

### 5. Collect output

Each subagent returns prose followed by **one** ```json fenced block on the last lines (schema below).

Extract the **last** ```json block. On parse failure, retry once: `"Your previous output was not valid JSON. Parser error: <err>. Reply with ONLY the JSON array, no prose."`

If the retry also fails, fall back: include the subagent's raw text as a single `concern`-level finding with `source: "<reviewer>/unparseable"`. Never abort the run on one bad subagent.

### 6. Cross-cutting synthesis

After all subagents return, the orchestrator does its own pass on the merged JSON + the diff:

1. **Dedup** exact `(file, line, message)` matches — keep one, merge `source` fields.
2. **Import boundary check.** Per `CLAUDE.md`'s layering, low-level packages (`address-codec/`, `keypairs/`, `binary-codec/`, `pkg/`) must NOT import from `xrpl/`. Flag violations as blocker / `cross-cutting/layering`.
3. **Shared-package consumers.** If a low-level package's exported symbol changed, `git grep` for callers in `xrpl/`. If a caller wasn't also updated in this diff, flag concern / `cross-cutting/consumer`.
4. **CHANGELOG check.** If any `.go` file under the four real packages changed, verify `CHANGELOG.md` was modified in this diff. If not, emit concern / `cross-cutting/changelog`.

### 7. Render

Apply in order:

1. `--severity` filter (blocker → blocker only; concern → blocker+concern; nit → all).
2. `--only` filter (drops findings whose `source` doesn't match; cross-cutting findings always pass).
3. **Per-file per-source nit cap = 5.** Excess collapsed to `+N more nits`.
4. **Global cap = 20.** Blockers and concerns never drop. If `blockers + concerns > 20`, drop nits entirely.
5. Group by file (path order). Within a file, sort by severity (blocker → concern → nit) then line number.

Print the report. No follow-up actions, no auto-fix.

## Subagent JSON schema

```json
[
  {
    "file": "binary-codec/types/signers.go",
    "line": 142,
    "severity": "blocker",
    "source": "go-errors/error-wrapping",
    "message": "Wraps with %v, loses error chain.",
    "suggestion": "Use %w."
  }
]
```

- **`file`** — repo-relative, forward slashes.
- **`line`** — 1-based; use `1` for file-level findings.
- **`severity`** — `"blocker"` (must fix), `"concern"` (likely problem), `"nit"` (minor).
- **`source`** — citation. `go-<skill>/<rule-filename-without-md>`, `xrpl-domain/XLS-NN` or `xrpl-domain/<category>`, `general/<category>`, `cross-cutting/<category>`.
- **`message`** — what's wrong and why. Reference identifiers by name.
- **`suggestion`** — optional, concrete fix.

## Output format

Header line: `Branch: <current> vs <base>` and severity counts. Then per-file sections. Markers: 🔴 blocker, 🟡 concern, 🔵 nit.

Banners (prepend if applicable):

- Dirty tree without `--include-uncommitted`: "⚠️ N uncommitted file(s) not in this review. Run with --include-uncommitted to include."
- Untracked `.go` files without `--include-untracked`: "⚠️ N untracked .go file(s) not reviewed: ..."
- Subagent unparseable: "⚠️ <reviewer> returned unparseable output."

If zero findings after filtering: print one line — `✅ No issues found. The changes look good.`

Don't add filler praise, "files reviewed" lists, or per-reviewer section headers (the `[<source>]` tag attributes each finding).
