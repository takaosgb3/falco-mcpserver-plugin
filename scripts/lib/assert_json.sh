#!/usr/bin/env bash
set -euo pipefail

# assert_json_field <json-file> <jq-filter> <expected>
assert_json_field() {
  local file="$1" filter="$2" expected="$3"
  local got
  got=$(jq -r "$filter" "$file") || { echo "[assert] jq failed: $file" >&2; return 2; }
  if [[ "$got" != "$expected" ]]; then
    echo "[assert] mismatch: $filter expected='$expected' got='$got' file=$file" >&2
    return 1
  fi
}

# assert_json_contains <json-file> <jq-filter> <substring>
assert_json_contains() {
  local file="$1" filter="$2" needle="$3"
  local got
  got=$(jq -r "$filter" "$file") || { echo "[assert] jq failed: $file" >&2; return 2; }
  if [[ "$got" != *"$needle"* ]]; then
    echo "[assert] not found: $filter ~ '$needle' in '$got' file=$file" >&2
    return 1
  fi
}

