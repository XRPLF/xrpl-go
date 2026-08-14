#!/bin/sh

set -eu

REPORT_PATH=${1:?Usage: summarize-go-test-report.sh REPORT_PATH TITLE}
TITLE=${2:?Usage: summarize-go-test-report.sh REPORT_PATH TITLE}

echo "## $TITLE"

if [ ! -s "$REPORT_PATH" ]; then
	echo
	echo "No structured test report was produced."
	exit 0
fi

echo
jq -rs '
  def count_action($action):
    [.[] | select(.Test != null and .Action == $action)] | length;
  "| Result | Tests |\n" +
  "| --- | ---: |\n" +
  "| Passed | \(count_action("pass")) |\n" +
  "| Failed | \(count_action("fail")) |\n" +
  "| Skipped | \(count_action("skip")) |"
' "$REPORT_PATH"

FAILED_TESTS=$(jq -rs '
  [
    .[]
    | select(.Action == "fail" and .Test != null)
    | "- `\(.Package)` — `\(.Test)`"
  ]
  | unique
  | .[]
' "$REPORT_PATH")

FAILED_PACKAGES=$(jq -rs '
  [
    .[]
    | select(.Action == "fail" and .Test == null and .Package != null)
    | "- `\(.Package)`"
  ]
  | unique
  | .[]
' "$REPORT_PATH")

if [ -n "$FAILED_TESTS" ]; then
	echo
	echo "### Failed tests"
	echo
	printf '%s\n' "$FAILED_TESTS"
fi

if [ -n "$FAILED_PACKAGES" ]; then
	echo
	echo "### Failed packages"
	echo
	printf '%s\n' "$FAILED_PACKAGES"
fi
