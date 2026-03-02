#!/usr/bin/env bash
# Smoke tests for: mxreq user (list|get|create|update|delete|rename|token|audit)
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/helpers.sh"

resolve_bin "${1:-}"

echo "Running user tests against $BIN ..."

# ---------------------------------------------------------------------------
# Help tests (offline)
# ---------------------------------------------------------------------------

assert_output_contains \
  "user --help shows Available Commands" \
  "Available Commands" \
  "$BIN" user --help

assert_output_contains \
  "user list --help shows List" \
  "List" \
  "$BIN" user list --help

assert_output_contains \
  "user list --help shows --details" \
  "--details" \
  "$BIN" user list --help

assert_output_contains \
  "user get --help shows <user>" \
  "<user>" \
  "$BIN" user get --help

assert_output_contains \
  "user create --help shows --login" \
  "--login" \
  "$BIN" user create --help

assert_output_contains \
  "user create --help shows --email" \
  "--email" \
  "$BIN" user create --help

assert_output_contains \
  "user create --help shows --password" \
  "--password" \
  "$BIN" user create --help

assert_output_contains \
  "user update --help shows <user>" \
  "<user>" \
  "$BIN" user update --help

assert_output_contains \
  "user delete --help shows <user>" \
  "<user>" \
  "$BIN" user delete --help

assert_output_contains \
  "user rename --help shows <newLogin>" \
  "<newLogin>" \
  "$BIN" user rename --help

assert_output_contains \
  "user token --help shows <user>" \
  "<user>" \
  "$BIN" user token --help

assert_output_contains \
  "user token --help shows --purpose" \
  "--purpose" \
  "$BIN" user token --help

assert_output_contains \
  "user audit --help shows <user>" \
  "<user>" \
  "$BIN" user audit --help

assert_output_contains \
  "user audit --help shows --start" \
  "--start" \
  "$BIN" user audit --help

assert_output_contains \
  "user audit --help shows --max" \
  "--max" \
  "$BIN" user audit --help

# ---------------------------------------------------------------------------
# Validation tests (offline)
# ---------------------------------------------------------------------------

assert_fail \
  "user create rejects missing --login" \
  "$BIN" user create --email "a@b.com" --password "pass"

assert_fail \
  "user create rejects missing --email" \
  "$BIN" user create --login "smokeuser" --password "pass"

assert_fail \
  "user create rejects missing --password" \
  "$BIN" user create --login "smokeuser" --email "a@b.com"

assert_fail \
  "user get rejects missing args" \
  "$BIN" user get

assert_fail \
  "user update rejects missing args" \
  "$BIN" user update

assert_fail \
  "user delete rejects missing args" \
  "$BIN" user delete

assert_fail \
  "user rename rejects missing args" \
  "$BIN" user rename

assert_fail \
  "user rename rejects single arg" \
  "$BIN" user rename olduser

assert_fail \
  "user token rejects missing args" \
  "$BIN" user token

assert_fail \
  "user token rejects missing --purpose" \
  "$BIN" user token someuser

assert_fail \
  "user audit rejects missing args" \
  "$BIN" user audit

# ---------------------------------------------------------------------------
# Live tests — CRUD lifecycle
# ---------------------------------------------------------------------------

if ! skip_live; then
  echo "Running user live tests ..."

  # Read-only tests (no admin required)
  assert \
    "user list succeeds" \
    "$BIN" user list

  assert \
    "user list -o json succeeds" \
    "$BIN" user list -o json

  # CRUD tests (require admin privileges)
  if ! skip_admin; then
    ID="$(unique_id)"
    USR_LOGIN="smoketest${ID}"
    USR_EMAIL="${USR_LOGIN}@smoketest.local"
    USR_RENAMED="smkrenamed${ID}"

    cleanup_user() {
      "$BIN" user delete "$USR_LOGIN" >/dev/null 2>&1 || true
      "$BIN" user delete "$USR_RENAMED" >/dev/null 2>&1 || true
    }
    trap cleanup_user EXIT

    assert_output_contains \
      "user create succeeds" \
      "created" \
      "$BIN" user create --login "$USR_LOGIN" --email "$USR_EMAIL" --password "Sm0keP@ss!"

    assert_output_contains \
      "user list shows created user" \
      "$USR_LOGIN" \
      "$BIN" user list

    assert_output_contains \
      "user get shows email" \
      "$USR_EMAIL" \
      "$BIN" user get "$USR_LOGIN"

    assert_output_contains \
      "user update succeeds" \
      "updated" \
      "$BIN" user update "$USR_LOGIN" --first "Smoke" --last "Test"

    assert_output_contains \
      "user rename succeeds" \
      "renamed" \
      "$BIN" user rename "$USR_LOGIN" "$USR_RENAMED"

    assert_output_contains \
      "user token create succeeds" \
      "Token created" \
      "$BIN" user token "$USR_RENAMED" --purpose "smoke test token"

    assert \
      "user audit succeeds" \
      "$BIN" user audit "$USR_RENAMED"

    assert_output_contains \
      "user delete succeeds" \
      "deleted" \
      "$BIN" user delete "$USR_RENAMED"

    assert_fail \
      "user delete again fails" \
      "$BIN" user delete "$USR_RENAMED"

    trap - EXIT
  fi
fi

# ---------------------------------------------------------------------------
print_report
