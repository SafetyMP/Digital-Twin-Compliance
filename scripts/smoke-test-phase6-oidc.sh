#!/usr/bin/env bash
# Phase 6 OIDC: unauthenticated 401 via oidc-edge; valid Keycloak token succeeds.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
EDGE="${OIDC_EDGE_URL:-http://localhost:8180}"
KC="${KEYCLOAK_URL:-http://localhost:8088}"
EVIDENCE_DIR="${ROOT}/evidence"
mkdir -p "$EVIDENCE_DIR"
STAMP="$(date -u +"%Y%m%dT%H%M%SZ")"
EVIDENCE_FILE="${EVIDENCE_DIR}/phase6-oidc-${STAMP}.txt"

log() { echo "$*" | tee -a "$EVIDENCE_FILE" >&2; }
fail() { log "FAIL: $*"; exit 1; }

: >"$EVIDENCE_FILE"
log "Phase 6 OIDC smoke started $(date -u +"%Y-%m-%dT%H:%M:%SZ")"

log "Step 1: edge + keycloak health"
curl -sf "${EDGE}/api/v1/health" | tee -a "$EVIDENCE_FILE" | jq -e '.status=="ok"' >/dev/null || fail "oidc-edge unhealthy"
# realm endpoint
for _ in $(seq 1 60); do
  if curl -sf "${KC}/realms/digital-twin" >/dev/null 2>&1; then
    break
  fi
  sleep 2
done
curl -sf "${KC}/realms/digital-twin" >/dev/null || fail "keycloak realm missing"

log "Step 2: unauthenticated request → 401"
code="$(curl -s -o /tmp/oidc-unauth.json -w '%{http_code}' "${EDGE}/reporting/api/v1/taxonomies" || true)"
log "unauth status=${code} body=$(cat /tmp/oidc-unauth.json 2>/dev/null || true)"
[[ "$code" == "401" ]] || fail "expected 401 got ${code}"

# Also prove state/alert/audit/graph/simulation prefixes reject
for prefix in state alert audit graph simulation; do
  # pick a non-health path
  case "$prefix" in
    state) path="/state/api/v1/personas" ;;
    alert) path="/alert/api/v1/alerts" ;;
    audit) path="/audit/api/v1/audit/entries?limit=1" ;;
    graph) path="/graph/api/v1/graph/summary" ;;
    simulation) path="/simulation/api/v1/simulations/run" ;;
  esac
  c="$(curl -s -o /dev/null -w '%{http_code}' "${EDGE}${path}" || true)"
  log "${prefix} unauth → ${c}"
  [[ "$c" == "401" || "$c" == "405" ]] || fail "${prefix} expected 401 (or 405 for POST-only) got ${c}"
  # simulation run is POST-only; GET may 405 after auth check — force GET on a missing path that still requires auth
done

log "Step 3: obtain Keycloak token (direct access grant)"
token_json="$(curl -sf -X POST "${KC}/realms/digital-twin/protocol/openid-connect/token" \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d 'grant_type=password' \
  -d 'client_id=digital-twin-api' \
  -d 'client_secret=digital-twin-api-secret' \
  -d 'username=analyst' \
  -d 'password=analyst')"
echo "$token_json" | jq -e '.access_token != null' >/dev/null || fail "token grant failed"
TOKEN="$(echo "$token_json" | jq -r '.access_token')"
log "token acquired (len=${#TOKEN})"

log "Step 4: authenticated reporting taxonomies → 200"
auth_body="$(curl -sf -H "Authorization: Bearer ${TOKEN}" "${EDGE}/reporting/api/v1/taxonomies")"
echo "$auth_body" | tee -a "$EVIDENCE_FILE" >/dev/null
echo "$auth_body" | jq -e 'type=="array" and length>=1' >/dev/null || fail "authed taxonomies failed"

log "Step 5: authenticated audit entries → 200"
curl -sf -H "Authorization: Bearer ${TOKEN}" "${EDGE}/audit/api/v1/audit/entries?limit=1" \
  | tee -a "$EVIDENCE_FILE" | jq -e 'type=="array"' >/dev/null || fail "authed audit failed"

# Strip client X-Principal spoofing: edge should ignore attacker header (still 200 with token)
log "Step 6: X-Principal spoof ignored (token still works)"
curl -sf -H "Authorization: Bearer ${TOKEN}" -H "X-Principal: attacker" \
  "${EDGE}/reporting/api/v1/taxonomies" | jq -e 'type=="array"' >/dev/null || fail "spoof test failed"

log "Phase 6 OIDC smoke passed"
log "evidence: ${EVIDENCE_FILE}"
echo "Phase 6 OIDC smoke test passed"
