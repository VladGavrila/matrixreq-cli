#!/usr/bin/env bash
# Smoke tests for: mxreq file (list|get|upload)
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/helpers.sh"

resolve_bin "${1:-}"

echo "Running file tests against $BIN ..."

# ---------------------------------------------------------------------------
# Help tests (offline)
# ---------------------------------------------------------------------------

assert_output_contains \
  "file --help shows Available Commands" \
  "Available Commands" \
  "$BIN" file --help

assert_output_contains \
  "file list --help shows List" \
  "List" \
  "$BIN" file list --help

assert_output_contains \
  "file get --help shows <fileNo>" \
  "<fileNo>" \
  "$BIN" file get --help

assert_output_contains \
  "file get --help shows --out" \
  "--out" \
  "$BIN" file get --help

assert_output_contains \
  "file upload --help shows <filePath>" \
  "<filePath>" \
  "$BIN" file upload --help

# ---------------------------------------------------------------------------
# Validation tests (offline)
# ---------------------------------------------------------------------------

assert_fail \
  "file get rejects missing args" \
  "$BIN" file get

assert_fail \
  "file get rejects single arg" \
  "$BIN" file get 1

assert_fail \
  "file upload rejects missing args" \
  "$BIN" file upload

# ---------------------------------------------------------------------------
# Live tests
# ---------------------------------------------------------------------------

if ! skip_live; then
  echo "Running file live tests ..."

  PROJ="${MXREQ_TEST_PROJECT:?Set MXREQ_TEST_PROJECT}"

  assert \
    "file list succeeds" \
    "$BIN" file list -p "$PROJ"
fi

# ---------------------------------------------------------------------------
print_report
