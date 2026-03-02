#!/usr/bin/env bash
# Shared test helpers for mxreq CLI smoke tests.
# Source this file from individual test scripts:
#   SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
#   source "$SCRIPT_DIR/helpers.sh"

PASS=0
FAIL=0
RESULTS=()

# ---------------------------------------------------------------------------
# Preflight
# ---------------------------------------------------------------------------

# resolve_bin sets BIN from the first argument or falls back to the default.
# Call once at the top of each test script.
resolve_bin() {
  BIN="${1:-./mxreq}"
  if [[ ! -x "$BIN" ]]; then
    echo "Binary not found: $BIN"
    echo "Run 'make build' first."
    exit 1
  fi
}

# ---------------------------------------------------------------------------
# Live-test gate
# ---------------------------------------------------------------------------

# skip_live returns 0 (true) when live tests are DISABLED, 1 when enabled.
# Usage: if skip_live; then return; fi
skip_live() {
  [[ "${MXREQ_LIVE_TESTS:-0}" != "1" ]]
}

# skip_admin returns 0 (true) when admin tests are DISABLED, 1 when enabled.
# Admin tests require elevated privileges (project/user/group create/delete).
# Usage: if skip_admin; then echo "skipping admin tests"; fi
skip_admin() {
  [[ "${MXREQ_ADMIN_TESTS:-0}" != "1" ]]
}

# ---------------------------------------------------------------------------
# Unique naming
# ---------------------------------------------------------------------------

# unique_id echoes a Unix epoch timestamp for unique resource naming.
unique_id() {
  date +%s
}

# ---------------------------------------------------------------------------
# Assertions
# ---------------------------------------------------------------------------

# assert NAME CMD [ARGS...] — pass if command exits 0.
assert() {
  local name="$1"
  shift
  if "$@" >/dev/null 2>&1; then
    RESULTS+=("PASS  $name")
    ((PASS++))
  else
    RESULTS+=("FAIL  $name")
    ((FAIL++))
  fi
}

# assert_fail NAME CMD [ARGS...] — pass if command exits non-zero.
assert_fail() {
  local name="$1"
  shift
  if "$@" >/dev/null 2>&1; then
    RESULTS+=("FAIL  $name")
    ((FAIL++))
  else
    RESULTS+=("PASS  $name")
    ((PASS++))
  fi
}

# assert_output_contains NAME PATTERN CMD [ARGS...] — pass if combined
# stdout+stderr contains PATTERN (fixed-string match).
assert_output_contains() {
  local name="$1"
  local pattern="$2"
  shift 2
  local out
  out=$("$@" 2>&1) || true
  if echo "$out" | grep -qF -- "$pattern"; then
    RESULTS+=("PASS  $name")
    ((PASS++))
  else
    RESULTS+=("FAIL  $name  (expected '$pattern' in output)")
    ((FAIL++))
  fi
}

# assert_output_not_contains NAME PATTERN CMD [ARGS...] — pass if combined
# stdout+stderr does NOT contain PATTERN (fixed-string match).
assert_output_not_contains() {
  local name="$1"
  local pattern="$2"
  shift 2
  local out
  out=$("$@" 2>&1) || true
  if echo "$out" | grep -qF -- "$pattern"; then
    RESULTS+=("FAIL  $name  (unexpected '$pattern' in output)")
    ((FAIL++))
  else
    RESULTS+=("PASS  $name")
    ((PASS++))
  fi
}

# ---------------------------------------------------------------------------
# Report
# ---------------------------------------------------------------------------

# print_report outputs the test summary and exits non-zero on any failure.
print_report() {
  echo ""
  echo "=============================="
  echo "  Test Report"
  echo "=============================="
  for r in "${RESULTS[@]}"; do
    echo "  $r"
  done
  echo "------------------------------"
  echo "  Total: $((PASS + FAIL))  Pass: $PASS  Fail: $FAIL"
  echo "=============================="

  if [[ $FAIL -gt 0 ]]; then
    exit 1
  fi
}
