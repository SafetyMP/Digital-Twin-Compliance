#!/usr/bin/env bash
# Bootstrap or update the deploy Compose stack (GHCR image for state-service).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.deploy.yml}"
ACTION="${1:-bootstrap}"

if [[ -z "${STATE_SERVICE_IMAGE:-}" ]]; then
  echo "ERROR: Set STATE_SERVICE_IMAGE (e.g. ghcr.io/safetymp/digital-twin-compliance/state-service:main)" >&2
  exit 1
fi

export COMPOSE_FILE

case "$ACTION" in
  up)
    docker compose -f "$COMPOSE_FILE" up -d --wait
    ;;
  pull)
    docker compose -f "$COMPOSE_FILE" pull state-service
    docker compose -f "$COMPOSE_FILE" up -d --wait state-service
    ;;
  bootstrap)
    docker compose -f "$COMPOSE_FILE" up -d --wait
    "$ROOT/scripts/seed.sh"
    "$ROOT/scripts/register-schemas.sh"
    "$ROOT/scripts/register-debezium-connector.sh"
    docker compose -f "$COMPOSE_FILE" restart state-service
    docker compose -f "$COMPOSE_FILE" up -d --wait state-service
    ;;
  smoke)
    "$ROOT/scripts/smoke-test.sh"
    ;;
  down)
    docker compose -f "$COMPOSE_FILE" down
    ;;
  *)
    echo "Usage: STATE_SERVICE_IMAGE=... $0 [up|pull|bootstrap|smoke|down]" >&2
    exit 1
    ;;
esac

echo "==> deploy-stack ($ACTION) complete."
