#!/usr/bin/env bash
# Smoke tests for: mxreq root help + version command
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/helpers.sh"

resolve_bin "${1:-}"

echo "Running version/root tests against $BIN ..."

# ---------------------------------------------------------------------------
# Help / flag tests (offline)
# ---------------------------------------------------------------------------

assert_output_contains \
  "root --help shows Available Commands" \
  "Available Commands" \
  "$BIN" --help

assert_output_contains \
  "root --help shows project command" \
  "project" \
  "$BIN" --help

assert_output_contains \
  "root --help shows --url flag" \
  "--url" \
  "$BIN" --help

assert_output_contains \
  "root --help shows --token flag" \
  "--token" \
  "$BIN" --help

assert_output_contains \
  "root --help shows -o flag" \
  "-o" \
  "$BIN" --help

assert_output_contains \
  "root --help shows -p flag" \
  "-p" \
  "$BIN" --help

assert \
  "version exits 0" \
  "$BIN" version

assert_output_contains \
  "version output contains mxreq" \
  "mxreq" \
  "$BIN" version

# ---------------------------------------------------------------------------
# Validation (offline)
# ---------------------------------------------------------------------------

assert_fail \
  "unknown command fails" \
  "$BIN" nosuchcommand

# ---------------------------------------------------------------------------
print_report
