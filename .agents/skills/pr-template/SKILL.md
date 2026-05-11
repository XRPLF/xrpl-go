---
name: pr-template
description: "Generate a filled PR description from a branch's changes compared to a base branch. Autodetects the repo's GitHub PR template (.github/PULL_REQUEST_TEMPLATE.md and variants), falling back to a built-in template if none exists. Trigger when the user asks: 'create a PR template', 'create a PR template for X branch to Y', 'generate PR description', 'write PR template', 'PR template for feature-branch to main', 'prepare PR', 'draft PR', 'PR description for X to Y'. The user may specify a feature branch and/or a base branch (e.g., 'for feat/foo to main'). If the feature branch is omitted, use the current branch. If the base branch is omitted, default to main."
---

# PR Template Generator

Analyzes a branch's commits and diff against a base branch, then generates a fully filled PR description. Autodetects the repo's GitHub PR template if one exists, otherwise uses a built-in fallback. Writes the result to `PR/<descriptive-name>.md`.

## How It Works

1. **Gather git data** — runs the script to collect branch name, commits, diff stats, and the full diff
2. **Analyze changes** — you read the output and understand what changed, why, and what type of change it is
3. **Generate template** — fill in the PR template with all sections populated
4. **Write file** — save to `PR/<descriptive-name>.md` in the project root

## Parsing User Input

The user will say things like:
- "create a PR template" → current branch, auto-detect base
- "create a PR template for feat/foo" → `feat/foo` branch, auto-detect base
- "create a PR template for feat/foo to main" → `feat/foo` branch, `main` base
- "create a PR template to develop" → current branch, `develop` base
- "PR template for my-branch to main" → `my-branch` branch, `main` base

Extract the **feature branch** (first branch mentioned, or omitted = current) and **base branch** (after "to", or omitted = auto-detect).

## Usage

### Step 1: Gather git information

The script path depends on your environment. Find this skill's install directory and run:

```bash
bash <skill-dir>/scripts/gather-git-info.sh [feature-branch] [base-branch]
```

Common locations:
- **claude.ai**: `/mnt/skills/user/pr-template/scripts/gather-git-info.sh`
- **Claude Code** (`npx skills add`): `.claude/skills/pr-template/scripts/gather-git-info.sh` (relative to project root)

To auto-detect, search for the script:
```bash
find /mnt/skills ~/.claude -name "gather-git-info.sh" -path "*/pr-template/*" 2>/dev/null | head -1
```

**Arguments:**
- `feature-branch` — branch with changes (default: current branch)
- `base-branch` — branch to compare against (default: `main`)

**Examples:**
```bash
# Current branch, auto-detect base
bash <skill-dir>/scripts/gather-git-info.sh

# Specific feature branch to specific base
bash <skill-dir>/scripts/gather-git-info.sh feat/add-signing main

# Current branch to specific base (pass empty string for feature branch)
bash <skill-dir>/scripts/gather-git-info.sh "" develop
```

### Step 2: Analyze and generate the template

After running the script, analyze the JSON output. The output includes:
- Commits, diff, and file list for understanding changes
- `pr_template_path` and `pr_template` fields for the repo's PR template

**Template selection:**
- If `pr_template` is non-empty, the repo has its own GitHub PR template. Use that template's structure and sections. Fill in every section based on the actual changes.
- If `pr_template` is empty, use the **Fallback Template** below.

When using a repo's PR template, preserve its exact structure, sections, and formatting. Fill in placeholders, check applicable checkboxes, and populate all sections with information from the diff.

### Step 3: Write the file

Create the `PR/` directory if it doesn't exist, then write the filled template to `PR/<descriptive-name>.md`.

The `<descriptive-name>` should be a kebab-case slug derived from the PR title (e.g., `add-counterparty-signing-utils.md`, `fix-amm-deposit-validation.md`).

## Fallback Template

Used only when no GitHub PR template is detected in the repo. Every field must be populated based on the actual changes.

```markdown
# <title>

## Description
This PR aims to <description>.

## Type of change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update
- [ ] Refactoring

## Checklist:
- [ ] My code follows the style guidelines of this project
- [ ] I have performed a self-review of my own code
- [ ] I have commented my code where needed
- [ ] I have made corresponding changes to the documentation
- [ ] My changes generate no new warnings
- [ ] I have added tests that prove my fix is effective
- [ ] New and existing unit tests pass locally with my changes

## Changes

- Change 1
- Change 2

## Notes (optional)

<notes or remove section>
```

## Filling Rules

Follow these rules when generating the template. When using the repo's own PR template, adapt these rules to match its sections.

### Title
- Use conventional commit format: `type(scope): description`
- Types: `feat`, `fix`, `refactor`, `docs`, `test`, `chore`, `perf`, `ci`, `build`
- Scope is optional but recommended, use the package/module/area name
- Description is imperative mood, lowercase, no period
- Example: `feat(xrpl): add counterparty signing utilities for LoanSet transactions`

### Description
- Single sentence starting with "This PR aims to..."
- Concise but specific about what the changes accomplish

### Type of change
- Check (`[x]`) all types that apply based on the actual changes
- Multiple types can be checked (e.g., a new feature that also includes docs)

### Checklist
- Check (`[x]`) items that are evidenced by the diff:
  - Tests present in diff -> check "added tests" and "unit tests pass"
  - Docs changed in diff -> check "corresponding changes to documentation"
  - Code follows patterns in the repo -> check "style guidelines"
- Always check "self-review" and "commented where needed"
- Leave unchecked items that cannot be verified from the diff alone

### Changes
- Do not list changelog/CHANGELOG updates as a separate change if they merely summarize the other changes in the PR
- Group by package/module/area when there are changes across multiple areas (use `### area` subheadings)
- Each bullet is a concrete change, what was added, modified, removed, or fixed
- Be specific: mention function/type/file names
- Use sub-bullets for related smaller changes under a main change

### Notes
- Include if there are important caveats, migration steps, breaking change details, or reviewer guidance
- Omit the section entirely if there's nothing noteworthy

## Present Results to User

After writing the file:

```text
PR template written to `PR/<descriptive-name>.md`

**Title:** <the title>
**Template:** <repo template path, or "built-in fallback">
**File:** `PR/<descriptive-name>.md`
**Changes detected:** <N> commits, <N> files changed
```

## Troubleshooting

- **"currently on main"** — switch to a feature branch before running
- **"no common ancestor"** — the branch has no shared history with the base; specify the correct base branch
- **Empty diff** — all changes are already merged or the branch is up to date
- **Large diff truncated** — diffs over 50KB are truncated; the template will still be generated from available data
