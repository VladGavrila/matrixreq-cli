#!/usr/bin/env bash
# Smoke tests for: mxreq item (get|create|update|delete|restore|copy|move|touch)
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/helpers.sh"

resolve_bin "${1:-}"

echo "Running item tests against $BIN ..."

# ---------------------------------------------------------------------------
# Help tests (offline)
# ---------------------------------------------------------------------------

assert_output_contains \
  "item --help shows Available Commands" \
  "Available Commands" \
  "$BIN" item --help

assert_output_contains \
  "item get --help shows <itemRef>" \
  "<itemRef>" \
  "$BIN" item get --help

assert_output_contains \
  "item get --help shows --history" \
  "--history" \
  "$BIN" item get --help

assert_output_contains \
  "item create --help shows --title" \
  "--title" \
  "$BIN" item create --help

assert_output_contains \
  "item create --help shows --folder" \
  "--folder" \
  "$BIN" item create --help

assert_output_contains \
  "item create --help shows --reason" \
  "--reason" \
  "$BIN" item create --help

assert_output_contains \
  "item update --help shows <itemRef>" \
  "<itemRef>" \
  "$BIN" item update --help

assert_output_contains \
  "item update --help shows --title" \
  "--title" \
  "$BIN" item update --help

assert_output_contains \
  "item update --help shows --reason" \
  "--reason" \
  "$BIN" item update --help

assert_output_contains \
  "item delete --help shows --reason" \
  "--reason" \
  "$BIN" item delete --help

assert_output_contains \
  "item restore --help shows --reason" \
  "--reason" \
  "$BIN" item restore --help

assert_output_contains \
  "item copy --help shows <targetFolder>" \
  "<targetFolder>" \
  "$BIN" item copy --help

assert_output_contains \
  "item move --help shows --reason" \
  "--reason" \
  "$BIN" item move --help

assert_output_contains \
  "item touch --help shows --reason" \
  "--reason" \
  "$BIN" item touch --help

# ---------------------------------------------------------------------------
# Validation tests (offline)
# ---------------------------------------------------------------------------

assert_fail \
  "item get rejects missing args" \
  "$BIN" item get

assert_fail \
  "item create rejects missing --title" \
  "$BIN" item create --folder F-REQ-1 -r "test"

assert_fail \
  "item create rejects missing --folder" \
  "$BIN" item create --title "Test" -r "test"

assert_fail \
  "item create rejects missing --reason" \
  "$BIN" item create --title "Test" --folder F-REQ-1

assert_fail \
  "item update rejects missing args" \
  "$BIN" item update

assert_output_contains \
  "item update --help shows --field" \
  "--field" \
  "$BIN" item update --help

assert_fail \
  "item delete rejects missing args" \
  "$BIN" item delete

assert_fail \
  "item delete rejects missing --reason" \
  "$BIN" item delete REQ-1

assert_fail \
  "item restore rejects missing args" \
  "$BIN" item restore

assert_fail \
  "item restore rejects missing --reason" \
  "$BIN" item restore REQ-1

assert_fail \
  "item copy rejects missing args" \
  "$BIN" item copy

assert_fail \
  "item copy rejects single arg" \
  "$BIN" item copy REQ-1

assert_fail \
  "item move rejects missing args" \
  "$BIN" item move

assert_fail \
  "item touch rejects missing args" \
  "$BIN" item touch

# ---------------------------------------------------------------------------
# Live tests — CRUD lifecycle
# ---------------------------------------------------------------------------

if ! skip_live; then
  echo "Running item live tests ..."

  PROJ="${MXREQ_TEST_PROJECT:?Set MXREQ_TEST_PROJECT}"
  FOLDER="${MXREQ_TEST_FOLDER:?Set MXREQ_TEST_FOLDER}"
  ID="$(unique_id)"
  ITEM_TITLE="SmokeItem${ID}"
  ITEM_REF=""

  cleanup_item() {
    if [[ -n "$ITEM_REF" ]]; then
      "$BIN" item delete "$ITEM_REF" -p "$PROJ" -r "smoke cleanup" >/dev/null 2>&1 || true
    fi
  }
  trap cleanup_item EXIT

  # Derive category from folder ref (e.g., F-SOFT-740 → SOFT)
  CATEGORY=$(echo "$FOLDER" | sed 's/^F-//; s/-[0-9]*$//')

  # Create — output: "Created item ID=<id>, serial=<serial>"
  CREATE_OUT=$("$BIN" item create -p "$PROJ" --folder "$FOLDER" --title "$ITEM_TITLE" -r "smoke test" 2>&1) || true
  if echo "$CREATE_OUT" | grep -qF "Created item"; then
    RESULTS+=("PASS  item create succeeds")
    ((PASS++))
    # Extract serial and build item ref (e.g., SOFT-3362)
    SERIAL=$(echo "$CREATE_OUT" | grep -oE 'serial=[0-9]+' | grep -oE '[0-9]+') || true
    if [[ -n "$SERIAL" ]]; then
      ITEM_REF="${CATEGORY}-${SERIAL}"
    fi
  else
    RESULTS+=("FAIL  item create succeeds  (expected 'Created item' in output)")
    ((FAIL++))
  fi

  if [[ -n "$ITEM_REF" ]]; then
    assert_output_contains \
      "item get shows title" \
      "$ITEM_TITLE" \
      "$BIN" item get "$ITEM_REF" -p "$PROJ"

    assert \
      "item get -o json succeeds" \
      "$BIN" item get "$ITEM_REF" -p "$PROJ" -o json

    assert_output_contains \
      "item update succeeds" \
      "Updated" \
      "$BIN" item update "$ITEM_REF" -p "$PROJ" --title "Updated${ITEM_TITLE}" -r "smoke update"

    assert_output_contains \
      "item touch succeeds" \
      "Touched" \
      "$BIN" item touch "$ITEM_REF" -p "$PROJ" -r "smoke touch"

    assert_output_contains \
      "item delete succeeds" \
      "Deleted" \
      "$BIN" item delete "$ITEM_REF" -p "$PROJ" -r "smoke delete"

    assert_output_contains \
      "item restore succeeds" \
      "Restored" \
      "$BIN" item restore "$ITEM_REF" -p "$PROJ" -r "smoke restore"

    # Final cleanup delete
    assert_output_contains \
      "item delete (cleanup) succeeds" \
      "Deleted" \
      "$BIN" item delete "$ITEM_REF" -p "$PROJ" -r "smoke cleanup"

    ITEM_REF=""  # already cleaned up
  fi

  trap - EXIT
fi

# ---------------------------------------------------------------------------
print_report
