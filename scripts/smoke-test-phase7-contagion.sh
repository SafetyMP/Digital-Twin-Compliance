#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SIM="${SIMULATION_SERVICE_URL:-http://localhost:8094}"
AUDIT="${AUDIT_SERVICE_URL:-http://localhost:8090}"
EVIDENCE_DIR="${ROOT}/evidence"
mkdir -p "$EVIDENCE_DIR"
STAMP="$(date -u +"%Y%m%dT%H%M%SZ")"
OUT="${EVIDENCE_DIR}/phase7-contagion-${STAMP}.txt"
: >"$OUT"

# Keep Phase 4 deterministic path available
curl -sf -X POST "${SIM}/api/v1/simulations/run" \
  -H 'Content-Type: application/json' \
  -d '{"scenarioId":"ecb-adverse-v1","parameters":{"smoke":"phase7-'"${STAMP}"'"}}' \
  | tee -a "$OUT" | jq -e '.runId != null' >/dev/null

body="$(curl -sf -X POST "${SIM}/api/v1/contagion/run" \
  -H 'Content-Type: application/json' \
  -d '{"maxHops":3,"correlationId":"smoke-contagion-'"${STAMP}"'"}')"
echo "$body" | tee -a "$OUT" >/dev/null
echo "$body" | jq -e '.result.infectedCount >= 1 and .evidenceRef != null and (.result.pathExplainability|length)>=1' >/dev/null
run_id="$(echo "$body" | jq -r '.runId')"
for _ in $(seq 1 30); do
  entry="$(curl -sf "${AUDIT}/api/v1/audit/entries?subjectId=${run_id}&limit=5" | jq -r '.[0].entryId // empty')"
  [[ -n "$entry" ]] && break
  sleep 1
done
[[ -n "${entry:-}" ]] || { echo "audit entry missing" >&2; exit 1; }
echo "contagion smoke passed runId=${run_id} auditEntryId=${entry} evidenceRef=$(echo "$body" | jq -r '.evidenceRef')" | tee -a "$OUT" >&2
echo "Phase 7 contagion smoke passed"
