#!/usr/bin/env bash
# Phase 3 local demo helper: health, policy eval samples, optional live alert, URLs.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
chmod +x "$ROOT/scripts/verify-audit-chain.sh" "$ROOT/mocks/simulators/payment-burst.sh" 2>/dev/null || true

AUDIT_URL="${AUDIT_SERVICE_URL:-http://localhost:8090}"
CEDAR_URL="${CEDAR_SERVICE_URL:-http://localhost:8091}"
DECISION_URL="${DECISION_SERVICE_URL:-http://localhost:8092}"
EXPLORER_URL="${AUDIT_EXPLORER_URL:-http://localhost:3002}"
ALERT_CONSOLE_URL="${ALERT_CONSOLE_URL:-http://localhost:3000}"
ALERT_URL="${ALERT_SERVICE_URL:-http://localhost:8085}"
ALERT_DB_URL="${ALERT_DB_URL:-postgres://alert:alert@localhost:5435/alerts?sslmode=disable}"
COMPOSE_FILE="${COMPOSE_FILE:-$ROOT/docker-compose.dev.yml}"

TRIGGER_ALERT=0
RESTART_POLICIES=0
SKIP_VERIFY=0

usage() {
  cat <<'EOF'
Usage: ./scripts/demo-phase3.sh [options]

Warm-check Phase 3 services and print demo URLs. Optionally restart policy
services (fixes empty policy volume mounts) and trigger a fresh INT-M001 alert.

Options:
  --trigger-alert     Run payment burst and wait for alert + evidenceRef
  --restart-policies  docker compose restart cedar-service decision-service
  --skip-verify       Skip hash-chain verification
  -h, --help          Show this help

Environment: AUDIT_SERVICE_URL, CEDAR_SERVICE_URL, DECISION_SERVICE_URL,
AUDIT_EXPLORER_URL, ALERT_CONSOLE_URL, ALERT_SERVICE_URL, ALERT_DB_URL

Full automated proof: ./scripts/smoke-test-phase3.sh
Runbook: docs/demo-phase3.md
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --trigger-alert) TRIGGER_ALERT=1; shift ;;
    --restart-policies) RESTART_POLICIES=1; shift ;;
    --skip-verify) SKIP_VERIFY=1; shift ;;
    -h|--help) usage; exit 0 ;;
    *) echo "unknown option: $1" >&2; usage >&2; exit 1 ;;
  esac
done

need_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "missing required command: $1" >&2
    exit 1
  fi
}

need_cmd curl
need_cmd jq

echo "=== Phase 3 demo prep ==="

if [[ "$RESTART_POLICIES" == "1" ]]; then
  echo "Restarting cedar-service and decision-service..."
  docker compose -f "$COMPOSE_FILE" restart cedar-service decision-service
  sleep 5
fi

echo "Checking service health..."
for url in "$AUDIT_URL" "$CEDAR_URL" "$DECISION_URL" "$ALERT_URL"; do
  if ! curl -sf "${url}/api/v1/health" >/dev/null; then
    echo "unhealthy: $url — run: docker compose -f docker-compose.dev.yml up -d --wait" >&2
    exit 1
  fi
done
echo "All Phase 3 APIs healthy."

echo ""
echo "Cedar INT-R003 (deny without role):"
deny=$(curl -sf -X POST "${CEDAR_URL}/api/v1/evaluate" \
  -H 'Content-Type: application/json' \
  -d '{"ruleCode":"INT-R003","principal":{"id":"demo","roles":[]},"resource":{"type":"TwinData","id":"t1","attributes":{"sensitivity":"high"}}}')
echo "$deny" | jq '{ruleCode, outcome, rationale}'
echo "$deny" | jq -e '.outcome == "Deny"' >/dev/null

echo ""
echo "Cedar INT-R003 (allow with Reporter role):"
allow=$(curl -sf -X POST "${CEDAR_URL}/api/v1/evaluate" \
  -H 'Content-Type: application/json' \
  -H 'X-Roles: Reporter' \
  -d '{"ruleCode":"INT-R003","principal":{"id":"demo"},"resource":{"type":"TwinData","id":"t1","attributes":{"sensitivity":"high"}}}')
echo "$allow" | jq '{ruleCode, outcome}'
echo "$allow" | jq -e '.outcome == "Allow"' >/dev/null

echo ""
echo "Zen BASEL-R001 (LCR 0.90 → Deny):"
zen=$(curl -sf -X POST "${DECISION_URL}/api/v1/evaluate" \
  -H 'Content-Type: application/json' \
  -d '{"ruleCode":"BASEL-R001","input":{"lcr":0.9,"personaId":"44444444-4444-4444-4444-444444444401"}}')
echo "$zen" | jq '{ruleCode, outcome, rationale}'
echo "$zen" | jq -e '.outcome == "Deny"' >/dev/null

entry_count=$(curl -sf "${AUDIT_URL}/api/v1/audit/entries?limit=1" | jq 'length')
echo ""
echo "Audit ledger: at least $entry_count indexed entries visible."

if [[ "$SKIP_VERIFY" != "1" ]]; then
  echo ""
  echo "Verifying hash chain..."
  "$ROOT/scripts/verify-audit-chain.sh"
fi

if [[ "$TRIGGER_ALERT" == "1" ]]; then
  if ! command -v psql >/dev/null 2>&1; then
    echo "--trigger-alert requires psql" >&2
    exit 1
  fi
  echo ""
  echo "Triggering INT-M001 payment burst (live demo moment)..."
  REDIS_CONTAINER="${REDIS_CONTAINER:-digitaltwin-redis-1}"
  BURST_ACCOUNT_ID="${BURST_ACCOUNT_ID:-aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaa01}"
  psql "$ALERT_DB_URL" -v ON_ERROR_STOP=1 -c "DELETE FROM compliance_alerts WHERE rule_code = 'INT-M001';" >/dev/null 2>&1 || true
  if docker ps --format '{{.Names}}' | grep -qx "$REDIS_CONTAINER"; then
    docker exec "$REDIS_CONTAINER" redis-cli DEL \
      "vel:${DEFAULT_TENANT_ID:-00000000-0000-0000-0000-000000000001}:${BURST_ACCOUNT_ID}:1h" >/dev/null 2>&1 || true
  fi
  "$ROOT/mocks/simulators/payment-burst.sh"
  echo "Waiting for alert + evidenceRef (up to 60s)..."
  found=""
  for _ in $(seq 1 120); do
    row=$(psql "$ALERT_DB_URL" -Atqc \
      "SELECT alert_id, evidence_ref FROM compliance_alerts WHERE rule_code = 'INT-M001' AND evidence_ref IS NOT NULL ORDER BY detected_at DESC LIMIT 1;" 2>/dev/null || true)
    if [[ -n "$row" ]]; then
      found="$row"
      break
    fi
    sleep 0.5
  done
  if [[ -z "$found" ]]; then
    echo "timed out waiting for INT-M001 alert with evidenceRef" >&2
    exit 1
  fi
  alert_id="${found%%|*}"
  evidence_ref="${found#*|}"
  echo "Fresh alert: $alert_id"
  echo "evidenceRef: $evidence_ref"
  echo "Alert detail: ${ALERT_CONSOLE_URL}/alerts/${alert_id}"
  echo "Audit entry:  ${EXPLORER_URL}/entries/${evidence_ref}"
fi

sample_ref=$(psql "$ALERT_DB_URL" -Atqc \
  "SELECT evidence_ref FROM compliance_alerts WHERE evidence_ref IS NOT NULL ORDER BY detected_at DESC LIMIT 1;" 2>/dev/null || true)

echo ""
echo "=== Demo URLs ==="
echo "Alert Console:    ${ALERT_CONSOLE_URL}"
echo "Audit Explorer:   ${EXPLORER_URL}"
echo "Cedar evaluate:   ${CEDAR_URL}/api/v1/evaluate"
echo "Zen evaluate:     ${DECISION_URL}/api/v1/evaluate"
echo "Audit API:        ${AUDIT_URL}/api/v1/audit/entries"
if [[ -n "$sample_ref" ]]; then
  echo ""
  echo "Sample linked alert → audit:"
  echo "  ${EXPLORER_URL}/entries/${sample_ref}"
fi
echo ""
echo "Demo prep OK. See docs/demo-phase3.md for the presentation script."
