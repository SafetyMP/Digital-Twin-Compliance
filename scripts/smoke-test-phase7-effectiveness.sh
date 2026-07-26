#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SIM="${SIMULATION_SERVICE_URL:-http://localhost:8094}"
AUDIT="${AUDIT_SERVICE_URL:-http://localhost:8090}"
EVIDENCE_DIR="${ROOT}/evidence"
mkdir -p "$EVIDENCE_DIR"
STAMP="$(date -u +"%Y%m%dT%H%M%SZ")"
OUT="${EVIDENCE_DIR}/phase7-effectiveness-${STAMP}.txt"

log() { echo "$*" | tee -a "$OUT" >&2; }
: >"$OUT"

body="$(curl -sf -X POST "${SIM}/api/v1/effectiveness/replay" \
  -H 'Content-Type: application/json' \
  -d '{"detections":["INT-M001","INT-M002","BASEL-M001"],"correlationId":"smoke-eff-'"${STAMP}"'"}')"
echo "$body" | tee -a "$OUT" >/dev/null
echo "$body" | jq -e '.metrics.coverage > 0 and .metrics.missRate > 0 and .evidenceRef != null' >/dev/null
run_id="$(echo "$body" | jq -r '.runId')"
for _ in $(seq 1 30); do
  entry="$(curl -sf "${AUDIT}/api/v1/audit/entries?subjectId=${run_id}&limit=5" | jq -r '.[0].entryId // empty')"
  [[ -n "$entry" ]] && break
  sleep 1
done
[[ -n "${entry:-}" ]] || { log "audit entry missing"; exit 1; }
log "effectiveness smoke passed runId=${run_id} evidenceRef=$(echo "$body" | jq -r '.evidenceRef')"
echo "Phase 7 effectiveness smoke passed"
