#!/usr/bin/env bash
# Smoke tests for: mxreq link (create|delete)
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/helpers.sh"

resolve_bin "${1:-}"

echo "Running link tests against $BIN ..."

# ---------------------------------------------------------------------------
# Help tests (offline)
# ---------------------------------------------------------------------------

assert_output_contains \
  "link --help shows Available Commands" \
  "Available Commands" \
  "$BIN" link --help

assert_output_contains \
  "link create --help shows <upItem>" \
  "<upItem>" \
  "$BIN" link create --help

assert_output_contains \
  "link create --help shows --reason" \
  "--reason" \
  "$BIN" link create --help

assert_output_contains \
  "link delete --help shows <upItem>" \
  "<upItem>" \
  "$BIN" link delete --help

assert_output_contains \
  "link delete --help shows --reason" \
  "--reason" \
  "$BIN" link delete --help

# ---------------------------------------------------------------------------
# Validation tests (offline)
# ---------------------------------------------------------------------------

assert_fail \
  "link create rejects missing args" \
  "$BIN" link create

assert_fail \
  "link create rejects single arg" \
  "$BIN" link create REQ-1

assert_fail \
  "link create rejects missing --reason" \
  "$BIN" link create REQ-1 SPEC-1

assert_fail \
  "link delete rejects missing args" \
  "$BIN" link delete

assert_fail \
  "link delete rejects single arg" \
  "$BIN" link delete REQ-1

# ---------------------------------------------------------------------------
print_report
