#!/usr/bin/env bash
# Smoke tests for: mxreq sign (create)
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/helpers.sh"

resolve_bin "${1:-}"

echo "Running sign tests against $BIN ..."

# ---------------------------------------------------------------------------
# Help tests (offline)
# ---------------------------------------------------------------------------

assert_output_contains \
  "sign --help shows Available Commands" \
  "Available Commands" \
  "$BIN" sign --help

assert_output_contains \
  "sign create --help shows <item>" \
  "<item>" \
  "$BIN" sign create --help

assert_output_contains \
  "sign create --help shows --password" \
  "--password" \
  "$BIN" sign create --help

assert_output_contains \
  "sign create --help shows --accept-comments" \
  "--accept-comments" \
  "$BIN" sign create --help

# ---------------------------------------------------------------------------
# Validation tests (offline)
# ---------------------------------------------------------------------------

assert_fail \
  "sign create rejects missing args" \
  "$BIN" sign create

assert_fail \
  "sign create rejects missing --password" \
  "$BIN" sign create SIGN-1

# ---------------------------------------------------------------------------
print_report
