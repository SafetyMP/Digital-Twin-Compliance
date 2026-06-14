#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BASE="${STATE_SERVICE_URL:-http://localhost:8080}"
CORE_URL="${CORE_BANKING_DB_URL:-postgres://core:core@localhost:5433/core_banking?sslmode=disable}"

echo "==> Waiting for state service..."
for i in $(seq 1 60); do
  if curl -sf "$BASE/api/v1/health" >/dev/null 2>&1; then
    break
  fi
  sleep 2
done

echo "==> Waiting for initial CDC sync (>=10 institutions)..."
DEADLINE=$((SECONDS + 120))
COUNT=0
while [[ "$SECONDS" -lt "$DEADLINE" ]]; do
  COUNT=$(curl -sf "$BASE/api/v1/personas?personaType=Institution&limit=200" | jq 'length')
  if [[ "$COUNT" -ge 10 ]]; then
    echo "Synced institution personas: $COUNT"
    break
  fi
  sleep 2
done

echo "==> 1. Health check"
HEALTH=$(curl -sf "$BASE/api/v1/health")
echo "$HEALTH" | grep -q '"status":"ok"'

echo "==> 2. Seeded institution personas"
echo "Institution count: $COUNT"
[[ "$COUNT" -ge 10 ]]

echo "==> 3. CDC -> state update test"
ENTITY_ID=$(psql "$CORE_URL" -tA -c "SELECT entity_id FROM legal_entities LIMIT 1")
BEFORE=$(curl -sf "$BASE/api/v1/personas/$ENTITY_ID" | jq '.stateVersion')
echo "Entity $ENTITY_ID state_version before: $BEFORE"

psql "$CORE_URL" -v ON_ERROR_STOP=1 -c \
  "UPDATE legal_entities SET legal_name = legal_name || ' (sync test)', updated_at = now() WHERE entity_id = '$ENTITY_ID';"

DEADLINE=$((SECONDS + 30))
AFTER="$BEFORE"
while [[ "$SECONDS" -lt "$DEADLINE" ]]; do
  AFTER=$(curl -sf "$BASE/api/v1/personas/$ENTITY_ID" | jq '.stateVersion')
  if [[ "$AFTER" -gt "$BEFORE" ]]; then
    break
  fi
  sleep 1
done

echo "state_version after: $AFTER"
[[ "$AFTER" -gt "$BEFORE" ]]

echo "==> Smoke test passed."
