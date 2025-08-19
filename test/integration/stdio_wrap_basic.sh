#!/usr/bin/env bash
set -euo pipefail

# Purpose: Run mcp-test-server (stdio) through mcp-audit-wrap and assert audit JSONL fields.
# Runner-only: Designed to run on GitHub Actions ubuntu-latest (no Falco, no kernel deps).

ROOT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")"/../.. && pwd)"
BIN_WRAP="$ROOT_DIR/mcp-audit-wrap"
BIN_SERVER="$ROOT_DIR/mcp-test-server"
OUT_DIR="$ROOT_DIR/test-results"
AUDIT_FILE="$OUT_DIR/audit_stdio.jsonl"

mkdir -p "$OUT_DIR"

if ! command -v jq >/dev/null 2>&1; then
  echo "[SKIP] jq not installed"; exit 0; fi

if [[ ! -x "$BIN_WRAP" || ! -x "$BIN_SERVER" ]]; then
  echo "[SKIP] binaries missing: $BIN_WRAP or $BIN_SERVER"; exit 0; fi

rm -f "$AUDIT_FILE"

# Send a single JSON-RPC call to tools.exec with 2048 bytes payload response
printf '{"jsonrpc":"2.0","id":1,"method":"tools.exec"}\n' \
| "$BIN_WRAP" --sink "$AUDIT_FILE" --client-process ci -- \
    "$BIN_SERVER" --mode stdio --resp-size-bytes 2048 >/dev/null

if [[ ! -s "$AUDIT_FILE" ]]; then
  echo "[FAIL] audit file not created: $AUDIT_FILE" >&2; exit 1; fi

# Take last line (single event expected) and check basic fields
last_line=$(tail -n 1 "$AUDIT_FILE")

sess=$(jq -r '.session_id // empty' <<<"$last_line")
host=$(jq -r '.server_host // empty' <<<"$last_line")
tls=$(jq -r '.tls // empty' <<<"$last_line")
resp=$(jq -r '.response_bytes // 0' <<<"$last_line")
client=$(jq -r '.client_process // empty' <<<"$last_line")

[[ -n "$sess" ]] || { echo "[FAIL] missing session_id" >&2; exit 1; }
[[ "$host" == "stdio" ]] || { echo "[FAIL] unexpected server_host=$host" >&2; exit 1; }
[[ "$tls" == "false" ]] || { echo "[FAIL] tls must be false for stdio" >&2; exit 1; }
[[ "$client" == "ci" ]] || { echo "[FAIL] client_process mismatch: $client" >&2; exit 1; }

# response bytes should be >= requested size (stdout counted by wrap)
if (( resp < 2048 )); then
  echo "[FAIL] response_bytes too small: $resp" >&2; exit 1
fi

echo "[PASS] stdio_wrap_basic: session=$sess host=$host tls=$tls resp_bytes=$resp"

# Step summary (if available)
if [[ -n "${GITHUB_STEP_SUMMARY:-}" ]]; then
  {
    echo "### stdio_wrap_basic"
    echo "- session: $sess"
    echo "- host: $host"
    echo "- tls: $tls"
    echo "- response_bytes: $resp"
    echo "- audit: $(basename "$AUDIT_FILE")"
  } >> "$GITHUB_STEP_SUMMARY"
fi

exit 0

