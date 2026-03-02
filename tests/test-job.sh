#!/usr/bin/env bash
# Smoke tests for: mxreq job (list|get|cancel|download)
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/helpers.sh"

resolve_bin "${1:-}"

echo "Running job tests against $BIN ..."

# ---------------------------------------------------------------------------
# Help tests (offline)
# ---------------------------------------------------------------------------

assert_output_contains \
  "job --help shows Available Commands" \
  "Available Commands" \
  "$BIN" job --help

assert_output_contains \
  "job list --help shows List" \
  "List" \
  "$BIN" job list --help

assert_output_contains \
  "job get --help shows <jobID>" \
  "<jobID>" \
  "$BIN" job get --help

assert_output_contains \
  "job cancel --help shows --reason" \
  "--reason" \
  "$BIN" job cancel --help

assert_output_contains \
  "job download --help shows <jobID>" \
  "<jobID>" \
  "$BIN" job download --help

assert_output_contains \
  "job download --help shows --out" \
  "--out" \
  "$BIN" job download --help

# ---------------------------------------------------------------------------
# Validation tests (offline)
# ---------------------------------------------------------------------------

assert_fail \
  "job get rejects missing args" \
  "$BIN" job get

assert_fail \
  "job cancel rejects missing args" \
  "$BIN" job cancel

assert_fail \
  "job cancel rejects missing --reason" \
  "$BIN" job cancel 1

assert_fail \
  "job download rejects missing args" \
  "$BIN" job download

assert_fail \
  "job download rejects single arg" \
  "$BIN" job download 1

# ---------------------------------------------------------------------------
# Live tests
# ---------------------------------------------------------------------------

if ! skip_live; then
  echo "Running job live tests ..."

  PROJ="${MXREQ_TEST_PROJECT:?Set MXREQ_TEST_PROJECT}"

  assert \
    "job list succeeds" \
    "$BIN" job list -p "$PROJ"
fi

# ---------------------------------------------------------------------------
print_report
