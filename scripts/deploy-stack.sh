#!/usr/bin/env bash
# Bootstrap or update the deploy Compose stack (GHCR images).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.deploy.yml}"
ACTION="${1:-bootstrap}"
REGISTRY_PREFIX="${REGISTRY_PREFIX:-ghcr.io/safetymp/digital-twin-compliance}"
IMAGE_TAG="${IMAGE_TAG:-main}"

if [[ -z "${STATE_SERVICE_IMAGE:-}" ]]; then
  echo "ERROR: Set STATE_SERVICE_IMAGE (e.g. ${REGISTRY_PREFIX}/state-service:${IMAGE_TAG})" >&2
  echo "       See docs/deployment.md for all *_IMAGE variables." >&2
  exit 1
fi

export COMPOSE_FILE

case "$ACTION" in
  up)
    docker compose -f "$COMPOSE_FILE" up -d --wait
    ;;
  pull)
    docker compose -f "$COMPOSE_FILE" pull
    docker compose -f "$COMPOSE_FILE" up -d --wait
    ;;
  bootstrap)
    docker compose -f "$COMPOSE_FILE" up -d --wait
    "$ROOT/scripts/seed.sh"
    "$ROOT/scripts/register-schemas.sh"
    "$ROOT/scripts/register-debezium-connector.sh"
    "$ROOT/scripts/create-kafka-topics.sh"
    docker compose -f "$COMPOSE_FILE" restart \
      state-service alert-service audit-service cedar-service decision-service
    docker compose -f "$COMPOSE_FILE" up -d --wait \
      state-service alert-service audit-service cedar-service decision-service
    ;;
  smoke)
    "$ROOT/scripts/smoke-test.sh"
    if [[ -n "${ALERT_SERVICE_IMAGE:-}" ]]; then
      "$ROOT/scripts/smoke-test-phase2.sh"
    fi
    if [[ -n "${AUDIT_SERVICE_IMAGE:-}" ]]; then
      "$ROOT/scripts/smoke-test-phase3.sh"
    fi
    ;;
  down)
    docker compose -f "$COMPOSE_FILE" down
    ;;
  *)
    echo "Usage: STATE_SERVICE_IMAGE=... [ALERT_* / AUDIT_* / CEDAR_* / DECISION_* / COMPLIANCE_CEP_*] $0 [up|pull|bootstrap|smoke|down]" >&2
    exit 1
    ;;
esac

echo "==> deploy-stack ($ACTION) complete."
