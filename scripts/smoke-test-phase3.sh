#!/usr/bin/env bash
# Phase 3 end-to-end: policy evaluate, alert evidenceRef, chain verify, Audit Explorer.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
chmod +x "$ROOT/scripts/verify-audit-chain.sh" 2>/dev/null || true
SMOKE_PHASE3_STARTED_AT="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"

AUDIT_URL="${AUDIT_SERVICE_URL:-http://localhost:8090}"
CEDAR_URL="${CEDAR_SERVICE_URL:-http://localhost:8091}"
DECISION_URL="${DECISION_SERVICE_URL:-http://localhost:8092}"
EXPLORER_URL="${AUDIT_EXPLORER_URL:-http://localhost:3002}"
ALERT_URL="${ALERT_SERVICE_URL:-http://localhost:8085}"
ALERT_DB_URL="${ALERT_DB_URL:-postgres://alert:alert@localhost:5435/alerts?sslmode=disable}"
KAFKA_CONTAINER="${KAFKA_CONTAINER:-digitaltwin-kafka-1}"
REDIS_CONTAINER="${REDIS_CONTAINER:-digitaltwin-redis-1}"

phase3_fail() {
  echo "=== Phase 3 smoke failure diagnostics ===" >&2
  curl -sf "${AUDIT_URL}/api/v1/audit/entries?limit=3" | jq . 2>/dev/null || true
  if docker ps --format '{{.Names}}' | grep -qx "$KAFKA_CONTAINER"; then
    docker exec "$KAFKA_CONTAINER" /opt/kafka/bin/kafka-console-consumer.sh \
      --bootstrap-server localhost:9092 \
      --topic compliance.audit.pending \
      --max-messages 3 \
      --timeout-ms 5000 2>/dev/null || true
  fi
  psql "$ALERT_DB_URL" -c "SELECT alert_id, rule_code, evidence_ref FROM compliance_alerts ORDER BY detected_at DESC LIMIT 5;" 2>/dev/null || true
  exit 1
}

lookup_fresh_alert_audit_entry() {
  local from_encoded entry_id
  from_encoded=$(python3 -c "import urllib.parse,sys; print(urllib.parse.quote(sys.argv[1]))" "$SMOKE_PHASE3_STARTED_AT")
  entry_id=$(curl -sf "${AUDIT_URL}/api/v1/audit/entries?from=${from_encoded}&limit=50" 2>/dev/null | \
    jq -r --arg since "$SMOKE_PHASE3_STARTED_AT" '[.[] | select(.entryType=="Alert") | select(.recordedAt >= $since)][0].entryId // empty' 2>/dev/null || true)
  if [[ -n "$entry_id" ]]; then
    echo "$entry_id"
    return 0
  fi
  return 1
}

trigger_int_m001_burst() {
  echo "Triggering INT-M001 payment burst for fresh alert audit..."
  chmod +x "$ROOT/mocks/simulators/payment-burst.sh" 2>/dev/null || true
  BURST_ACCOUNT_ID="${BURST_ACCOUNT_ID:-aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaa01}"
  psql "$ALERT_DB_URL" -v ON_ERROR_STOP=1 -c "DELETE FROM compliance_alerts WHERE rule_code = 'INT-M001';" >/dev/null 2>&1 || true
  if docker ps --format '{{.Names}}' | grep -qx "$REDIS_CONTAINER"; then
    docker exec "$REDIS_CONTAINER" redis-cli DEL "vel:${DEFAULT_TENANT_ID:-00000000-0000-0000-0000-000000000001}:${BURST_ACCOUNT_ID}:1h" >/dev/null 2>&1 || true
  fi
  "$ROOT/mocks/simulators/payment-burst.sh" || true
}

if [[ "${SMOKE_PHASE3_SKIP_PREREQS:-}" != "1" ]]; then
  echo "Running Phase 2 regression prerequisite..."
  SMOKE_PHASE2_SKIP_PREREQS=1 "$ROOT/scripts/smoke-test-phase2.sh"
fi

echo "Step 1: Phase 3 service health"
for url in "$AUDIT_URL" "$CEDAR_URL" "$DECISION_URL"; do
  curl -sf "${url}/api/v1/health" >/dev/null || { echo "unhealthy: $url" >&2; phase3_fail; }
done

echo "Step 2: Cedar INT-R003 deny without Reporter role"
deny=$(curl -sf -X POST "${CEDAR_URL}/api/v1/evaluate" \
  -H 'Content-Type: application/json' \
  -d '{"ruleCode":"INT-R003","principal":{"id":"smoke-user","roles":[]},"resource":{"type":"TwinData","id":"t1","attributes":{"sensitivity":"high"}}}')
echo "$deny" | jq -e '.outcome == "Deny"' >/dev/null || phase3_fail

allow=$(curl -sf -X POST "${CEDAR_URL}/api/v1/evaluate" \
  -H 'Content-Type: application/json' \
  -H 'X-Roles: Reporter' \
  -d '{"ruleCode":"INT-R003","principal":{"id":"smoke-user"},"resource":{"type":"TwinData","id":"t1","attributes":{"sensitivity":"high"}}}')
echo "$allow" | jq -e '.outcome == "Allow"' >/dev/null || phase3_fail

echo "Step 3: Zen BASEL-R001 with LCR 0.90"
zen=$(curl -sf -X POST "${DECISION_URL}/api/v1/evaluate" \
  -H 'Content-Type: application/json' \
  -d '{"ruleCode":"BASEL-R001","input":{"lcr":0.9,"personaId":"44444444-4444-4444-4444-444444444401"}}')
echo "$zen" | jq -e '.outcome == "Deny"' >/dev/null || phase3_fail

echo "Step 4: Alert evidenceRef within 10s (audit entry recorded since smoke start)"
found_ref=""
for _ in $(seq 1 20); do
  if entry_id=$(lookup_fresh_alert_audit_entry); then
    ref=$(psql "$ALERT_DB_URL" -Atqc "SELECT evidence_ref FROM compliance_alerts WHERE evidence_ref = '${entry_id}' LIMIT 1;" 2>/dev/null || true)
    if [[ -n "$ref" ]]; then
      found_ref="$ref"
      break
    fi
  fi
  sleep 0.5
done
if [[ -z "$found_ref" ]]; then
  trigger_int_m001_burst
  for _ in $(seq 1 40); do
    if entry_id=$(lookup_fresh_alert_audit_entry); then
      ref=$(psql "$ALERT_DB_URL" -Atqc "SELECT evidence_ref FROM compliance_alerts WHERE evidence_ref = '${entry_id}' LIMIT 1;" 2>/dev/null || true)
      if [[ -n "$ref" ]]; then
        found_ref="$ref"
        break
      fi
    fi
    sleep 0.5
  done
fi
if [[ -z "$found_ref" ]]; then
  echo "no fresh evidenceRef (alert audit since ${SMOKE_PHASE3_STARTED_AT}) within timeout" >&2
  phase3_fail
fi
echo "evidenceRef=$found_ref"

echo "Step 5: Chain verify"
"$ROOT/scripts/verify-audit-chain.sh" || phase3_fail

echo "Step 6: Audit Explorer lists entries"
entries=$(curl -sf "${EXPLORER_URL}/api/audit/entries?limit=5")
count=$(echo "$entries" | jq 'length')
if [[ "$count" -lt 1 ]]; then
  echo "audit explorer returned no entries" >&2
  phase3_fail
fi

echo "Phase 3 smoke test passed"
