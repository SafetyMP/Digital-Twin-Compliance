#!/usr/bin/env bash
# Phase 4 end-to-end: graph seed, simulation run, audit linkage, Phase 3 regression.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
chmod +x "$ROOT/scripts/wait-graph-seeded.sh" "$ROOT/scripts/verify-audit-chain.sh" 2>/dev/null || true

GRAPH_URL="${GRAPH_SERVICE_URL:-http://localhost:8093}"
SIM_URL="${SIMULATION_SERVICE_URL:-http://localhost:8094}"
GRAPH_UI="${GRAPH_EXPLORER_URL:-http://localhost:3003}"
SIM_UI="${SIMULATION_CONSOLE_URL:-http://localhost:3004}"
AUDIT_URL="${AUDIT_SERVICE_URL:-http://localhost:8090}"
NEO4J_HTTP="${NEO4J_HTTP_URL:-http://localhost:7474}"

SMOKE_PHASE4_STARTED_AT="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"

phase4_fail() {
  echo "=== Phase 4 smoke failure diagnostics ===" >&2
  curl -sf "${GRAPH_URL}/api/v1/graph/summary" | jq . 2>/dev/null || true
  curl -sf "${AUDIT_URL}/api/v1/audit/entries?limit=5" | jq . 2>/dev/null || true
  exit 1
}

if [[ "${SMOKE_PHASE4_SKIP_PREREQS:-}" != "1" ]]; then
  echo "Running Phase 3 regression prerequisite..."
  SMOKE_PHASE3_SKIP_PREREQS=1 "$ROOT/scripts/smoke-test-phase3.sh"
fi

echo "Step 1: Phase 4 service health"
curl -sf "${NEO4J_HTTP}/" >/dev/null || { echo "neo4j browser unreachable" >&2; phase4_fail; }
for url in "$GRAPH_URL" "$SIM_URL"; do
  curl -sf "${url}/api/v1/health" >/dev/null || { echo "unhealthy: $url" >&2; phase4_fail; }
done
for url in "$GRAPH_UI" "$SIM_UI"; do
  curl -sf "${url}/" >/dev/null || { echo "ui unreachable: $url" >&2; phase4_fail; }
done

echo "Step 2: Wait for graph seed"
"$ROOT/scripts/wait-graph-seeded.sh"

echo "Step 3: Graph API returns seed institution"
nodes=$(curl -sf "${GRAPH_URL}/api/v1/graph/nodes?name=Delta")
echo "$nodes" | jq -e '[.[] | select(.name | test("Delta Independent Bank"; "i"))] | length > 0' >/dev/null || phase4_fail

echo "Step 4: Run simulation scenario"
CORRELATION_ID="smoke-phase4-$(date +%s)"
run=$(curl -sf -X POST "${SIM_URL}/api/v1/simulations/run" \
  -H 'Content-Type: application/json' \
  -d "{\"scenarioId\":\"ecb-adverse-v1\",\"correlationId\":\"${CORRELATION_ID}\",\"parameters\":{\"smoke\":\"${CORRELATION_ID}\"}}")
echo "$run" | jq -e '.runId != null and .metrics.stressedCet1 != null' >/dev/null || phase4_fail
echo "$run" | jq -e '[.decisions[] | select(.ruleCode == "COREP-R001")][0].outcome == "Deny"' >/dev/null || phase4_fail
echo "$run" | jq -e '[.decisions[] | select(.ruleCode == "COREP-R002")][0].outcome == "Deny"' >/dev/null || phase4_fail
run_id=$(echo "$run" | jq -r '.runId')

echo "Step 5: Audit Service verify SimulationRun entry"
for _ in $(seq 1 30); do
  entry_id=$(curl -sf "${AUDIT_URL}/api/v1/audit/entries?subjectId=${run_id}&limit=10" 2>/dev/null | \
    jq -r '[.[] | select(.entryType=="SimulationRun")][0].entryId // empty' 2>/dev/null || true)
  if [[ -n "$entry_id" ]]; then
    break
  fi
  sleep 1
done
if [[ -z "$entry_id" ]]; then
  echo "SimulationRun audit entry not found for runId=${run_id}" >&2
  phase4_fail
fi
"$ROOT/scripts/verify-audit-chain.sh"

echo "Phase 4 smoke test passed"
