#!/bin/bash
# fetch-xls.sh — Fetch and summarize an XRPL Standard (XLS-N) specification
#
# Usage:
#   bash fetch-xls.sh <xls-number> [--full]
#
# Arguments:
#   xls-number  XLS amendment number (e.g. 70, 30, 102)
#   --full      Output the raw README.md without compression
#
# Examples:
#   bash fetch-xls.sh 70
#   bash fetch-xls.sh 30 --full
#   bash fetch-xls.sh 102

set -e

REPO="XRPLF/XRPL-Standards"
API_BASE="https://api.github.com/repos/${REPO}/contents"
RAW_BASE="https://raw.githubusercontent.com/${REPO}/master"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# --- Parse arguments ---
XLS_NUM=""
FULL_MODE=false

for arg in "$@"; do
  case "$arg" in
    --full) FULL_MODE=true ;;
    [0-9]*) XLS_NUM="$arg" ;;
    *) echo "Unknown argument: $arg" >&2; exit 1 ;;
  esac
done

if [ -z "$XLS_NUM" ]; then
  echo "Usage: fetch-xls.sh <xls-number> [--full]" >&2
  exit 1
fi

# Zero-pad to 4 digits
XLS_PADDED=$(printf "%04d" "$XLS_NUM")

echo "Fetching XLS-${XLS_PADDED} from XRPLF/XRPL-Standards..." >&2

# --- Discover directory name via GitHub Contents API ---
API_RESPONSE=$(curl -sf \
  -H "Accept: application/vnd.github.v3+json" \
  "${API_BASE}/" 2>/dev/null) || {
  echo "Error: Failed to reach GitHub API. Check your network connection." >&2
  exit 1
}

# Find the directory matching XLS-NNNN (case-insensitive prefix)
DIR_NAME=$(echo "$API_RESPONSE" | \
  python3 -c "
import sys, json, re
data = json.load(sys.stdin)
pattern = re.compile(r'^XLS-${XLS_PADDED}', re.IGNORECASE)
for item in data:
    if item.get('type') == 'dir' and pattern.match(item.get('name', '')):
        print(item['name'])
        break
" 2>/dev/null)

if [ -z "$DIR_NAME" ]; then
  echo "Error: XLS-${XLS_PADDED} not found in XRPLF/XRPL-Standards." >&2
  echo "Run list-xls.sh to see all available standards." >&2
  exit 1
fi

echo "Found: ${DIR_NAME}" >&2

# --- Fetch README.md ---
README_URL="${RAW_BASE}/${DIR_NAME}/README.md"
README_CONTENT=$(curl -sf "$README_URL" 2>/dev/null) || {
  echo "Error: Could not fetch README.md from ${README_URL}" >&2
  exit 1
}

if $FULL_MODE; then
  echo "=== ${DIR_NAME} (full) ==="
  echo ""
  echo "$README_CONTENT"
else
  # Extract and compress
  EXTRACTED=$(echo "$README_CONTENT" | python3 "${SCRIPT_DIR}/extract-spec.py" 2>/dev/null) || {
    echo "Error: extract-spec.py failed. Is Python 3 installed?" >&2
    exit 1
  }
  echo "=== ${DIR_NAME} ==="
  echo ""
  echo "$EXTRACTED"
fi
