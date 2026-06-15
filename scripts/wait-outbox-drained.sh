#!/usr/bin/env bash
# Wait for state-service outbox backlog to drain after restart (CI bring-up catch-up).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
STATE_URL="${STATE_DB_URL:-postgres://state:state@localhost:5434/twin_state?sslmode=disable}"
OUTBOX_DRAIN_WAIT_SEC="${STATE_OUTBOX_DRAIN_WAIT_SEC:-120}"

# shellcheck source=scripts/smoke-lib-phase2.sh
source "$ROOT/scripts/smoke-lib-phase2.sh"

echo "==> Wait for outbox drain (timeout=${OUTBOX_DRAIN_WAIT_SEC}s)"
pending="$(psql "${STATE_URL}" -tA -c "SELECT COUNT(*) FROM outbox WHERE published_at IS NULL;" 2>/dev/null | tr -d '[:space:]')"
echo "unpublished outbox rows at start: ${pending:-?}"

if ! wait_outbox_drained "$OUTBOX_DRAIN_WAIT_SEC"; then
  smoke_twin_pipeline_debug
  exit 1
fi
