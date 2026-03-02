#!/usr/bin/env bash
# Smoke tests for: mxreq search
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/helpers.sh"

resolve_bin "${1:-}"

echo "Running search tests against $BIN ..."

# ---------------------------------------------------------------------------
# Help tests (offline)
# ---------------------------------------------------------------------------

assert_output_contains \
  "search --help shows <query>" \
  "<query>" \
  "$BIN" search --help

assert_output_contains \
  "search --help shows --filter" \
  "--filter" \
  "$BIN" search --help

assert_output_contains \
  "search --help shows --minimal" \
  "--minimal" \
  "$BIN" search --help

# ---------------------------------------------------------------------------
# Validation tests (offline)
# ---------------------------------------------------------------------------

assert_fail \
  "search rejects missing args" \
  "$BIN" search

# ---------------------------------------------------------------------------
# Live tests
# ---------------------------------------------------------------------------

if ! skip_live; then
  echo "Running search live tests ..."

  PROJ="${MXREQ_TEST_PROJECT:?Set MXREQ_TEST_PROJECT}"

  assert \
    "search succeeds" \
    "$BIN" search "*" -p "$PROJ"

  assert \
    "search -o json succeeds" \
    "$BIN" search "*" -p "$PROJ" -o json

  assert \
    "search --minimal succeeds" \
    "$BIN" search "*" -p "$PROJ" --minimal

  assert \
    "search --filter REQ succeeds" \
    "$BIN" search "*" -p "$PROJ" --filter REQ
fi

# ---------------------------------------------------------------------------
print_report
