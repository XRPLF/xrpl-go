#!/usr/bin/env bash

set -euo pipefail

if (($# != 2)); then
	echo "Usage: $0 <all-api-changes-file> <incompatible-api-changes-file>" >&2
	exit 2
fi

api_changes_file=$1
incompatible_changes_file=$2

if [[ ! -f "$api_changes_file" || ! -f "$incompatible_changes_file" ]]; then
	echo "API compatibility report input is missing." >&2
	exit 2
fi

if [[ -z "${GITHUB_STEP_SUMMARY:-}" ]]; then
	echo "GITHUB_STEP_SUMMARY is not set." >&2
	exit 2
fi

{
	echo "## Exported API changes"
	if [[ -s "$api_changes_file" ]]; then
		echo '```text'
		cat "$api_changes_file"
		echo '```'
	else
		echo "No exported API changes."
	fi
	echo
	echo "## Compatibility result"
} >>"$GITHUB_STEP_SUMMARY"

if [[ ! -s "$incompatible_changes_file" ]]; then
	echo "✅ No breaking exported API changes were detected." >>"$GITHUB_STEP_SUMMARY"
	echo "No breaking exported API changes were detected."
	exit 0
fi

title_acknowledged=false
checkbox_acknowledged=false

# The exclamation mark must immediately precede the first colon in the PR title.
if [[ "${PR_TITLE:-}" =~ ^[^:]+!: ]]; then
	title_acknowledged=true
fi

if printf '%s\n' "${PR_BODY:-}" | grep -Eiq '^[[:space:]]*[-*][[:space:]]+\[[xX]\][[:space:]]+Breaking[[:space:]]+change([[:space:]]|$)'; then
	checkbox_acknowledged=true
fi

{
	echo "⚠️ **apidiff** detected these breaking exported API changes:"
	echo
	echo '```text'
	cat "$incompatible_changes_file"
	echo '```'
	echo
	echo "### Required acknowledgement"
	if [[ "$title_acknowledged" == true ]]; then
		echo '- ✅ The PR title has `!` immediately before its first `:`.'
	else
		echo '- ❌ The PR title must have `!` immediately before its first `:`.'
	fi
	if [[ "$checkbox_acknowledged" == true ]]; then
		echo '- ✅ `Breaking change` is checked under `Type of change`.'
	else
		echo '- ❌ Check `Breaking change` under `Type of change` in the PR description.'
	fi
} >>"$GITHUB_STEP_SUMMARY"

if [[ "$title_acknowledged" == true && "$checkbox_acknowledged" == true ]]; then
	echo >>"$GITHUB_STEP_SUMMARY"
	echo "The breaking change is explicitly acknowledged. Reviewers must verify that it is intentional and documented." >>"$GITHUB_STEP_SUMMARY"
	echo "::notice title=Acknowledged breaking API change::apidiff detected a breaking exported API change. The PR title and description acknowledge it."
	cat "$incompatible_changes_file"
	exit 0
fi

{
	echo
	echo "### How to resolve this failure"
	echo "Restore API compatibility, or acknowledge an intentional breaking change:"
	echo '1. Use a title such as `feat!: remove old API` or `fix(client)!: change method signature`.'
	echo '2. Check `Breaking change` under `Type of change` in the PR description.'
} >>"$GITHUB_STEP_SUMMARY"

missing_requirements=()
if [[ "$title_acknowledged" != true ]]; then
	missing_requirements+=("add ! immediately before the first colon in the PR title")
fi
if [[ "$checkbox_acknowledged" != true ]]; then
	missing_requirements+=("check Breaking change in the PR description")
fi
printf -v missing_message '%s; ' "${missing_requirements[@]}"
missing_message=${missing_message%; }

echo "Detected an unacknowledged breaking exported API change." >&2
cat "$incompatible_changes_file" >&2
echo "Correct the API change or acknowledge it. Missing: $missing_message." >&2
echo "::error title=Unacknowledged breaking API change::apidiff detected a breaking exported API change. Correct it or acknowledge it. Missing: $missing_message."
exit 1
