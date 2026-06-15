#!/usr/bin/env bash
set -euo pipefail

CONNECT_URL="${DEBEZIUM_CONNECT_URL:-http://localhost:8083}"
CONNECTOR_NAME="core-banking-cdc"

CONFIG=$(cat <<'EOF'
{
  "name": "core-banking-cdc",
  "config": {
    "connector.class": "io.debezium.connector.postgresql.PostgresConnector",
    "database.hostname": "core-banking-db",
    "database.port": "5432",
    "database.user": "core",
    "database.password": "core",
    "database.dbname": "core_banking",
    "database.server.name": "domain.events",
    "plugin.name": "pgoutput",
    "publication.name": "dbz_publication",
    "slot.name": "debezium_core_banking",
    "table.include.list": "public.legal_entities,public.accounts,public.instruments,public.payments",
    "topic.prefix": "domain.events",
    "key.converter": "org.apache.kafka.connect.json.JsonConverter",
    "value.converter": "org.apache.kafka.connect.json.JsonConverter",
    "key.converter.schemas.enable": "false",
    "value.converter.schemas.enable": "false",
    "snapshot.mode": "initial"
  }
}
EOF
)

if curl -sf "$CONNECT_URL/connectors/$CONNECTOR_NAME" >/dev/null 2>&1; then
  echo "Connector $CONNECTOR_NAME already exists; updating..."
  echo "$CONFIG" | jq '.config' | curl -sf -X PUT \
    -H "Content-Type: application/json" \
    -d @- \
    "$CONNECT_URL/connectors/$CONNECTOR_NAME/config" >/dev/null
else
  echo "Creating connector $CONNECTOR_NAME..."
  echo "$CONFIG" | curl -sf -X POST \
    -H "Content-Type: application/json" \
    -d @- \
    "$CONNECT_URL/connectors" >/dev/null
fi

echo "Debezium connector registered."
