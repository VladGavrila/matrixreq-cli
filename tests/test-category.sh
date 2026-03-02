#!/usr/bin/env bash
# Smoke tests for: mxreq category (list|get|create|update|delete|settings)
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/helpers.sh"

resolve_bin "${1:-}"

echo "Running category tests against $BIN ..."

# ---------------------------------------------------------------------------
# Help tests (offline)
# ---------------------------------------------------------------------------

assert_output_contains \
  "category --help shows Available Commands" \
  "Available Commands" \
  "$BIN" category --help

assert \
  "cat alias works" \
  "$BIN" cat --help

assert_output_contains \
  "category list --help shows List" \
  "List" \
  "$BIN" category list --help

assert_output_contains \
  "category get --help shows <category>" \
  "<category>" \
  "$BIN" category get --help

assert_output_contains \
  "category create --help shows <label>" \
  "<label>" \
  "$BIN" category create --help

assert_output_contains \
  "category create --help shows --reason" \
  "--reason" \
  "$BIN" category create --help

assert_output_contains \
  "category update --help shows --label" \
  "--label" \
  "$BIN" category update --help

assert_output_contains \
  "category update --help shows --reason" \
  "--reason" \
  "$BIN" category update --help

assert_output_contains \
  "category delete --help shows --reason" \
  "--reason" \
  "$BIN" category delete --help

assert_output_contains \
  "category settings --help shows <category>" \
  "<category>" \
  "$BIN" category settings --help

assert_output_contains \
  "category update --help shows --order" \
  "--order" \
  "$BIN" category update --help

# ---------------------------------------------------------------------------
# Validation tests (offline)
# ---------------------------------------------------------------------------

assert_fail \
  "category get rejects missing args" \
  "$BIN" category get

assert_fail \
  "category create rejects missing args" \
  "$BIN" category create

assert_fail \
  "category create rejects single arg" \
  "$BIN" category create "OnlyLabel"

assert_fail \
  "category create rejects missing --reason" \
  "$BIN" category create "Label" "SHORT"

assert_fail \
  "category update rejects missing args" \
  "$BIN" category update

assert_fail \
  "category update rejects missing --reason" \
  "$BIN" category update FAKECAT

assert_fail \
  "category delete rejects missing args" \
  "$BIN" category delete

assert_fail \
  "category delete rejects missing --reason" \
  "$BIN" category delete FAKECAT

assert_fail \
  "category settings rejects missing args" \
  "$BIN" category settings

# ---------------------------------------------------------------------------
# Live tests — CRUD lifecycle
# ---------------------------------------------------------------------------

if ! skip_live; then
  echo "Running category live tests ..."

  PROJ="${MXREQ_TEST_PROJECT:?Set MXREQ_TEST_PROJECT}"

  # Read-only tests
  assert \
    "category list succeeds" \
    "$BIN" category list -p "$PROJ"

  assert \
    "category settings succeeds" \
    "$BIN" category settings SOFT -p "$PROJ"

  # CRUD tests (require admin privileges)
  if ! skip_admin; then
    ID="$(unique_id)"
    CAT_SHORT="SMKCAT${ID}"
    CAT_LABEL="Smoke Cat ${ID}"

    cleanup_cat() {
      "$BIN" category delete "$CAT_SHORT" -p "$PROJ" -r "smoke cleanup" >/dev/null 2>&1 || true
    }
    trap cleanup_cat EXIT

    assert \
      "category list succeeds" \
      "$BIN" category list -p "$PROJ"

    assert_output_contains \
      "category create succeeds" \
      "created" \
      "$BIN" category create "$CAT_LABEL" "$CAT_SHORT" -p "$PROJ" -r "smoke test"

    assert_output_contains \
      "category get shows label" \
      "$CAT_LABEL" \
      "$BIN" category get "$CAT_SHORT" -p "$PROJ"

    assert_output_contains \
      "category update succeeds" \
      "updated" \
      "$BIN" category update "$CAT_SHORT" -p "$PROJ" --label "Updated ${CAT_LABEL}" -r "smoke update"

    assert_output_contains \
      "category delete succeeds" \
      "deleted" \
      "$BIN" category delete "$CAT_SHORT" -p "$PROJ" -r "smoke cleanup"

    trap - EXIT
  fi
fi

# ---------------------------------------------------------------------------
print_report
