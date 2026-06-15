#!/usr/bin/env bash
# Canary: core CDC -> state-service -> twin_personas before Phase 2 smoke (INT-M002 / BASEL-M001).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

STATE_URL="${STATE_DB_URL:-postgres://state:state@localhost:5434/twin_state?sslmode=disable}"
CORE_URL="${CORE_BANKING_DB_URL:-postgres://core:core@localhost:5433/core_banking?sslmode=disable}"
STATE_HTTP="${STATE_SERVICE_URL:-http://localhost:8080}"
CONNECT_URL="${DEBEZIUM_CONNECT_URL:-http://localhost:8083}"
REDIS_CONTAINER="${REDIS_CONTAINER:-digitaltwin-redis-1}"
KAFKA_CONTAINER="${KAFKA_CONTAINER:-digitaltwin-kafka-1}"
TENANT_ID="${DEFAULT_TENANT_ID:-00000000-0000-0000-0000-000000000001}"
CANARY_WAIT_SEC="${STATE_TWIN_CANARY_WAIT_SEC:-60}"
OUTBOX_CANARY_WAIT_SEC="${STATE_OUTBOX_CANARY_WAIT_SEC:-30}"

# Delta Independent Bank — seeded institution with liquidity columns (BASEL-M001 path).
CANARY_ENTITY="${STATE_TWIN_CANARY_ENTITY:-44444444-4444-4444-4444-444444444401}"

if ! docker ps --format '{{.Names}}' | grep -qx "$REDIS_CONTAINER"; then
  DETECTED=$(docker ps --format '{{.Names}}' | grep -E 'redis-1$' | head -1 || true)
  if [[ -n "$DETECTED" ]]; then
    REDIS_CONTAINER="$DETECTED"
  fi
fi

# shellcheck source=scripts/smoke-lib-phase2.sh
source "$ROOT/scripts/smoke-lib-phase2.sh"

echo "==> Verify state twin pipeline (entity=$CANARY_ENTITY)"

curl -sf "$STATE_HTTP/api/v1/health" | jq -e '.status == "ok"' >/dev/null

CONNECTOR_STATE=$(curl -sf "$CONNECT_URL/connectors/core-banking-cdc/status" | jq -r '.connector.state // "UNKNOWN"')
TASK_STATE=$(curl -sf "$CONNECT_URL/connectors/core-banking-cdc/status" | jq -r '.tasks[0].state // "UNKNOWN"')
if [[ "$CONNECTOR_STATE" != "RUNNING" || "$TASK_STATE" != "RUNNING" ]]; then
  echo "Debezium not RUNNING (connector=$CONNECTOR_STATE task=$TASK_STATE)" >&2
  smoke_twin_pipeline_debug
  exit 1
fi

BEFORE_VERSION="$(twin_state_version "$CANARY_ENTITY")"
if [[ -z "$BEFORE_VERSION" ]]; then
  BEFORE_VERSION="0"
fi

OUTBOX_ID_BEFORE="$(max_outbox_id)"
if [[ -z "$OUTBOX_ID_BEFORE" ]]; then
  OUTBOX_ID_BEFORE="0"
fi

psql "$CORE_URL" -v ON_ERROR_STOP=1 -c "
  UPDATE legal_entities
  SET updated_at = now()
  WHERE entity_id = '$CANARY_ENTITY';
" >/dev/null

MIN_VERSION="$BEFORE_VERSION"

if ! wait_twin_state_version_gt "$CANARY_ENTITY" "$MIN_VERSION" "$CANARY_WAIT_SEC" "state twin canary"; then
  echo "--- state twin pipeline debug ---" >&2
  psql "$CORE_URL" -tA -c "SELECT entity_id, lcr, updated_at FROM legal_entities WHERE entity_id = '$CANARY_ENTITY';" 2>/dev/null >&2 || true
  psql "$STATE_URL" -tA -c "SELECT persona_id, state_version, left(current_state::text, 200) FROM twin_personas WHERE persona_id = '$CANARY_ENTITY';" 2>/dev/null >&2 || true
  smoke_twin_pipeline_debug
  exit 1
fi

if ! wait_outbox_published_after_id "$OUTBOX_ID_BEFORE" "$OUTBOX_CANARY_WAIT_SEC" "canary outbox"; then
  smoke_twin_pipeline_debug
  exit 1
fi

echo "State twin pipeline ok (entity=$CANARY_ENTITY)"
