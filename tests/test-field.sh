#!/usr/bin/env bash
# Smoke tests for: mxreq field (get|add|update|delete)
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/helpers.sh"

resolve_bin "${1:-}"

echo "Running field tests against $BIN ..."

# ---------------------------------------------------------------------------
# Help tests (offline)
# ---------------------------------------------------------------------------

assert_output_contains \
  "field --help shows Available Commands" \
  "Available Commands" \
  "$BIN" field --help

assert_output_contains \
  "field get --help shows <itemRef>" \
  "<itemRef>" \
  "$BIN" field get --help

assert_output_contains \
  "field get --help shows <fieldName>" \
  "<fieldName>" \
  "$BIN" field get --help

assert_output_contains \
  "field add --help shows --label" \
  "--label" \
  "$BIN" field add --help

assert_output_contains \
  "field add --help shows --type" \
  "--type" \
  "$BIN" field add --help

assert_output_contains \
  "field add --help shows --param" \
  "--param" \
  "$BIN" field add --help

assert_output_contains \
  "field add --help shows --reason" \
  "--reason" \
  "$BIN" field add --help

assert_output_contains \
  "field update --help shows --field-id" \
  "--field-id" \
  "$BIN" field update --help

assert_output_contains \
  "field update --help shows --reason" \
  "--reason" \
  "$BIN" field update --help

assert_output_contains \
  "field delete --help shows --field-id" \
  "--field-id" \
  "$BIN" field delete --help

# ---------------------------------------------------------------------------
# Validation tests (offline)
# ---------------------------------------------------------------------------

assert_fail \
  "field get rejects missing args" \
  "$BIN" field get

assert_fail \
  "field get rejects single arg" \
  "$BIN" field get REQ-1

assert_fail \
  "field add rejects missing args" \
  "$BIN" field add

assert_fail \
  "field add rejects missing --label" \
  "$BIN" field add FAKECAT -r "test" --type text --param ""

assert_fail \
  "field add rejects missing --type" \
  "$BIN" field add FAKECAT -r "test" --label "Test" --param ""

assert_fail \
  "field add rejects missing --param" \
  "$BIN" field add FAKECAT -r "test" --label "Test" --type text

assert_fail \
  "field add rejects missing --reason" \
  "$BIN" field add FAKECAT --label "Test" --type text --param ""

assert_fail \
  "field update rejects missing --field-id" \
  "$BIN" field update -r "test"

assert_fail \
  "field update rejects missing --reason" \
  "$BIN" field update --field-id 1

assert_fail \
  "field delete rejects missing args" \
  "$BIN" field delete

# ---------------------------------------------------------------------------
print_report
