#!/usr/bin/env bash
# Verify audit hash chain via audit-service API.
set -euo pipefail

AUDIT_URL="${AUDIT_SERVICE_URL:-http://localhost:8090}"
res=$(curl -sS "${AUDIT_URL}/api/v1/audit/verify" || true)
if [[ -z "$res" ]]; then
  echo "audit verify: unreachable at ${AUDIT_URL}" >&2
  exit 1
fi
valid=$(echo "$res" | jq -r '.valid // false')
if [[ "$valid" != "true" ]]; then
  echo "audit chain invalid: $res" >&2
  exit 1
fi
count=$(echo "$res" | jq -r '.checkedCount // 0')
echo "Audit chain valid ($count entries checked)"
