#!/usr/bin/env bash
# Site adversarial oracle: hermetic CECT probes always; tier-3 cedar auth deny when stack is up.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"

test -f .corp-harness/site.json
test -f docs/adr/011-phase5-reporting-foundation.md
test -f docs/adr/013-phase7-cutting-edge-foundation.md

# Site-local adversarial probes (no remote attack; contract falsification only)
./scripts/check-cutting-edge-claims.sh

# Invalid Cedar draft must not look like a complete permit statement
invalid="${ROOT}/fixtures/phase7/reg2policy/invalid.cedar"
test -f "$invalid" || { echo "adversarial: FAIL: missing invalid fixture" >&2; exit 1; }
if grep -q ');' "$invalid"; then
  echo "adversarial: FAIL: invalid fixture unexpectedly well-formed" >&2
  exit 1
fi
# Proposal path must never land under policies/
if [[ -f "${ROOT}/policies/cedar/proposed-valid.cedar" ]]; then
  echo "adversarial: FAIL: auto-deploy artifact under policies/cedar" >&2
  exit 1
fi

echo "adversarial: hermetic OK (claim lint + invalid reg2policy shape + no auto-deploy)"

# Tier-3: cedar-service anonymous evaluate deny (requires warm stack)
CEDAR="${CEDAR_SERVICE_URL:-http://localhost:8091}"
log() { echo ""; echo "== adversarial: $* =="; }

if curl -fsS "$CEDAR/api/v1/health" >/dev/null 2>&1; then
  log "anonymous_cedar_evaluate (expect 401)"
  code=$(curl -s -o /tmp/dt-adversarial.json -w "%{http_code}" \
    -X POST "$CEDAR/api/v1/evaluate" \
    -H "Content-Type: application/json" \
    -d '{"ruleCode":"BASEL-R001","input":{"lcr":0.9}}')
  if [[ "$code" != "401" ]]; then
    echo "adversarial: FAIL: expected HTTP 401 from anonymous evaluate, got ${code}" >&2
    exit 1
  fi
  echo "  ${code} (as expected)"
  echo ""
  echo "adversarial: ok (hermetic + cedar 401)"
else
  echo "adversarial: skip cedar 401 (cedar-service not running at $CEDAR)"
  echo "adversarial: ok (hermetic only; corporate adversary is separate)"
fi

exit 0
