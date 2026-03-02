#!/usr/bin/env bash
# Smoke tests for: mxreq export + import
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/helpers.sh"

resolve_bin "${1:-}"

echo "Running export/import tests against $BIN ..."

# ---------------------------------------------------------------------------
# Help tests (offline)
# ---------------------------------------------------------------------------

assert_output_contains \
  "export --help shows <itemList>" \
  "<itemList>" \
  "$BIN" export --help

assert_output_contains \
  "import --help shows <file>" \
  "<file>" \
  "$BIN" import --help

assert_output_contains \
  "import --help shows --reason" \
  "--reason" \
  "$BIN" import --help

# ---------------------------------------------------------------------------
# Validation tests (offline)
# ---------------------------------------------------------------------------

assert_fail \
  "export rejects missing args" \
  "$BIN" export

assert_fail \
  "import rejects missing args" \
  "$BIN" import

assert_fail \
  "import rejects missing --reason" \
  "$BIN" import /tmp/fakefile.xml

# ---------------------------------------------------------------------------
print_report
