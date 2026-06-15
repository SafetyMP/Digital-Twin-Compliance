#!/usr/bin/env bash
#
# Kafka payload contract tests (publisher + consumer golden fixtures).
#
# Usage: ./scripts/check-kafka-contracts.sh

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

if [[ ! -f contracts/kafka/README.md ]]; then
  echo "contracts/kafka missing" >&2
  exit 1
fi

echo "== Kafka payload contracts =="

echo "-> state-service (twin.state.updated publisher)"
(cd services/state-service && go test ./internal/consumer/ -run 'TestKafkaContract_' -count=1)

echo "-> alert-service (compliance.alerts consumer)"
(cd services/alert-service && go test ./internal/events/ -run 'TestKafkaContract_' -count=1)

echo "-> compliance-cep (Flink consumers + alert publisher shape)"
if command -v mvn >/dev/null 2>&1; then
  (cd jobs/compliance-cep && mvn -q test -Dtest=KafkaContractTest)
else
  docker run --rm \
    -v "$ROOT/jobs/compliance-cep:/app" \
    -v "$ROOT/contracts:/contracts" \
    -w /app \
    maven:3.9-eclipse-temurin-17 \
    mvn -q test -Dtest=KafkaContractTest
fi

echo "Kafka payload contracts passed"
