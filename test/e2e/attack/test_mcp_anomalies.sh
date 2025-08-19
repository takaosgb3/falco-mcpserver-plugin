#!/usr/bin/env bash
set -euo pipefail

DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
ROOT="$DIR/../../.."
LIB="$ROOT/scripts/lib/assert_json.sh"

if ! command -v jq >/dev/null 2>&1; then echo "[SKIP] jq not installed"; exit 0; fi

if [[ ! -f "$LIB" ]]; then echo "[SKIP] assert lib missing"; exit 0; fi
source "$LIB"

unapproved="$ROOT/test/data/mcp_audit/violation_unapproved_host.json"
insecure="$ROOT/test/data/mcp_audit/violation_insecure_tls.json"

if [[ ! -f "$unapproved" || ! -f "$insecure" ]]; then
  echo "[SKIP] sample anomalies missing"; exit 0
fi

echo "[INFO] Offline anomaly samples present; asserting fields"
assert_json_field "$unapproved" '.server_host' 'unknown.example.com'
assert_json_field "$insecure" '.tls' 'false'
echo "[PASS] anomaly samples verified"

# Note: Full Falco-based detection assertions will be added when plugin wiring is ready.

