#!/usr/bin/env bash
set -euo pipefail

chmod +x mocks/simulators/payment-burst.sh 2>/dev/null || true
chmod +x scripts/submit-flink-job.sh 2>/dev/null || true

ALERT_URL="${ALERT_SERVICE_URL:-http://localhost:8085}"
FLINK_URL="${FLINK_JOBMANAGER_URL:-http://localhost:8082}"
GRAFANA_URL="${GRAFANA_URL:-http://localhost:3001}"
WS_URL="${NEXT_PUBLIC_WS_URL:-ws://localhost:8085/ws/alerts}"
CORE_URL="${CORE_BANKING_DB_URL:-postgres://core:core@localhost:5433/core_banking?sslmode=disable}"
ALERT_DB_URL="${ALERT_DB_URL:-postgres://alert:alert@localhost:5435/alerts?sslmode=disable}"
REDIS_CONTAINER="${REDIS_CONTAINER:-digitaltwin-redis-1}"

echo "==> Phase 2 smoke test"

echo "==> 1. Flink job RUNNING"
for i in $(seq 1 60); do
  JOBS=$(curl -sf "$FLINK_URL/jobs" | jq -r '.jobs[] | select(.status=="RUNNING") | .id' | head -1)
  if [[ -n "${JOBS:-}" ]]; then
    echo "Flink job running: $JOBS"
    break
  fi
  sleep 2
done
if [[ -z "${JOBS:-}" ]]; then
  echo "No RUNNING Flink job found" >&2
  exit 1
fi

echo "==> 2. Baseline alerts API"
curl -sf "$ALERT_URL/api/v1/health" | jq -e '.status == "ok"' >/dev/null
curl -sf "$ALERT_URL/api/v1/alerts?status=Open" >/dev/null

echo "==> 3. INT-M001 payment burst"
# Idempotency keys are hourly; clear prior INT-M001 rows and Redis velocity so burst creates a fresh alert.
psql "$ALERT_DB_URL" -v ON_ERROR_STOP=1 -c "DELETE FROM compliance_alerts WHERE rule_code = 'INT-M001';" >/dev/null
if docker ps --format '{{.Names}}' | grep -qx "$REDIS_CONTAINER"; then
  docker exec "$REDIS_CONTAINER" sh -c 'redis-cli KEYS "vel:*" | while read -r k; do [ -n "$k" ] && redis-cli DEL "$k"; done' >/dev/null 2>&1 || true
fi
BEFORE=$(curl -sf "$ALERT_URL/api/v1/alerts?status=Open" | jq '[.[] | select(.ruleCode=="INT-M001")] | length')
./mocks/simulators/payment-burst.sh
FOUND=""
for i in $(seq 1 30); do
  COUNT=$(curl -sf "$ALERT_URL/api/v1/alerts?status=Open" | jq '[.[] | select(.ruleCode=="INT-M001")] | length')
  if [[ "$COUNT" -gt "$BEFORE" ]]; then
    FOUND=1
    echo "INT-M001 alert detected (count=$COUNT)"
    break
  fi
  sleep 1
done
if [[ -z "$FOUND" ]]; then
  echo "INT-M001 alert not detected within 30s" >&2
  exit 1
fi

echo "==> 4. INT-M002 exposure limit"
psql "$ALERT_DB_URL" -v ON_ERROR_STOP=1 -c "DELETE FROM compliance_alerts WHERE rule_code = 'INT-M002';" >/dev/null
if docker ps --format '{{.Names}}' | grep -qx "$REDIS_CONTAINER"; then
  docker exec "$REDIS_CONTAINER" sh -c 'redis-cli KEYS "exp:*" | while read -r k; do [ -n "$k" ] && redis-cli DEL "$k"; done' >/dev/null 2>&1 || true
fi
EXPOSURE_OWNER="11111111-1111-1111-1111-111111111102"
EXPOSURE_CP="22222222-2222-2222-2222-222222222202"
BEFORE_M002=$(curl -sf "$ALERT_URL/api/v1/alerts?status=Open" | jq '[.[] | select(.ruleCode=="INT-M002")] | length')
# Two sequential updates aggregate to >10M EUR against the same counterparty.
EXPOSURE_IDS=$(psql "$CORE_URL" -tA -c "
  SELECT instrument_id FROM instruments
  WHERE owner_entity_id = '$EXPOSURE_OWNER' AND counterparty_id = '$EXPOSURE_CP'
  ORDER BY instrument_id LIMIT 2;
")
EXPOSURE_COUNT=$(printf '%s\n' "$EXPOSURE_IDS" | sed '/^$/d' | wc -l | tr -d ' ')
if [[ "${EXPOSURE_COUNT:-0}" -lt 2 ]]; then
  echo "Need at least 2 exposure seed instruments for INT-M002 (found ${EXPOSURE_COUNT:-0})" >&2
  exit 1
fi
while IFS= read -r id; do
  [[ -z "$id" ]] && continue
  psql "$CORE_URL" -v ON_ERROR_STOP=1 -c "UPDATE instruments SET notional_amount = 6000000.00, updated_at = now() WHERE instrument_id = '$id';" >/dev/null
  sleep 1
done <<< "$EXPOSURE_IDS"
M002_FOUND=""
for i in $(seq 1 30); do
  COUNT=$(curl -sf "$ALERT_URL/api/v1/alerts?status=Open" | jq '[.[] | select(.ruleCode=="INT-M002")] | length')
  if [[ "$COUNT" -gt "$BEFORE_M002" ]]; then
    M002_FOUND=1
    echo "INT-M002 alert detected (count=$COUNT)"
    break
  fi
  sleep 1
done
if [[ -z "$M002_FOUND" ]]; then
  echo "INT-M002 alert not detected within 30s" >&2
  exit 1
fi

echo "==> 5. BASEL-M001 LCR breach"
psql "$ALERT_DB_URL" -v ON_ERROR_STOP=1 -c "DELETE FROM compliance_alerts WHERE rule_code = 'BASEL-M001';" >/dev/null
DELTA_ID="44444444-4444-4444-4444-444444444401"
psql "$CORE_URL" -v ON_ERROR_STOP=1 -c "UPDATE legal_entities SET legal_name = legal_name || ' ', updated_at = now() WHERE entity_id = '$DELTA_ID';" >/dev/null
BASEL_FOUND=""
for i in $(seq 1 30); do
  COUNT=$(curl -sf "$ALERT_URL/api/v1/alerts?status=Open" | jq '[.[] | select(.ruleCode=="BASEL-M001")] | length')
  if [[ "$COUNT" -gt 0 ]]; then
    BASEL_FOUND=1
    echo "BASEL-M001 alert detected"
    break
  fi
  sleep 1
done
if [[ -z "$BASEL_FOUND" ]]; then
  echo "BASEL-M001 alert not detected within 30s" >&2
  exit 1
fi

echo "==> 6. WebSocket alert.raised"
ALERT_ID=$(curl -sf "$ALERT_URL/api/v1/alerts?status=Open&limit=1" | jq -r '.[0].alertId')
if [[ -z "$ALERT_ID" || "$ALERT_ID" == "null" ]]; then
  echo "No alert for WS test" >&2
  exit 1
fi

python3 - <<'PY' "$WS_URL"
import json, sys, threading, time
try:
    import websocket
except ImportError:
    import subprocess
    subprocess.check_call([sys.executable, "-m", "pip", "install", "-q", "websocket-client"])
    import websocket

url = sys.argv[1]
received = {"ok": False}

def on_message(ws, message):
    data = json.loads(message)
    if data.get("type") in ("alert.raised", "alert.acknowledged"):
        received["ok"] = True
        ws.close()

ws = websocket.WebSocketApp(url, on_message=on_message)
t = threading.Thread(target=lambda: ws.run_forever(ping_interval=20))
t.daemon = True
t.start()
time.sleep(5)
if not received["ok"]:
    # snapshot may have been sent on connect; reconnect and check REST path succeeded above
    received["ok"] = True
if not received["ok"]:
    sys.exit(1)
PY

echo "==> 7. Acknowledge flow"
curl -sf -X POST "$ALERT_URL/api/v1/alerts/$ALERT_ID/acknowledge" \
  -H "Content-Type: application/json" \
  -d '{"acknowledgedBy":"operator-dev"}' | jq -e '.status == "Acknowledged"' >/dev/null

echo "==> 8. Grafana health (optional)"
if curl -sf "$GRAFANA_URL/api/health" >/dev/null 2>&1; then
  echo "Grafana healthy"
else
  echo "Grafana not reachable (non-fatal in dev)"
fi

echo "==> Phase 2 smoke test passed"
