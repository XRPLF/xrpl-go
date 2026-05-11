#!/bin/bash
# list-xls.sh — List all available XLS standards from XRPLF/XRPL-Standards
#
# Usage:
#   bash list-xls.sh [--json]
#
# Options:
#   --json   Output machine-readable JSON array

set -e

REPO="XRPLF/XRPL-Standards"
API_BASE="https://api.github.com/repos/${REPO}/contents"

JSON_MODE=false
for arg in "$@"; do
  case "$arg" in
    --json) JSON_MODE=true ;;
    *) echo "Unknown argument: $arg" >&2; exit 1 ;;
  esac
done

echo "Fetching XLS standards list..." >&2

API_RESPONSE=$(curl -sf \
  -H "Accept: application/vnd.github.v3+json" \
  "${API_BASE}/" 2>/dev/null) || {
  echo "Error: Failed to reach GitHub API. Check your network connection." >&2
  exit 1
}

if $JSON_MODE; then
  echo "$API_RESPONSE" | python3 -c "
import sys, json, re
data = json.load(sys.stdin)
pattern = re.compile(r'^XLS-(\d+)', re.IGNORECASE)
results = []
for item in data:
    name = item.get('name', '')
    m = pattern.match(name)
    if m and item.get('type') == 'dir':
        results.append({'number': int(m.group(1)), 'name': name})
results.sort(key=lambda x: x['number'])
print(json.dumps(results, indent=2))
"
else
  echo "$API_RESPONSE" | python3 -c "
import sys, json, re
data = json.load(sys.stdin)
pattern = re.compile(r'^XLS-(\d+)-?(.*)', re.IGNORECASE)
rows = []
for item in data:
    name = item.get('name', '')
    m = pattern.match(name)
    if m and item.get('type') == 'dir':
        num = int(m.group(1))
        slug = m.group(2)
        rows.append((num, slug, name))
rows.sort(key=lambda x: x[0])
print(f'{'NUM':<6}  {'SLUG':<40}  DIRECTORY')
print('-' * 80)
for num, slug, name in rows:
    print(f'{num:<6}  {slug:<40}  {name}')
print()
print(f'Total: {len(rows)} standards')
"
fi
