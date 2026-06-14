#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

CORE_URL="${CORE_BANKING_DB_URL:-postgres://core:core@localhost:5433/core_banking?sslmode=disable}"
STATE_URL="${STATE_DB_URL:-postgres://state:state@localhost:5434/twin_state?sslmode=disable}"

echo "==> Applying core banking migrations..."
psql "$CORE_URL" -v ON_ERROR_STOP=1 -f mocks/core-banking/migrations/001_source_tables.sql

ENTITY_COUNT=$(psql "$CORE_URL" -tA -c "SELECT COUNT(*) FROM legal_entities" 2>/dev/null || echo "0")
if [[ "${ENTITY_COUNT:-0}" -lt 10 ]]; then
  echo "==> Seeding core banking data..."
  psql "$CORE_URL" -v ON_ERROR_STOP=1 -f mocks/core-banking/seed/seed.sql
else
  echo "==> Core banking already seeded ($ENTITY_COUNT institutions); skipping seed."
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

echo "==> Seed complete."
