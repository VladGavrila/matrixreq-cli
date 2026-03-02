#!/usr/bin/env bash
# Run all mxreq CLI smoke tests.
# Usage:
#   ./tests/run-all.sh [binary]
#   MXREQ_LIVE_TESTS=1 ./tests/run-all.sh

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_ARG="${1:-}"

TOTAL=0
PASSED=0
FAILED=0
FAILED_SCRIPTS=()

echo "========================================"
echo "  mxreq CLI Smoke Tests"
echo "========================================"
if [[ "${MXREQ_LIVE_TESTS:-0}" == "1" ]]; then
  if [[ "${MXREQ_ADMIN_TESTS:-0}" == "1" ]]; then
    echo "  Mode: offline + live + admin"
  else
    echo "  Mode: offline + live (no admin CRUD)"
  fi
else
  echo "  Mode: offline only"
fi
echo ""

# ---------------------------------------------------------------------------
# Live-test folder provisioning
# ---------------------------------------------------------------------------
_CREATED_TEST_FOLDER=""

if [[ "${MXREQ_LIVE_TESTS:-0}" == "1" && -z "${MXREQ_TEST_FOLDER:-}" ]]; then
  _BIN="${BIN_ARG:-./mxreq}"
  _PROJ="${MXREQ_TEST_PROJECT:-}"

  if [[ -z "$_PROJ" ]]; then
    echo "ERROR: MXREQ_TEST_PROJECT must be set when MXREQ_LIVE_TESTS=1" >&2
    exit 1
  fi
  if [[ ! -x "$_BIN" ]]; then
    echo "ERROR: binary not found: $_BIN" >&2
    exit 1
  fi

  # Find a real root folder from the project tree
  ROOT_FOLDER=$("$_BIN" project tree -p "$_PROJ" -o json 2>/dev/null | grep -oE '"F-[A-Z]+-[0-9]+"' | head -1 | tr -d '"')
  if [[ -z "$ROOT_FOLDER" ]]; then
    echo "ERROR: could not find any folder in project $_PROJ tree" >&2
    exit 1
  fi
  FIRST_CAT=$(echo "$ROOT_FOLDER" | sed 's/^F-//; s/-[0-9]*$//')

  _ID="$(date +%s)"
  FOLDER_JSON=$("$_BIN" folder create -p "$_PROJ" --parent "$ROOT_FOLDER" \
    --label "SmokeTestFolder${_ID}" -r "smoke test setup" -o json 2>&1)
  FOLDER_SERIAL=$(echo "$FOLDER_JSON" | grep -o '"serial": *[0-9]*' | grep -o '[0-9]*')

  if [[ -z "$FOLDER_SERIAL" ]]; then
    echo "ERROR: failed to create smoke test folder: $FOLDER_JSON" >&2
    exit 1
  fi

  MXREQ_TEST_FOLDER="F-${FIRST_CAT}-${FOLDER_SERIAL}"
  export MXREQ_TEST_FOLDER
  _CREATED_TEST_FOLDER="$MXREQ_TEST_FOLDER"
  echo "Created smoke test folder: ${MXREQ_TEST_FOLDER}"
  echo ""
fi

cleanup_test_folder() {
  if [[ -n "$_CREATED_TEST_FOLDER" ]]; then
    echo ""
    echo "Cleaning up smoke test folder ${_CREATED_TEST_FOLDER} ..."
    "${BIN_ARG:-./mxreq}" item delete "$_CREATED_TEST_FOLDER" \
      -p "${MXREQ_TEST_PROJECT:-}" -r "smoke test cleanup" >/dev/null 2>&1 || true
  fi
}
trap cleanup_test_folder EXIT

for script in "$SCRIPT_DIR"/test-*.sh; do
  name="$(basename "$script")"
  ((TOTAL++))
  echo "--- $name ---"
  if bash "$script" "$BIN_ARG"; then
    ((PASSED++))
  else
    ((FAILED++))
    FAILED_SCRIPTS+=("$name")
  fi
  echo ""
done

echo "========================================"
echo "  Overall: $TOTAL scripts, $PASSED passed, $FAILED failed"
echo "========================================"

if [[ $FAILED -gt 0 ]]; then
  echo "  Failed scripts:"
  for s in "${FAILED_SCRIPTS[@]}"; do
    echo "    - $s"
  done
  exit 1
fi
