#!/usr/bin/env bash
# Smoke tests for: mxreq branch (create|clone|merge|info|history)
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/helpers.sh"

resolve_bin "${1:-}"

echo "Running branch tests against $BIN ..."

# ---------------------------------------------------------------------------
# Help tests (offline)
# ---------------------------------------------------------------------------

assert_output_contains \
  "branch --help shows Available Commands" \
  "Available Commands" \
  "$BIN" branch --help

assert_output_contains \
  "branch create --help shows <label>" \
  "<label>" \
  "$BIN" branch create --help

assert_output_contains \
  "branch create --help shows --tag" \
  "--tag" \
  "$BIN" branch create --help

assert_output_contains \
  "branch create --help shows --keep-permissions" \
  "--keep-permissions" \
  "$BIN" branch create --help

assert_output_contains \
  "branch clone --help shows <label>" \
  "<label>" \
  "$BIN" branch clone --help

assert_output_contains \
  "branch clone --help shows --keep-history" \
  "--keep-history" \
  "$BIN" branch clone --help

assert_output_contains \
  "branch merge --help shows <branchProject>" \
  "<branchProject>" \
  "$BIN" branch merge --help

assert_output_contains \
  "branch merge --help shows --reason" \
  "--reason" \
  "$BIN" branch merge --help

assert_output_contains \
  "branch merge --help shows --direction" \
  "--direction" \
  "$BIN" branch merge --help

assert_output_contains \
  "branch info --help shows info" \
  "info" \
  "$BIN" branch info --help

assert_output_contains \
  "branch history --help shows history" \
  "history" \
  "$BIN" branch history --help

# ---------------------------------------------------------------------------
# Validation tests (offline)
# ---------------------------------------------------------------------------

assert_fail \
  "branch create rejects missing args" \
  "$BIN" branch create

assert_fail \
  "branch create rejects single arg" \
  "$BIN" branch create "OnlyLabel"

assert_fail \
  "branch clone rejects missing args" \
  "$BIN" branch clone

assert_fail \
  "branch clone rejects single arg" \
  "$BIN" branch clone "OnlyLabel"

assert_fail \
  "branch merge rejects missing args" \
  "$BIN" branch merge

# ---------------------------------------------------------------------------
print_report
