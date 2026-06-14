#!/usr/bin/env bash
set -euo pipefail

FLINK_URL="${FLINK_JOBMANAGER_URL:-http://localhost:8082}"
JAR_PATH="${FLINK_JAR_PATH:-jobs/compliance-cep/target/compliance-cep-0.1.0-SNAPSHOT.jar}"
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if [[ ! -f "$ROOT/$JAR_PATH" ]]; then
  echo "Building Flink job..."
  if command -v mvn >/dev/null 2>&1; then
    (cd "$ROOT/jobs/compliance-cep" && mvn -q package -DskipTests)
  else
    docker run --rm -v "$ROOT/jobs/compliance-cep:/app" -w /app maven:3.9-eclipse-temurin-17 mvn -q package -DskipTests
  fi
fi

JAR_PATH="$ROOT/$JAR_PATH"

echo "Waiting for Flink JobManager at $FLINK_URL..."
for i in $(seq 1 60); do
  if curl -sf "$FLINK_URL/overview" >/dev/null 2>&1; then
    break
  fi
  sleep 2
done

echo "Cancelling existing RUNNING compliance-cep jobs (if any)..."
EXISTING=$(curl -sf "${FLINK_URL}/jobs/overview" 2>/dev/null | jq -r '.jobs[] | select(.name=="compliance-cep" and .state=="RUNNING") | .jid' || true)
for jid in $EXISTING; do
  echo "Cancelling job $jid"
  curl -sf -X PATCH "${FLINK_URL}/jobs/${jid}?mode=cancel" >/dev/null || true
done

echo "Uploading JAR..."
UPLOAD=$(curl -sf -X POST -H "Expect:" -F "jarfile=@${JAR_PATH}" "${FLINK_URL}/jars/upload")
JAR_ID=$(basename "$(echo "$UPLOAD" | jq -r '.filename')")

PROGRAM_ARGS="--kafka kafka:9092 --redisHost redis --redisPort 6379 --tenantId 00000000-0000-0000-0000-000000000001 --velocityMax ${CEP_VELOCITY_MAX_PER_HOUR:-50} --exposureLimit ${CEP_EXPOSURE_LIMIT_EUR:-10000000} --lcrMinimum ${CEP_LCR_MINIMUM:-1.0} --parallelism ${FLINK_PARALLELISM:-1}"

echo "Submitting job from $JAR_ID..."
RUN=$(curl -sf -X POST "${FLINK_URL}/jars/${JAR_ID}/run" \
  -H "Content-Type: application/json" \
  -d "{\"entryClass\":\"com.digitaltwin.jobs.cep.ComplianceCepJob\",\"programArgs\":\"${PROGRAM_ARGS}\",\"parallelism\":${FLINK_PARALLELISM:-1}}")

echo "$RUN" | jq .
echo "Flink job submitted."
