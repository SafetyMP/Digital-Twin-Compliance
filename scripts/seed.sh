#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

CORE_URL="${CORE_BANKING_DB_URL:-postgres://core:core@localhost:5433/core_banking?sslmode=disable}"
STATE_URL="${STATE_DB_URL:-postgres://state:state@localhost:5434/twin_state?sslmode=disable}"

echo "==> Applying core banking migrations..."
psql "$CORE_URL" -v ON_ERROR_STOP=1 -f mocks/core-banking/migrations/001_source_tables.sql
psql "$CORE_URL" -v ON_ERROR_STOP=1 -f mocks/core-banking/migrations/002_payments.sql

ENTITY_COUNT=$(psql "$CORE_URL" -tA -c "SELECT COUNT(*) FROM legal_entities" 2>/dev/null || echo "0")
if [[ "${ENTITY_COUNT:-0}" -lt 10 ]]; then
  echo "==> Seeding core banking data..."
  psql "$CORE_URL" -v ON_ERROR_STOP=1 -f mocks/core-banking/seed/seed.sql
  psql "$CORE_URL" -v ON_ERROR_STOP=1 -f mocks/core-banking/seed/002_phase2_exposure.sql
else
  echo "==> Core banking already seeded ($ENTITY_COUNT institutions); skipping seed."
  psql "$CORE_URL" -v ON_ERROR_STOP=1 -f mocks/core-banking/seed/002_phase2_exposure.sql 2>/dev/null || true
fi

ALERT_URL="${ALERT_DB_URL:-postgres://alert:alert@localhost:5435/alerts?sslmode=disable}"
if psql "$ALERT_URL" -c "SELECT 1" >/dev/null 2>&1; then
  echo "==> Applying alert store migrations..."
  psql "$ALERT_URL" -v ON_ERROR_STOP=1 -f services/alert-service/migrations/001_alerts.sql
fi

echo "==> Applying state store migrations..."
psql "$STATE_URL" -v ON_ERROR_STOP=1 -f services/state-service/migrations/001_init.sql

echo "==> Registering Avro schemas..."
if curl -sf "${SCHEMA_REGISTRY_URL:-http://localhost:8081}/subjects" >/dev/null 2>&1; then
  "$ROOT/scripts/register-schemas.sh"
fi

echo "==> Registering Debezium connector..."
if curl -sf "${DEBEZIUM_CONNECT_URL:-http://localhost:8083}/connectors" >/dev/null 2>&1; then
  "$ROOT/scripts/register-debezium-connector.sh"
fi

echo "==> Ensuring Kafka topics..."
if docker ps --format '{{.Names}}' | grep -qx "${KAFKA_CONTAINER:-digitaltwin-kafka-1}"; then
  "$ROOT/scripts/create-kafka-topics.sh"
fi

echo "==> Seed complete."
