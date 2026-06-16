#!/usr/bin/env bash
# Phase 2 end-to-end smoke — localization map:
#   INT-M001  → Redis vel:{tenant}:{account}:1h  (Debezium → Flink velocity)
#   INT-M002  → Redis exp:{tenant}:{owner}:{cp}  (instrument twin → Flink exposure)
#   BASEL-M001 → Redis lcr:{tenant}:{entity}     (institution twin → Flink LCR)
#   alerts empty after Redis OK → Flink compliance.alerts → alert-service consumer
#
# Run a subset: SMOKE_PHASE2_ONLY=M002 (or M001,BASEL comma-separated).
# Skip Flink/API prereqs when stack is already warm: SMOKE_PHASE2_SKIP_PREREQS=1
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

chmod +x "$ROOT/mocks/simulators/payment-burst.sh" 2>/dev/null || true
chmod +x "$ROOT/scripts/submit-flink-job.sh" 2>/dev/null || true

ALERT_URL="${ALERT_SERVICE_URL:-http://localhost:8085}"
CONSOLE_URL="${ALERT_CONSOLE_URL:-http://localhost:3000}"
FLINK_URL="${FLINK_JOBMANAGER_URL:-http://localhost:8082}"
GRAFANA_URL="${GRAFANA_URL:-http://localhost:3030}"
WS_URL="${ALERT_SERVICE_WS_URL:-${NEXT_PUBLIC_WS_URL:-ws://localhost:8085/ws/alerts}}"
CORE_URL="${CORE_BANKING_DB_URL:-postgres://core:core@localhost:5433/core_banking?sslmode=disable}"
STATE_URL="${STATE_DB_URL:-postgres://state:state@localhost:5434/twin_state?sslmode=disable}"
ALERT_DB_URL="${ALERT_DB_URL:-postgres://alert:alert@localhost:5435/alerts?sslmode=disable}"
REDIS_CONTAINER="${REDIS_CONTAINER:-digitaltwin-redis-1}"
ALERT_WAIT_SEC="${SMOKE_PHASE2_ALERT_WAIT_SEC:-30}"
REDIS_WAIT_SEC="${SMOKE_PHASE2_REDIS_WAIT_SEC:-45}"
TENANT_ID="${DEFAULT_TENANT_ID:-00000000-0000-0000-0000-000000000001}"
KAFKA_CONTAINER="${KAFKA_CONTAINER:-digitaltwin-kafka-1}"
CONNECT_URL="${DEBEZIUM_CONNECT_URL:-http://localhost:8083}"

if ! docker ps --format '{{.Names}}' | grep -qx "$REDIS_CONTAINER"; then
  DETECTED=$(docker ps --format '{{.Names}}' | grep -E 'redis-1$' | head -1 || true)
  if [[ -n "$DETECTED" ]]; then
    REDIS_CONTAINER="$DETECTED"
  fi
fi

if ! docker ps --format '{{.Names}}' | grep -qx "$KAFKA_CONTAINER"; then
  DETECTED_KAFKA=$(docker ps --format '{{.Names}}' | grep -E 'kafka-1$' | head -1 || true)
  if [[ -n "$DETECTED_KAFKA" ]]; then
    KAFKA_CONTAINER="$DETECTED_KAFKA"
  fi
fi

CEP_VELOCITY_MAX="${CEP_VELOCITY_MAX_PER_HOUR:-50}"
CEP_EXPOSURE_LIMIT="${CEP_EXPOSURE_LIMIT_EUR:-10000000}"
CEP_LCR_MIN="${CEP_LCR_MINIMUM:-1.0}"

# shellcheck source=scripts/smoke-lib-phase2.sh
source "$ROOT/scripts/smoke-lib-phase2.sh"

smoke_step_enabled() {
  local name="$1"
  local only="${SMOKE_PHASE2_ONLY:-}"
  [[ -z "$only" ]] && return 0
  local part
  IFS=',' read -ra PARTS <<< "$only"
  for part in "${PARTS[@]}"; do
    part="${part// /}"
    [[ "$part" == "$name" ]] && return 0
  done
  return 1
}

smoke_needs_prereqs() {
  [[ "${SMOKE_PHASE2_SKIP_PREREQS:-}" == "1" ]] && return 1
  [[ -z "${SMOKE_PHASE2_ONLY:-}" ]] && return 0
  smoke_step_enabled M001 || smoke_step_enabled M002 || smoke_step_enabled BASEL \
    || smoke_step_enabled WS || smoke_step_enabled ACK || smoke_step_enabled REDIS \
    || smoke_step_enabled GRAFANA
}

dump_int_m001_debug() {
  echo "--- INT-M001 debug ---" >&2
  curl -sf "$CONNECT_URL/connectors/core-banking-cdc/status" 2>/dev/null | jq . || echo "debezium status unavailable" >&2
  PAYMENT_ROWS=$(psql "$CORE_URL" -tA -c "SELECT COUNT(*) FROM payments;" 2>/dev/null || echo "?")
  echo "payments rows in core DB: $PAYMENT_ROWS" >&2
  if docker ps --format '{{.Names}}' | grep -qx "$KAFKA_CONTAINER"; then
    docker exec "$KAFKA_CONTAINER" /opt/kafka/bin/kafka-get-offsets.sh \
      --bootstrap-server localhost:9092 --topic domain.events.public.payments 2>/dev/null | head -5 >&2 || true
    docker exec "$KAFKA_CONTAINER" /opt/kafka/bin/kafka-get-offsets.sh \
      --bootstrap-server localhost:9092 --topic compliance.alerts 2>/dev/null | head -5 >&2 || true
  fi
  if docker ps --format '{{.Names}}' | grep -qx "$REDIS_CONTAINER"; then
    docker exec "$REDIS_CONTAINER" redis-cli KEYS "vel:${TENANT_ID}:*" 2>/dev/null | head -5 >&2 || true
    if [[ -n "${BURST_ACCOUNT_ID:-}" ]]; then
      echo "redis vel for burst account: $(docker exec "$REDIS_CONTAINER" redis-cli GET "vel:${TENANT_ID}:${BURST_ACCOUNT_ID}:1h" 2>/dev/null || echo '?')" >&2
    fi
  fi
  curl -sf "$ALERT_URL/api/v1/alerts?status=Open" 2>/dev/null | jq '[.[] | {ruleCode, idempotencyKey}]' >&2 || true
  psql "$ALERT_DB_URL" -tA -c "SELECT rule_code, status, idempotency_key FROM compliance_alerts ORDER BY detected_at DESC LIMIT 5;" 2>/dev/null >&2 || true
}

dump_basel_m001_debug() {
  local delta_id="${1:-44444444-4444-4444-4444-444444444401}"
  echo "--- BASEL-M001 debug ---" >&2
  psql "$CORE_URL" -tA -c "SELECT entity_id, lcr, hqla, net_cash_outflows_30d FROM legal_entities WHERE entity_id = '$delta_id';" 2>/dev/null >&2 || true
  psql "$STATE_URL" -tA -c "SELECT persona_id, state_version, left(current_state::text, 500) FROM twin_personas WHERE persona_id = '$delta_id';" 2>/dev/null >&2 || true
  echo "twin liquidity lcr: $(twin_institution_lcr "$delta_id" || echo '?')" >&2
  smoke_twin_pipeline_debug
  if docker ps --format '{{.Names}}' | grep -qx "$KAFKA_CONTAINER"; then
    docker exec "$KAFKA_CONTAINER" /opt/kafka/bin/kafka-get-offsets.sh \
      --bootstrap-server localhost:9092 --topic twin.state.updated 2>/dev/null | head -5 >&2 || true
    docker exec "$KAFKA_CONTAINER" /opt/kafka/bin/kafka-get-offsets.sh \
      --bootstrap-server localhost:9092 --topic compliance.alerts 2>/dev/null | head -5 >&2 || true
  fi
  if docker ps --format '{{.Names}}' | grep -qx "$REDIS_CONTAINER"; then
    echo "redis lcr for delta: $(docker exec "$REDIS_CONTAINER" redis-cli GET "lcr:${TENANT_ID}:${delta_id}" 2>/dev/null || echo '?')" >&2
  fi
  curl -sf "$ALERT_URL/api/v1/alerts?status=Open" 2>/dev/null | jq '[.[] | select(.ruleCode=="BASEL-M001")]' >&2 || true
}

dump_int_m002_debug() {
  local owner_id="${1:-11111111-1111-1111-1111-111111111102}"
  local cp_id="${2:-22222222-2222-2222-2222-222222222202}"
  echo "--- INT-M002 debug ---" >&2
  psql "$CORE_URL" -c "SELECT instrument_id, owner_entity_id, counterparty_id, notional_amount, currency FROM instruments WHERE owner_entity_id = '$owner_id' AND counterparty_id = '$cp_id' ORDER BY instrument_id LIMIT 5;" 2>/dev/null >&2 || true
  psql "$STATE_URL" -tA -c "SELECT persona_id, state_version, left(current_state::text, 400) FROM twin_personas WHERE persona_type = 'Instrument' AND current_state::text LIKE '%$owner_id%' ORDER BY updated_at DESC LIMIT 3;" 2>/dev/null >&2 || true
  smoke_twin_pipeline_debug
  if docker ps --format '{{.Names}}' | grep -qx "$REDIS_CONTAINER"; then
    echo "redis exp aggregate: $(docker exec "$REDIS_CONTAINER" redis-cli GET "exp:${TENANT_ID}:${owner_id}:${cp_id}" 2>/dev/null || echo '?')" >&2
    docker exec "$REDIS_CONTAINER" redis-cli KEYS "exp:${TENANT_ID}:*" 2>/dev/null | head -8 >&2 || true
  fi
  if docker ps --format '{{.Names}}' | grep -qx "$KAFKA_CONTAINER"; then
    docker exec "$KAFKA_CONTAINER" /opt/kafka/bin/kafka-get-offsets.sh \
      --bootstrap-server localhost:9092 --topic twin.state.updated 2>/dev/null | head -5 >&2 || true
  fi
  curl -sf "$ALERT_URL/api/v1/alerts?status=Open" 2>/dev/null | jq '[.[] | select(.ruleCode=="INT-M002")]' >&2 || true
}

echo "==> Phase 2 smoke test${SMOKE_PHASE2_ONLY:+, only=$SMOKE_PHASE2_ONLY}"

if smoke_needs_prereqs; then
  echo "==> 0. Preflight"
  curl -sf "$ALERT_URL/api/v1/health" | jq -e '.status == "ok"' >/dev/null
  INSTRUMENT_ROWS=$(psql "$CORE_URL" -tA -c "SELECT COUNT(*) FROM instruments;" 2>/dev/null || echo "0")
  if [[ "${INSTRUMENT_ROWS:-0}" -lt 1 ]]; then
    echo "Preflight failed: no instruments in core banking (run ./scripts/seed.sh)" >&2
    exit 1
  fi
  EXPOSURE_SEED=$(psql "$CORE_URL" -tA -c "
    SELECT COUNT(*) FROM instruments i
    WHERE i.instrument_id IN (SELECT instrument_id FROM instruments ORDER BY instrument_id LIMIT 2)
      AND i.owner_entity_id = '11111111-1111-1111-1111-111111111102'
      AND i.counterparty_id = '22222222-2222-2222-2222-222222222202';
  " 2>/dev/null || echo "0")
  if [[ "${EXPOSURE_SEED:-0}" -lt 2 ]]; then
    echo "Preflight failed: need INT-M002 seed pair on first 2 instruments (found ${EXPOSURE_SEED:-0}; run ./scripts/seed.sh)" >&2
    exit 1
  fi
  echo "Preflight ok (instruments=$INSTRUMENT_ROWS, exposure_seed_pair=$EXPOSURE_SEED)"

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

CHECKPOINT_STATS=$(curl -sf "$FLINK_URL/jobs/$JOBS/checkpoints")
COMPLETED=$(echo "$CHECKPOINT_STATS" | jq -r '.counts.completed // 0')
FAILED=$(echo "$CHECKPOINT_STATS" | jq -r '.counts.failed // 0')
TOTAL=$((COMPLETED + FAILED))
if [[ "$TOTAL" -gt 0 ]]; then
  SUCCESS_RATE=$(python3 - <<PY
completed = int("$COMPLETED")
failed = int("$FAILED")
total = completed + failed
print(f"{(completed / total) * 100:.2f}")
PY
)
  echo "Flink checkpoints: completed=$COMPLETED failed=$FAILED success_rate=${SUCCESS_RATE}%"
  python3 - <<PY
rate = float("$SUCCESS_RATE")
import sys
sys.exit(0 if rate >= 99.0 else 1)
PY
else
  echo "Flink checkpoints: no history yet (non-fatal during cold start)"
fi

  echo "==> 2. Baseline alerts API"
  curl -sf "$ALERT_URL/api/v1/health" | jq -e '.status == "ok"' >/dev/null
  curl -sf "$ALERT_URL/api/v1/alerts?status=Open" >/dev/null
fi

if smoke_step_enabled M001; then
echo "==> 3. INT-M001 payment burst"
# Idempotency keys are hourly; clear prior INT-M001 rows and Redis velocity so burst creates a fresh alert.
psql "$ALERT_DB_URL" -v ON_ERROR_STOP=1 -c "DELETE FROM compliance_alerts WHERE rule_code = 'INT-M001';" >/dev/null
BURST_ACCOUNT_ID=$(psql "$CORE_URL" -tA -c "SELECT account_id FROM accounts LIMIT 1;")
if docker ps --format '{{.Names}}' | grep -qx "$REDIS_CONTAINER"; then
  docker exec "$REDIS_CONTAINER" sh -c 'redis-cli KEYS "vel:*" | while read -r k; do [ -n "$k" ] && redis-cli DEL "$k"; done' >/dev/null 2>&1 || true
fi
BEFORE=$(curl -sf "$ALERT_URL/api/v1/alerts?status=Open" | jq '[.[] | select(.ruleCode=="INT-M001")] | length')
"$ROOT/mocks/simulators/payment-burst.sh"
BURST_END=$(python3 -c "import time; print(int(time.time() * 1000))")
VEL_MIN=$((CEP_VELOCITY_MAX + 1))
if ! wait_redis_gte "vel:${TENANT_ID}:${BURST_ACCOUNT_ID}:1h" "$VEL_MIN" 30 "INT-M001 velocity"; then
  dump_int_m001_debug
  exit 1
fi
FOUND=""
CONSUME_MS=""
for i in $(seq 1 "$ALERT_WAIT_SEC"); do
  COUNT=$(curl -sf "$ALERT_URL/api/v1/alerts?status=Open" | jq '[.[] | select(.ruleCode=="INT-M001")] | length')
  if [[ "$COUNT" -gt "$BEFORE" ]]; then
    FOUND=1
    NOW_MS=$(python3 -c "import time; print(int(time.time() * 1000))")
    CONSUME_MS=$((NOW_MS - BURST_END))
    echo "INT-M001 alert detected (count=$COUNT, consume_latency_ms=$CONSUME_MS)"
    break
  fi
  sleep 1
done
if [[ -z "$FOUND" ]]; then
  echo "INT-M001 alert not detected within ${ALERT_WAIT_SEC}s" >&2
  dump_int_m001_debug
  exit 1
fi
if [[ "$CONSUME_MS" -gt 2000 ]]; then
  echo "INT-M001 consume latency ${CONSUME_MS}ms exceeds 2000ms p99 budget" >&2
  exit 1
fi

curl -sf -o /dev/null -w "alert-console HTTP %{http_code}\n" "$CONSOLE_URL/"
UI_MS=$(( $(python3 -c "import time; print(int(time.time() * 1000))") - BURST_END ))
if [[ "$UI_MS" -gt 5000 ]]; then
  echo "Alert API visible ${UI_MS}ms after burst exceeds 5000ms UI budget" >&2
  exit 1
fi
echo "Alert visible via API within ${UI_MS}ms (UI shell reachable at $CONSOLE_URL)"
fi

if smoke_step_enabled M002; then
echo "==> 4. INT-M002 exposure limit"
psql "$ALERT_DB_URL" -v ON_ERROR_STOP=1 -c "DELETE FROM compliance_alerts WHERE rule_code = 'INT-M002';" >/dev/null
if docker ps --format '{{.Names}}' | grep -qx "$REDIS_CONTAINER"; then
  docker exec "$REDIS_CONTAINER" sh -c 'redis-cli --scan --pattern "exp*" | while read -r k; do [ -n "$k" ] && redis-cli DEL "$k"; done' >/dev/null 2>&1 || true
fi
EXPOSURE_OWNER="11111111-1111-1111-1111-111111111102"
EXPOSURE_CP="22222222-2222-2222-2222-222222222202"
BEFORE_M002=$(curl -sf "$ALERT_URL/api/v1/alerts?status=Open" | jq '[.[] | select(.ruleCode=="INT-M002")] | length')
# Same pair as mocks/core-banking/seed/002_phase2_exposure.sql (first 2 instruments by id).
EXPOSURE_IDS=$(psql "$CORE_URL" -tA -c "
  SELECT instrument_id FROM instruments ORDER BY instrument_id LIMIT 2;
")
EXPOSURE_COUNT=$(printf '%s\n' "$EXPOSURE_IDS" | sed '/^$/d' | wc -l | tr -d ' ')
if [[ "${EXPOSURE_COUNT:-0}" -lt 2 ]]; then
  echo "Need at least 2 instruments for INT-M002 (found ${EXPOSURE_COUNT:-0})" >&2
  dump_int_m002_debug "$EXPOSURE_OWNER" "$EXPOSURE_CP"
  exit 1
fi
# Seed leaves these instruments at 6M; reset baseline so Flink exposure deltas are non-zero after Redis clear.
psql "$CORE_URL" -v ON_ERROR_STOP=1 -c "
  UPDATE instruments
  SET notional_amount = 1000000.00, updated_at = now()
  WHERE instrument_id IN (
    SELECT instrument_id FROM instruments ORDER BY instrument_id LIMIT 2
  );
" >/dev/null
while IFS= read -r id; do
  [[ -z "$id" ]] && continue
  if ! wait_twin_notional "$id" "1000000" "$REDIS_WAIT_SEC" "INT-M002 baseline $id"; then
    dump_int_m002_debug "$EXPOSURE_OWNER" "$EXPOSURE_CP"
    exit 1
  fi
done <<< "$EXPOSURE_IDS"
# Two sequential updates aggregate to >10M EUR against the same counterparty.
while IFS= read -r id; do
  [[ -z "$id" ]] && continue
  psql "$CORE_URL" -v ON_ERROR_STOP=1 -c "UPDATE instruments SET notional_amount = 6000000.00, updated_at = now() WHERE instrument_id = '$id';" >/dev/null
  if ! wait_twin_notional "$id" "6000000" "$REDIS_WAIT_SEC" "INT-M002 bump $id"; then
    dump_int_m002_debug "$EXPOSURE_OWNER" "$EXPOSURE_CP"
    exit 1
  fi
done <<< "$EXPOSURE_IDS"
if ! wait_redis_gte "exp:${TENANT_ID}:${EXPOSURE_OWNER}:${EXPOSURE_CP}" "$CEP_EXPOSURE_LIMIT" "$REDIS_WAIT_SEC" "INT-M002 exposure"; then
  dump_int_m002_debug "$EXPOSURE_OWNER" "$EXPOSURE_CP"
  exit 1
fi
M002_FOUND=""
for i in $(seq 1 "$ALERT_WAIT_SEC"); do
  COUNT=$(curl -sf "$ALERT_URL/api/v1/alerts?status=Open" | jq '[.[] | select(.ruleCode=="INT-M002")] | length')
  if [[ "$COUNT" -gt "$BEFORE_M002" ]]; then
    M002_FOUND=1
    echo "INT-M002 alert detected (count=$COUNT)"
    break
  fi
  sleep 1
done
if [[ -z "$M002_FOUND" ]]; then
  echo "INT-M002 alert not detected within ${ALERT_WAIT_SEC}s" >&2
  dump_int_m002_debug "$EXPOSURE_OWNER" "$EXPOSURE_CP"
  exit 1
fi
fi

if smoke_step_enabled BASEL; then
echo "==> 5. BASEL-M001 LCR breach"
psql "$ALERT_DB_URL" -v ON_ERROR_STOP=1 -c "DELETE FROM compliance_alerts WHERE rule_code = 'BASEL-M001';" >/dev/null
DELTA_ID="44444444-4444-4444-4444-444444444401"
if docker ps --format '{{.Names}}' | grep -qx "$REDIS_CONTAINER"; then
  docker exec "$REDIS_CONTAINER" redis-cli DEL "lcr:${TENANT_ID}:${DELTA_ID}" >/dev/null 2>&1 || true
fi
# Phase 2 spec: lower institution LCR below CEP minimum (default 1.0) so enrichment + twin.state.updated fire BASEL-M001.
psql "$CORE_URL" -v ON_ERROR_STOP=1 -c "
  UPDATE legal_entities
  SET
    lcr = 0.90,
    hqla = COALESCE(hqla, 450000000.00),
    net_cash_outflows_30d = COALESCE(net_cash_outflows_30d, 473684211.00),
    liquidity_currency = COALESCE(liquidity_currency, 'EUR'),
    updated_at = now()
  WHERE entity_id = '$DELTA_ID';
" >/dev/null
BASEL_TARGET_LCR="0.90"
if ! wait_twin_lcr_at_most "$DELTA_ID" "$BASEL_TARGET_LCR" "$REDIS_WAIT_SEC" "BASEL-M001 twin"; then
  dump_basel_m001_debug "$DELTA_ID"
  exit 1
fi
if ! wait_redis_lt "lcr:${TENANT_ID}:${DELTA_ID}" "$CEP_LCR_MIN" "$REDIS_WAIT_SEC" "BASEL-M001 LCR"; then
  dump_basel_m001_debug "$DELTA_ID"
  exit 1
fi
BASEL_FOUND=""
for i in $(seq 1 "$ALERT_WAIT_SEC"); do
  COUNT=$(curl -sf "$ALERT_URL/api/v1/alerts?status=Open" | jq '[.[] | select(.ruleCode=="BASEL-M001")] | length')
  if [[ "$COUNT" -gt 0 ]]; then
    BASEL_FOUND=1
    echo "BASEL-M001 alert detected"
    break
  fi
  sleep 1
done
if [[ -z "$BASEL_FOUND" ]]; then
  echo "BASEL-M001 alert not detected within ${ALERT_WAIT_SEC}s" >&2
  dump_basel_m001_debug "$DELTA_ID"
  exit 1
fi
fi

if smoke_step_enabled WS || smoke_step_enabled ACK; then
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
received = {"ok": False, "type": None}

def on_message(ws, message):
    data = json.loads(message)
    if data.get("type") == "alert.raised":
        received["ok"] = True
        received["type"] = data.get("type")
        ws.close()

ws = websocket.WebSocketApp(url, on_message=on_message)
t = threading.Thread(target=lambda: ws.run_forever(ping_interval=20))
t.daemon = True
t.start()
for _ in range(50):
    if received["ok"]:
        break
    time.sleep(0.1)
if not received["ok"]:
    sys.exit(1)
PY
echo "WebSocket alert.raised snapshot received"
fi

if smoke_step_enabled ACK; then
echo "==> 7. Acknowledge flow"
python3 - <<'PY' "$WS_URL" "$ALERT_ID" "$ALERT_URL"
import json, sys, threading, time
try:
    import websocket
    import urllib.request
except ImportError:
    import subprocess
    subprocess.check_call([sys.executable, "-m", "pip", "install", "-q", "websocket-client"])
    import websocket
    import urllib.request

url, alert_id, alert_url = sys.argv[1], sys.argv[2], sys.argv[3]
received = {"ok": False}

def on_message(ws, message):
    data = json.loads(message)
    payload = data.get("payload") or {}
    if data.get("type") == "alert.acknowledged" and payload.get("alertId") == alert_id:
        received["ok"] = True
        ws.close()

ws = websocket.WebSocketApp(url, on_message=on_message)
t = threading.Thread(target=lambda: ws.run_forever(ping_interval=20))
t.daemon = True
t.start()
time.sleep(1)
req = urllib.request.Request(
    f"{alert_url}/api/v1/alerts/{alert_id}/acknowledge",
    data=json.dumps({"acknowledgedBy": "operator-dev"}).encode(),
    headers={"Content-Type": "application/json"},
    method="POST",
)
with urllib.request.urlopen(req) as resp:
    body = json.loads(resp.read())
    if body.get("status") != "Acknowledged":
        sys.exit(1)
for _ in range(50):
    if received["ok"]:
        break
    time.sleep(0.1)
if not received["ok"]:
    sys.exit(1)
PY
echo "Acknowledge persisted and WebSocket alert.acknowledged received"
fi

if smoke_step_enabled REDIS; then
echo "==> 8. Redis feature keys"
if docker ps --format '{{.Names}}' | grep -qx "$REDIS_CONTAINER"; then
  VEL=$(docker exec "$REDIS_CONTAINER" redis-cli KEYS "vel:*" | head -1)
  EXP=$(docker exec "$REDIS_CONTAINER" redis-cli KEYS "exp:*" | head -1)
  LCR=$(docker exec "$REDIS_CONTAINER" redis-cli KEYS "lcr:*" | head -1)
  if [[ -z "$VEL" || -z "$EXP" || -z "$LCR" ]]; then
    echo "Missing Redis feature keys (vel=$VEL exp=$EXP lcr=$LCR)" >&2
    exit 1
  fi
  echo "Redis features: vel=$VEL exp=$EXP lcr=$LCR"
else
  echo "Redis container not found" >&2
  exit 1
fi
fi

if smoke_step_enabled GRAFANA; then
echo "==> 9. Grafana health (optional)"
if curl -sf "$GRAFANA_URL/api/health" >/dev/null 2>&1; then
  echo "Grafana healthy"
else
  echo "Grafana not reachable (non-fatal in dev)"
fi
fi

echo "==> Phase 2 smoke test passed${SMOKE_PHASE2_ONLY:+, only=$SMOKE_PHASE2_ONLY}"
