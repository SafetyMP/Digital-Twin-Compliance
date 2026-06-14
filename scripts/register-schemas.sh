#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REGISTRY="${SCHEMA_REGISTRY_URL:-http://localhost:8081}"

register_schema() {
  local subject="$1"
  local file="$2"
  echo "Registering $subject from $file"
  local body
  body=$(jq -n --arg schema "$(jq -c . "$file")" '{schema: $schema}')
  local code
  code=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
    -H "Content-Type: application/vnd.schemaregistry.v1+json" \
    --data "$body" \
    "$REGISTRY/subjects/${subject}/versions")
  if [[ "$code" != "200" && "$code" != "409" ]]; then
    echo "Schema registration failed for $subject (HTTP $code)" >&2
    exit 1
  fi
}

echo "Waiting for Schema Registry at $REGISTRY..."
for i in $(seq 1 30); do
  if curl -sf "$REGISTRY/subjects" >/dev/null 2>&1; then
    break
  fi
  sleep 2
done

register_schema "domain.events-value" "$ROOT/schemas/avro/event-envelope.avsc"
register_schema "twin.state.updated-value" "$ROOT/schemas/avro/event-envelope.avsc"
register_schema "entity-state-updated" "$ROOT/schemas/avro/entity-state-updated.avsc"
register_schema "twin-state-updated" "$ROOT/schemas/avro/twin-state-updated.avsc"
register_schema "compliance.alerts-value" "$ROOT/schemas/avro/event-envelope.avsc"
register_schema "compliance-alert-raised" "$ROOT/schemas/avro/compliance-alert-raised.avsc"
register_schema "compliance-alert-resolved" "$ROOT/schemas/avro/compliance-alert-resolved.avsc"

echo "Schemas registered."
