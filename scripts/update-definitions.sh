#!/usr/bin/env bash
# Maintainer script: refreshes binary-codec/definitions/definitions.json from a
# node's server_definitions response.
#
# The embedded document must stay a verbatim snapshot of what a node reports, so
# this script never edits entries. It unwraps the JSON-RPC `result` envelope,
# drops the request-scoped `status` key, and writes the rest back with sorted
# keys and two-space indentation.
#
# Prerequisites:
#   - curl
#   - jq
#
# Usage:
#   bash scripts/update-definitions.sh                                        # mainnet
#   bash scripts/update-definitions.sh --node http://127.0.0.1:5005/          # localnet
#   NODE_URL=https://s.devnet.rippletest.net:51234/ bash scripts/update-definitions.sh

set -euo pipefail

NODE_URL="${NODE_URL:-https://s1.ripple.com:51234/}"
REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DEFINITIONS_FILE="$REPO_ROOT/binary-codec/definitions/definitions.json"

usage() {
	echo "Usage: bash scripts/update-definitions.sh [--node URL]"
	echo
	echo "  --node URL  node to fetch server_definitions from (default: $NODE_URL)"
	echo "              also settable through the NODE_URL environment variable"
}

while [[ $# -gt 0 ]]; do
	case "$1" in
	--node)
		NODE_URL="$2"
		shift 2
		;;
	-h | --help)
		usage
		exit 0
		;;
	*)
		echo "Unknown argument: $1" >&2
		usage >&2
		exit 1
		;;
	esac
done

require_command() {
	if ! command -v "$1" >/dev/null 2>&1; then
		echo "Missing required command: $1" >&2
		exit 1
	fi
}

require_command curl
require_command jq

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

rpc() {
	curl --silent --show-error --fail --max-time 60 \
		-X POST "$NODE_URL" \
		-H 'Content-Type: application/json' \
		-d "{\"method\":\"$1\",\"params\":[{}]}"
}

echo "Fetching server_definitions from $NODE_URL..."
rpc server_definitions >"$TMP_DIR/response.json"

# A node that rejects the request still answers with HTTP 200, so the payload
# decides whether the response is usable.
if ! jq -e '.result | (has("FIELDS") and has("TYPES") and has("hash"))' "$TMP_DIR/response.json" >/dev/null; then
	echo "Unexpected server_definitions response:" >&2
	head -c 500 "$TMP_DIR/response.json" >&2
	echo >&2
	exit 1
fi

jq -S '.result | del(.status)' "$TMP_DIR/response.json" >"$TMP_DIR/definitions.json"

OLD_HASH="$(jq -r '.hash // "none"' "$DEFINITIONS_FILE" 2>/dev/null || echo "none")"
NEW_HASH="$(jq -r '.hash' "$TMP_DIR/definitions.json")"

# Reporting the build the snapshot came from is a convenience, so a node that
# does not answer server_info must not discard the definitions already fetched.
NODE_VERSION="$(rpc server_info | jq -r '.result.info.rippled_version // .result.info.build_version // "unknown"' || echo "unknown")"

mv "$TMP_DIR/definitions.json" "$DEFINITIONS_FILE"

echo "Node version: $NODE_VERSION"
if [ "$OLD_HASH" = "$NEW_HASH" ]; then
	echo "Definitions unchanged (hash $NEW_HASH)."
else
	echo "Definitions updated: $OLD_HASH -> $NEW_HASH"
	echo "Review the diff and run 'make test-binary-codec'."
fi
