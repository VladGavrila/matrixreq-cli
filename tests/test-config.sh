#!/usr/bin/env bash
# Smoke tests for: mxreq config init|show
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/helpers.sh"

resolve_bin "${1:-}"

echo "Running config tests against $BIN ..."

# ---------------------------------------------------------------------------
# Help tests (offline)
# ---------------------------------------------------------------------------

assert_output_contains \
  "config --help shows Available Commands" \
  "Available Commands" \
  "$BIN" config --help

assert_output_contains \
  "config init --help shows Initialize" \
  "Initialize" \
  "$BIN" config init --help

assert_output_contains \
  "config show --help shows Show" \
  "Show" \
  "$BIN" config show --help

# ---------------------------------------------------------------------------
print_report
