#!/usr/bin/env bash
# Smoke tests for: mxreq project (list|get|create|delete|tree|access|audit|hide|unhide)
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/helpers.sh"

resolve_bin "${1:-}"

echo "Running project tests against $BIN ..."

# ---------------------------------------------------------------------------
# Help tests (offline)
# ---------------------------------------------------------------------------

assert_output_contains \
  "project --help shows Available Commands" \
  "Available Commands" \
  "$BIN" project --help

assert_output_contains \
  "project list --help shows List all projects" \
  "List all projects" \
  "$BIN" project list --help

assert_output_contains \
  "project get --help shows [project]" \
  "[project]" \
  "$BIN" project get --help

assert_output_contains \
  "project create --help shows <label> <shortLabel>" \
  "<label>" \
  "$BIN" project create --help

assert_output_contains \
  "project delete --help shows <project>" \
  "<project>" \
  "$BIN" project delete --help

assert_output_contains \
  "project tree --help shows --filter" \
  "--filter" \
  "$BIN" project tree --help

assert_output_contains \
  "project audit --help shows --start" \
  "--start" \
  "$BIN" project audit --help

assert_output_contains \
  "project audit --help shows --max" \
  "--max" \
  "$BIN" project audit --help

assert_output_contains \
  "project hide --help shows --reason" \
  "--reason" \
  "$BIN" project hide --help

assert_output_contains \
  "project hide --help shows <project>" \
  "<project>" \
  "$BIN" project hide --help

assert_output_contains \
  "project unhide --help shows --reason" \
  "--reason" \
  "$BIN" project unhide --help

assert_output_contains \
  "project unhide --help shows <project> <newShortLabel>" \
  "<newShortLabel>" \
  "$BIN" project unhide --help

assert \
  "proj alias works" \
  "$BIN" proj --help

# ---------------------------------------------------------------------------
# Validation tests (offline)
# ---------------------------------------------------------------------------

assert_fail \
  "project create rejects missing args" \
  "$BIN" project create

assert_fail \
  "project create rejects single arg" \
  "$BIN" project create "OnlyLabel"

assert_fail \
  "project delete rejects missing args" \
  "$BIN" project delete

assert_fail \
  "project hide rejects missing args" \
  "$BIN" project hide

assert_fail \
  "project hide rejects missing --reason" \
  "$BIN" project hide FAKEPROJ

assert_fail \
  "project unhide rejects missing args" \
  "$BIN" project unhide

assert_fail \
  "project unhide rejects single arg" \
  "$BIN" project unhide FAKEPROJ

assert_fail \
  "project unhide rejects missing --reason" \
  "$BIN" project unhide FAKEPROJ NEWLABEL

# ---------------------------------------------------------------------------
# Live tests — CRUD lifecycle
# ---------------------------------------------------------------------------

if ! skip_live; then
  echo "Running project live tests ..."

  PROJ="${MXREQ_TEST_PROJECT:-SW_Sandbox}"

  # Read-only tests (no admin required)
  assert \
    "project list succeeds" \
    "$BIN" project list

  assert \
    "project list -o json succeeds" \
    "$BIN" project list -o json

  assert \
    "project get succeeds" \
    "$BIN" project get "$PROJ"

  assert \
    "project get -o json succeeds" \
    "$BIN" project get "$PROJ" -o json

  assert \
    "project tree succeeds" \
    "$BIN" project tree -p "$PROJ"

  assert \
    "project access succeeds" \
    "$BIN" project access -p "$PROJ"

  assert \
    "project audit succeeds" \
    "$BIN" project audit -p "$PROJ"

  # CRUD tests (require admin privileges)
  if ! skip_admin; then
    ID="$(unique_id)"
    PROJ_SHORT="SMKPRJ${ID}"
    PROJ_LABEL="Smoke Test ${ID}"

    cleanup_project() {
      "$BIN" project delete "$PROJ_SHORT" >/dev/null 2>&1 || true
    }
    trap cleanup_project EXIT

    assert_output_contains \
      "project create succeeds" \
      "created" \
      "$BIN" project create "$PROJ_LABEL" "$PROJ_SHORT"

    assert_output_contains \
      "project list shows created project" \
      "$PROJ_SHORT" \
      "$BIN" project list

    assert_output_contains \
      "project get shows created project" \
      "$PROJ_LABEL" \
      "$BIN" project get "$PROJ_SHORT"

    assert_output_contains \
      "project delete succeeds" \
      "deleted" \
      "$BIN" project delete "$PROJ_SHORT"

    assert_fail \
      "project delete again fails" \
      "$BIN" project delete "$PROJ_SHORT"

    trap - EXIT
  fi
fi

# ---------------------------------------------------------------------------
print_report
