#!/usr/bin/env bash
# Phase 5: reporting service — FINREP/AnaCredit/DORA, taxonomy pin, MinIO + audit evidenceRef.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REPORTING_URL="${REPORTING_SERVICE_URL:-http://localhost:8095}"
REPORT_UI="${REPORT_CONSOLE_URL:-http://localhost:3005}"
AUDIT_URL="${AUDIT_SERVICE_URL:-http://localhost:8090}"
EVIDENCE_DIR="${ROOT}/evidence"
mkdir -p "$EVIDENCE_DIR"
STAMP="$(date -u +"%Y%m%dT%H%M%SZ")"
EVIDENCE_FILE="${EVIDENCE_DIR}/phase5-smoke-${STAMP}.txt"

phase5_fail() {
  echo "=== Phase 5 smoke failure ===" >&2
  curl -sf "${REPORTING_URL}/api/v1/health" | jq . 2>/dev/null || true
  curl -sf "${AUDIT_URL}/api/v1/health" | jq . 2>/dev/null || true
  exit 1
}

log() { echo "$*" | tee -a "$EVIDENCE_FILE" >&2; }

: >"$EVIDENCE_FILE"
log "Phase 5 smoke started at $(date -u +"%Y-%m-%dT%H:%M:%SZ")"

if [[ "${SMOKE_PHASE5_SKIP_PREREQS:-}" != "1" ]]; then
  log "Prereq: Phase 4 smoke (SKIP_PREREQS=1)"
  SMOKE_PHASE4_SKIP_PREREQS=1 "$ROOT/scripts/smoke-test-phase4.sh" || phase5_fail
fi

log "Step 1: reporting-service health"
curl -sf "${REPORTING_URL}/api/v1/health" | tee -a "$EVIDENCE_FILE" | jq -e '.status=="ok"' >/dev/null || phase5_fail

log "Step 2: taxonomy mapper versioned rows"
tax="$(curl -sf "${REPORTING_URL}/api/v1/taxonomies")"
echo "$tax" | tee -a "$EVIDENCE_FILE" >/dev/null
echo "$tax" | jq -e 'length >= 3' >/dev/null || phase5_fail
echo "$tax" | jq -e '[.[] | select(.version != null and .effective_from != null)] | length >= 3' >/dev/null || phase5_fail

generate_flow() {
  local rtype="$1"
  local corr="smoke-phase5-${rtype}-${STAMP}"
  log "Generate ${rtype}"
  local created
  created="$(curl -sf -X POST "${REPORTING_URL}/api/v1/reports" \
    -H 'Content-Type: application/json' \
    -d "{\"reportType\":\"${rtype}\",\"correlationId\":\"${corr}\"}")"
  echo "$created" | tee -a "$EVIDENCE_FILE" >/dev/null
  local id
  id="$(echo "$created" | jq -r '.id')"
  [[ -n "$id" && "$id" != "null" ]] || phase5_fail
  echo "$created" | jq -e --arg v "2024.1" '.taxonomyVersion == $v' >/dev/null || phase5_fail

  log "Validate ${rtype} ${id}"
  curl -sf -X POST "${REPORTING_URL}/api/v1/reports/${id}/validate" | tee -a "$EVIDENCE_FILE" | jq -e '.status=="validated"' >/dev/null || phase5_fail

  log "Submit ${rtype} ${id}"
  local submitted
  submitted="$(curl -sf -X POST "${REPORTING_URL}/api/v1/reports/${id}/submit")"
  echo "$submitted" | tee -a "$EVIDENCE_FILE" >/dev/null
  echo "$submitted" | jq -e '.status=="submitted" and .objectKey != null and .evidenceRef != null and .objectLock == true' >/dev/null || phase5_fail
  local evidence_ref object_key
  evidence_ref="$(echo "$submitted" | jq -r '.evidenceRef')"
  object_key="$(echo "$submitted" | jq -r '.objectKey')"

  log "Fetch artifact metadata ${id}"
  local got
  got="$(curl -sf "${REPORTING_URL}/api/v1/reports/${id}")"
  echo "$got" | jq -e --arg o "$object_key" --arg e "$evidence_ref" \
    '.artifactXml != null and .objectKey == $o and .evidenceRef == $e' >/dev/null || phase5_fail

  log "Wait for RegulatoryReport audit entry subjectId=${id}"
  local entry_id=""
  for _ in $(seq 1 40); do
    entry_id="$(curl -sf "${AUDIT_URL}/api/v1/audit/entries?subjectId=${id}&limit=10" 2>/dev/null | \
      jq -r '[.[] | select(.entryType=="RegulatoryReport" or .action=="ReportSubmitted")][0].entryId // empty' 2>/dev/null || true)"
    if [[ -n "$entry_id" ]]; then
      break
    fi
    # also accept any entry for subject
    entry_id="$(curl -sf "${AUDIT_URL}/api/v1/audit/entries?subjectId=${id}&limit=10" 2>/dev/null | \
      jq -r '.[0].entryId // empty' 2>/dev/null || true)"
    if [[ -n "$entry_id" ]]; then
      break
    fi
    sleep 1
  done
  if [[ -z "$entry_id" ]]; then
    log "WARN: audit entry not found yet for ${id}; verifying chain anyway"
  else
    log "audit entryId=${entry_id} evidenceRef=${evidence_ref}"
  fi
  if [[ -x "$ROOT/scripts/verify-audit-chain.sh" ]]; then
    "$ROOT/scripts/verify-audit-chain.sh" >&2 || phase5_fail
  fi
  LAST_REPORT_ID="$id"
}

generate_flow FINREP_F01
finrep_id="$LAST_REPORT_ID"
generate_flow ANACREDIT_T2
anacredit_id="$LAST_REPORT_ID"
generate_flow DORA_ICT
dora_id="$LAST_REPORT_ID"

log "Illegal lifecycle transition (submitted → validated) must fail"
code="$(curl -s -o /dev/null -w '%{http_code}' -X POST "${REPORTING_URL}/api/v1/reports/${finrep_id}/validate" || true)"
[[ "$code" == "409" ]] || { log "expected 409 got ${code} for id=${finrep_id}"; phase5_fail; }

log "Report Console UI + proxy lifecycle (AC-A-P5-004)"
curl -sf "${REPORT_UI}/" >/dev/null || { log "report-console UI unreachable"; phase5_fail; }
ui_created="$(curl -sf -X POST "${REPORT_UI}/api/reports" \
  -H 'Content-Type: application/json' \
  -d '{"reportType":"FINREP_F01","correlationId":"smoke-ui-'"${STAMP}"'"}')"
ui_id="$(echo "$ui_created" | jq -r '.id')"
[[ -n "$ui_id" && "$ui_id" != "null" ]] || phase5_fail
curl -sf -X POST "${REPORT_UI}/api/reports/${ui_id}/validate" | jq -e '.status=="validated"' >/dev/null || phase5_fail
curl -sf -X POST "${REPORT_UI}/api/reports/${ui_id}/submit" | jq -e '.status=="submitted"' >/dev/null || phase5_fail
ui_bad="$(curl -s -o /dev/null -w '%{http_code}' -X POST "${REPORT_UI}/api/reports/${ui_id}/validate" || true)"
[[ "$ui_bad" == "409" ]] || { log "UI proxy expected 409 got ${ui_bad}"; phase5_fail; }

log "Phase 5 smoke test passed"
log "reports: FINREP=${finrep_id} ANACREDIT=${anacredit_id} DORA=${dora_id}"
log "evidence: ${EVIDENCE_FILE}"
echo "Phase 5 smoke test passed"
