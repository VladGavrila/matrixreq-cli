#!/usr/bin/env bash
# Smoke tests for: mxreq todo (list|create|done)
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/helpers.sh"

resolve_bin "${1:-}"

echo "Running todo tests against $BIN ..."

# ---------------------------------------------------------------------------
# Help tests (offline)
# ---------------------------------------------------------------------------

assert_output_contains \
  "todo --help shows Available Commands" \
  "Available Commands" \
  "$BIN" todo --help

assert_output_contains \
  "todo list --help shows --done" \
  "--done" \
  "$BIN" todo list --help

assert_output_contains \
  "todo list --help shows --future" \
  "--future" \
  "$BIN" todo list --help

assert_output_contains \
  "todo list --help shows --all" \
  "--all" \
  "$BIN" todo list --help

assert_output_contains \
  "todo create --help shows <item>" \
  "<item>" \
  "$BIN" todo create --help

assert_output_contains \
  "todo create --help shows --text" \
  "--text" \
  "$BIN" todo create --help

assert_output_contains \
  "todo done --help shows <todoID>" \
  "<todoID>" \
  "$BIN" todo done --help

assert_output_contains \
  "todo done --help shows --hard-delete" \
  "--hard-delete" \
  "$BIN" todo done --help

# ---------------------------------------------------------------------------
# Validation tests (offline)
# ---------------------------------------------------------------------------

assert_fail \
  "todo create rejects missing args" \
  "$BIN" todo create

assert_fail \
  "todo create rejects missing --text" \
  "$BIN" todo create REQ-1

assert_fail \
  "todo done rejects missing args" \
  "$BIN" todo done

# ---------------------------------------------------------------------------
# Live tests
# ---------------------------------------------------------------------------

if ! skip_live; then
  echo "Running todo live tests ..."

  PROJ="${MXREQ_TEST_PROJECT:?Set MXREQ_TEST_PROJECT}"

  assert \
    "todo list --all succeeds" \
    "$BIN" todo list --all

  assert \
    "todo list -p project succeeds" \
    "$BIN" todo list -p "$PROJ"
fi

# ---------------------------------------------------------------------------
print_report
