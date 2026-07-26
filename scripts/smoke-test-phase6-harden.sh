#!/usr/bin/env bash
# Phase 6 harden: TLS edge, OTel cross-service trace, DR/explainability docs, CI load residual.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TLS_URL="${TLS_EDGE_URL:-https://localhost:8443}"
JAEGER="${JAEGER_URL:-http://localhost:16686}"
EDGE="${OIDC_EDGE_URL:-http://localhost:8180}"
KC="${KEYCLOAK_URL:-http://localhost:8088}"
EVIDENCE_DIR="${ROOT}/evidence"
mkdir -p "$EVIDENCE_DIR"
STAMP="$(date -u +"%Y%m%dT%H%M%SZ")"
EVIDENCE_FILE="${EVIDENCE_DIR}/phase6-harden-${STAMP}.txt"

log() { echo "$*" | tee -a "$EVIDENCE_FILE" >&2; }
fail() { log "FAIL: $*"; exit 1; }

: >"$EVIDENCE_FILE"
log "Phase 6 harden smoke started $(date -u +"%Y-%m-%dT%H:%M:%SZ")"

log "Step 1: TLS edge health (self-signed → curl -k)"
curl -skf "${TLS_URL}/healthz" | tee -a "$EVIDENCE_FILE" | grep -q ok || fail "tls edge"
echo "tls_ok=1" >>"$EVIDENCE_FILE"
cp /dev/null "${EVIDENCE_DIR}/phase6-tls-${STAMP}.txt"
echo "curl -k ${TLS_URL}/healthz => ok" >"${EVIDENCE_DIR}/phase6-tls-${STAMP}.txt"

log "Step 2: docs — explainability, deployment residuals, DR runbooks"
for f in \
  docs/explainability.md \
  docs/deployment.md \
  docs/runbooks/kafka-recovery.md \
  docs/runbooks/flink-recovery.md \
  docs/runbooks/immudb-recovery.md \
  scripts/load-profile/README.md; do
  test -f "${ROOT}/${f}" || fail "missing ${f}"
done
grep -q "Residual" "${ROOT}/docs/deployment.md" || fail "deployment residuals missing"

log "Step 3: CI-sized load profile (honest residual)"
chmod +x "${ROOT}/scripts/load-profile/run-ci-sized.sh"
"${ROOT}/scripts/load-profile/run-ci-sized.sh" | tee -a "$EVIDENCE_FILE"
grep -q "residual=full_10k_not_run_on_ci_hosts" "${EVIDENCE_DIR}"/phase6-load-*.txt || fail "load residual missing"

log "Step 4: generate cross-service traffic for OTel (edge → reporting)"
TOKEN="$(curl -sf -X POST "${KC}/realms/digital-twin/protocol/openid-connect/token" \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d 'grant_type=password' \
  -d 'client_id=digital-twin-api' \
  -d 'client_secret=digital-twin-api-secret' \
  -d 'username=analyst' \
  -d 'password=analyst' | jq -r '.access_token')"
[[ -n "$TOKEN" && "$TOKEN" != "null" ]] || fail "token"
for _ in $(seq 1 5); do
  curl -sf -H "Authorization: Bearer ${TOKEN}" "${EDGE}/reporting/api/v1/taxonomies" >/dev/null
  curl -skf -H "Authorization: Bearer ${TOKEN}" "${TLS_URL}/reporting/api/v1/taxonomies" >/dev/null || true
done
sleep 3

log "Step 5: Jaeger has traces for oidc-edge and/or reporting-service"
otel_ok=0
for svc in oidc-edge reporting-service; do
  body="$(curl -sf "${JAEGER}/api/traces?service=${svc}&limit=5" || true)"
  echo "$body" | tee -a "$EVIDENCE_FILE" >/dev/null
  if echo "$body" | jq -e '.data | length >= 1' >/dev/null 2>&1; then
    log "traces found for ${svc}"
    otel_ok=1
  fi
done
[[ "$otel_ok" == "1" ]] || fail "no otel traces in jaeger"
echo "otel_ok=1" >"${EVIDENCE_DIR}/phase6-otel-${STAMP}.txt"

log "Phase 6 harden smoke passed"
echo "Phase 6 harden smoke test passed"
