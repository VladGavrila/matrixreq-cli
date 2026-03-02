#!/usr/bin/env bash
# Smoke tests for: mxreq report (generate|signed)
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/helpers.sh"

resolve_bin "${1:-}"

echo "Running report tests against $BIN ..."

# ---------------------------------------------------------------------------
# Help tests (offline)
# ---------------------------------------------------------------------------

assert_output_contains \
  "report --help shows Available Commands" \
  "Available Commands" \
  "$BIN" report --help

assert_output_contains \
  "report generate --help shows <report>" \
  "<report>" \
  "$BIN" report generate --help

assert_output_contains \
  "report generate --help shows --format" \
  "--format" \
  "$BIN" report generate --help

assert_output_contains \
  "report generate --help shows --signed" \
  "--signed" \
  "$BIN" report generate --help

assert_output_contains \
  "report signed --help shows <signItem>" \
  "<signItem>" \
  "$BIN" report signed --help

assert_output_contains \
  "report signed --help shows --format" \
  "--format" \
  "$BIN" report signed --help

# ---------------------------------------------------------------------------
# Validation tests (offline)
# ---------------------------------------------------------------------------

assert_fail \
  "report generate rejects missing args" \
  "$BIN" report generate

assert_fail \
  "report signed rejects missing args" \
  "$BIN" report signed

# ---------------------------------------------------------------------------
print_report
