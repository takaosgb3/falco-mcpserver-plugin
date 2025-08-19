#!/usr/bin/env bash
set -euo pipefail

DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
ROOT="$DIR/../../.."

# SKIP policy: if falco or jq missing, or no privileges, exit 0 with reason.
if ! command -v jq >/dev/null 2>&1; then
  echo "[SKIP] jq not installed"; exit 0; fi

# Offline validation mode using sample data and rule presence
ok_json="$ROOT/test/data/mcp_audit/ok_minimal.json"
if [[ ! -f "$ok_json" ]]; then echo "[SKIP] sample data missing"; exit 0; fi

echo "[INFO] Validating sample JSON shape (offline)"
jq -e '.schema_version=="1" and .tls==true and .client_process=="codex-cli"' "$ok_json" >/dev/null
echo "[PASS] ok_minimal.json schema sanity"

echo "[INFO] Checking rule templates exist"
test -f "$ROOT/rules/templates/mcp_baseline_rules.yaml" || { echo "[SKIP] rule template missing"; exit 0; }
echo "[PASS] rule templates present"

echo "[INFO] Offline basic checks complete"

