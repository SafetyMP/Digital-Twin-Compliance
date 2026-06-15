#!/usr/bin/env bash
# Pre-create Phase 2 Kafka topics (Debezium may skip empty tables; Flink needs these at startup).
set -euo pipefail

KAFKA_CONTAINER="${KAFKA_CONTAINER:-digitaltwin-kafka-1}"
BOOTSTRAP="${KAFKA_BOOTSTRAP:-localhost:9092}"
PARTITIONS="${KAFKA_TOPIC_PARTITIONS:-3}"
REPLICATION="${KAFKA_TOPIC_REPLICATION:-1}"

if ! docker ps --format '{{.Names}}' | grep -qx "$KAFKA_CONTAINER"; then
  echo "Kafka container $KAFKA_CONTAINER not running; skipping topic creation."
  exit 0
fi

create_topic() {
  local topic="$1"
  docker exec "$KAFKA_CONTAINER" /opt/kafka/bin/kafka-topics.sh \
    --bootstrap-server "$BOOTSTRAP" \
    --create --if-not-exists \
    --topic "$topic" \
    --partitions "$PARTITIONS" \
    --replication-factor "$REPLICATION" >/dev/null
  echo "Ensured Kafka topic $topic exists (${PARTITIONS} partitions)."
}

create_topic "domain.events.public.payments"
create_topic "compliance.alerts"
create_topic "compliance.alerts.dlq"
create_topic "twin.state.updated"
