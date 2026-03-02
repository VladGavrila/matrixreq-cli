#!/usr/bin/env bash
# Smoke tests for: mxreq group (list|get|create|delete|add-user|remove-user)
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/helpers.sh"

resolve_bin "${1:-}"

echo "Running group tests against $BIN ..."

# ---------------------------------------------------------------------------
# Help tests (offline)
# ---------------------------------------------------------------------------

assert_output_contains \
  "group --help shows Available Commands" \
  "Available Commands" \
  "$BIN" group --help

assert_output_contains \
  "group list --help shows --details" \
  "--details" \
  "$BIN" group list --help

assert_output_contains \
  "group get --help shows <groupID>" \
  "<groupID>" \
  "$BIN" group get --help

assert_output_contains \
  "group create --help shows <name>" \
  "<name>" \
  "$BIN" group create --help

assert_output_contains \
  "group delete --help shows <groupID>" \
  "<groupID>" \
  "$BIN" group delete --help

assert_output_contains \
  "group add-user --help shows <groupID>" \
  "<groupID>" \
  "$BIN" group add-user --help

assert_output_contains \
  "group remove-user --help shows <groupName>" \
  "<groupName>" \
  "$BIN" group remove-user --help

# ---------------------------------------------------------------------------
# Validation tests (offline)
# ---------------------------------------------------------------------------

assert_fail \
  "group get rejects missing args" \
  "$BIN" group get

assert_fail \
  "group create rejects missing args" \
  "$BIN" group create

assert_fail \
  "group delete rejects missing args" \
  "$BIN" group delete

assert_fail \
  "group add-user rejects missing args" \
  "$BIN" group add-user

assert_fail \
  "group add-user rejects single arg" \
  "$BIN" group add-user 1

assert_fail \
  "group remove-user rejects missing args" \
  "$BIN" group remove-user

assert_fail \
  "group remove-user rejects single arg" \
  "$BIN" group remove-user somegroup

# ---------------------------------------------------------------------------
# Live tests — CRUD lifecycle
# ---------------------------------------------------------------------------

if ! skip_live; then
  echo "Running group live tests ..."

  # Read-only tests (no admin required)
  assert \
    "group list succeeds" \
    "$BIN" group list

  # CRUD tests (require admin privileges)
  if ! skip_admin; then
    ID="$(unique_id)"
    GRP_NAME="smokegroup${ID}"
    GRP_ID=""

    cleanup_group() {
      if [[ -n "$GRP_ID" ]]; then
        "$BIN" group delete "$GRP_ID" >/dev/null 2>&1 || true
      fi
    }
    trap cleanup_group EXIT

    # Create and capture group ID
    CREATE_OUT=$("$BIN" group create "$GRP_NAME" 2>&1) || true
    if echo "$CREATE_OUT" | grep -qF "created"; then
      RESULTS+=("PASS  group create succeeds")
      ((PASS++))
      GRP_ID=$(echo "$CREATE_OUT" | grep -oE '[0-9]+' | head -1) || true
    else
      RESULTS+=("FAIL  group create succeeds  (expected 'created' in output)")
      ((FAIL++))
    fi

    assert_output_contains \
      "group list shows created group" \
      "$GRP_NAME" \
      "$BIN" group list

    if [[ -n "$GRP_ID" ]]; then
      assert_output_contains \
        "group delete succeeds" \
        "deleted" \
        "$BIN" group delete "$GRP_ID"
      GRP_ID=""
    fi

    trap - EXIT
  fi
fi

# ---------------------------------------------------------------------------
print_report
