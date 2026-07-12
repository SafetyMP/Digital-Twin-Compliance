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
    maven:3.9-eclipse-temurin-17@sha256:1ed5d1f54416b706707b4f3238f63a20bb06aab27c6d240090a2bb9ad895ed45 \
    mvn -q test -Dtest=KafkaContractTest
fi

echo "-> audit-service (compliance.audit.pending consumer)"
(cd services/audit-service && go test ./internal/events/ -run 'TestKafkaContract_' -count=1)

echo "-> alert-service (compliance.audit.pending publisher)"
(cd services/alert-service && go test ./internal/audit/ -run 'TestKafkaContract_' -count=1)

echo "-> cedar-service (rule-decision + audit envelope)"
(cd services/cedar-service && go test ./internal/audit/ -run 'TestKafkaContract_' -count=1)

echo "-> decision-service (compliance.audit.pending publisher)"
(cd services/decision-service && go test ./internal/audit/ -run 'TestKafkaContract_' -count=1)

echo "Kafka payload contracts passed"

