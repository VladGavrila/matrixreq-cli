#!/usr/bin/env bash
# Smoke tests for: mxreq admin status|license|monitor|settings
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/helpers.sh"

resolve_bin "${1:-}"

echo "Running admin tests against $BIN ..."

# ---------------------------------------------------------------------------
# Help tests (offline)
# ---------------------------------------------------------------------------

assert_output_contains \
  "admin --help shows Available Commands" \
  "Available Commands" \
  "$BIN" admin --help

assert_output_contains \
  "admin status --help shows instance status" \
  "status" \
  "$BIN" admin status --help

assert_output_contains \
  "admin license --help shows license" \
  "license" \
  "$BIN" admin license --help

assert_output_contains \
  "admin monitor --help shows monitoring" \
  "monitor" \
  "$BIN" admin monitor --help

assert_output_contains \
  "admin settings --help shows --set-key" \
  "--set-key" \
  "$BIN" admin settings --help

assert_output_contains \
  "admin settings --help shows --set-value" \
  "--set-value" \
  "$BIN" admin settings --help

# ---------------------------------------------------------------------------
# Live tests
# ---------------------------------------------------------------------------

if ! skip_live; then
  echo "Running admin live tests ..."

  assert_output_contains \
    "admin status succeeds" \
    "Version" \
    "$BIN" admin status

  assert \
    "admin status -o json succeeds" \
    "$BIN" admin status -o json

  assert \
    "admin license succeeds" \
    "$BIN" admin license

  if ! skip_admin; then
    assert \
      "admin monitor succeeds" \
      "$BIN" admin monitor
  fi

  assert \
    "admin settings succeeds" \
    "$BIN" admin settings
fi

# ---------------------------------------------------------------------------
print_report
