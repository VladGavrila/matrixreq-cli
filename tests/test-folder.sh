#!/usr/bin/env bash
# Smoke tests for: mxreq folder (get|create)
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/helpers.sh"

resolve_bin "${1:-}"

echo "Running folder tests against $BIN ..."

# ---------------------------------------------------------------------------
# Help tests (offline)
# ---------------------------------------------------------------------------

assert_output_contains \
  "folder --help shows Available Commands" \
  "Available Commands" \
  "$BIN" folder --help

assert_output_contains \
  "folder get --help shows <folderRef>" \
  "<folderRef>" \
  "$BIN" folder get --help

assert_output_contains \
  "folder get --help shows --history" \
  "--history" \
  "$BIN" folder get --help

assert_output_contains \
  "folder create --help shows --parent" \
  "--parent" \
  "$BIN" folder create --help

assert_output_contains \
  "folder create --help shows --label" \
  "--label" \
  "$BIN" folder create --help

assert_output_contains \
  "folder create --help shows --reason" \
  "--reason" \
  "$BIN" folder create --help

assert_output_contains \
  "folder update --help shows <folderRef>" \
  "<folderRef>" \
  "$BIN" folder update --help

assert_output_contains \
  "folder update --help shows --title" \
  "--title" \
  "$BIN" folder update --help

assert_output_contains \
  "folder update --help shows --description" \
  "--description" \
  "$BIN" folder update --help

assert_output_contains \
  "folder update --help shows --field" \
  "--field" \
  "$BIN" folder update --help

assert_output_contains \
  "folder update --help shows --reason" \
  "--reason" \
  "$BIN" folder update --help

# ---------------------------------------------------------------------------
# Validation tests (offline)
# ---------------------------------------------------------------------------

assert_fail \
  "folder get rejects missing args" \
  "$BIN" folder get

assert_fail \
  "folder create rejects missing --parent" \
  "$BIN" folder create --label "Test" -r "test"

assert_fail \
  "folder create rejects missing --label" \
  "$BIN" folder create --parent F-REQ-1 -r "test"

assert_fail \
  "folder create rejects missing --reason" \
  "$BIN" folder create --parent F-REQ-1 --label "Test"

assert_fail \
  "folder update rejects missing args" \
  "$BIN" folder update

# ---------------------------------------------------------------------------
# Live tests
# ---------------------------------------------------------------------------

if ! skip_live; then
  echo "Running folder live tests ..."

  PROJ="${MXREQ_TEST_PROJECT:?Set MXREQ_TEST_PROJECT}"
  PARENT="${MXREQ_TEST_FOLDER:?Set MXREQ_TEST_FOLDER}"
  ID="$(unique_id)"
  CATEGORY=$(echo "$PARENT" | sed 's/^F-//; s/-[0-9]*$//')
  SUB_FOLDER_REF=""

  cleanup_subfolder() {
    if [[ -n "$SUB_FOLDER_REF" ]]; then
      "$BIN" item delete "$SUB_FOLDER_REF" -p "$PROJ" -r "smoke cleanup" >/dev/null 2>&1 || true
    fi
  }
  trap cleanup_subfolder EXIT

  FOLDER_JSON=$("$BIN" folder create -p "$PROJ" --parent "$PARENT" \
    --label "SmokeFolder${ID}" -r "smoke test" -o json 2>&1) || true
  FOLDER_SERIAL=$(echo "$FOLDER_JSON" | grep -o '"serial": *[0-9]*' | grep -o '[0-9]*')
  if [[ -n "$FOLDER_SERIAL" ]]; then
    RESULTS+=("PASS  folder create succeeds")
    ((PASS++))
    SUB_FOLDER_REF="F-${CATEGORY}-${FOLDER_SERIAL}"
  else
    RESULTS+=("FAIL  folder create succeeds  (expected serial in json output)")
    ((FAIL++))
  fi

  assert \
    "folder get succeeds" \
    "$BIN" folder get "$PARENT" -p "$PROJ"

  if [[ -n "$SUB_FOLDER_REF" ]]; then
    assert_output_contains \
      "folder update title succeeds" \
      "Updated" \
      "$BIN" folder update "$SUB_FOLDER_REF" -p "$PROJ" --title "UpdatedFolder${ID}" -r "smoke update"

    assert_output_contains \
      "folder update -o json succeeds" \
      "itemRef" \
      "$BIN" folder update "$SUB_FOLDER_REF" -p "$PROJ" --title "FinalFolder${ID}" -r "smoke update json" -o json
  fi

  trap - EXIT
  cleanup_subfolder
fi

# ---------------------------------------------------------------------------
print_report
