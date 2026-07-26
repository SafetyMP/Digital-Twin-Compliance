#!/usr/bin/env bash
# Site oracle (hermetic early gate). Expanded by later CECT packets.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

fail() { echo "verify: FAIL: $*" >&2; exit 1; }

test -f docs/adr/011-phase5-reporting-foundation.md || fail "missing ADR-011"
test -x scripts/verify.sh || fail "scripts/verify.sh not executable"
test -x scripts/adversarial.sh || fail "scripts/adversarial.sh not executable"
test -x scripts/harness/verify.sh || fail "scripts/harness/verify.sh not executable"
test -x scripts/harness/adversarial.sh || fail "scripts/harness/adversarial.sh not executable"
test -f .corp-harness/site.json || fail "missing .corp-harness/site.json"

site_id="$(python3 -c 'import json; print(json.load(open(".corp-harness/site.json"))["site_id"])')"
[[ "$site_id" == "digital-twin" ]] || fail "site_id=$site_id (expected digital-twin)"

bash -n scripts/verify.sh scripts/adversarial.sh scripts/harness/verify.sh scripts/harness/adversarial.sh
test -f docs/adr/012-phase6-hardening-foundation.md || fail "missing ADR-012"
test -f docs/adr/013-phase7-cutting-edge-foundation.md || fail "missing ADR-013"
test -x scripts/smoke-test-phase5.sh || fail "missing smoke-test-phase5.sh"
test -d services/reporting-service/reporting_service || fail "missing reporting-service"
test -x scripts/smoke-test-phase6-oidc.sh || fail "missing phase6 oidc smoke"
test -x scripts/smoke-test-phase6-harden.sh || fail "missing phase6 harden smoke"
test -f docker-compose.hardening.yml || fail "missing hardening compose"
test -f docs/explainability.md || fail "missing explainability"
test -f docs/deployment.md || fail "missing deployment.md"

# Phase 7 cutting-edge (ADR-013) hermetic gates
test -x scripts/smoke-test-phase7-effectiveness.sh || fail "missing phase7 effectiveness smoke"
test -x scripts/smoke-test-phase7-contagion.sh || fail "missing phase7 contagion smoke"
test -x scripts/smoke-test-phase7-reg2policy.sh || fail "missing phase7 reg2policy smoke"
test -x scripts/smoke-test-phase7-graph.sh || fail "missing phase7 graph smoke"
test -x scripts/check-cutting-edge-claims.sh || fail "missing claim lint"
test -x scripts/phase7/reg2policy.sh || fail "missing reg2policy script"
test -f fixtures/phase7/breaches.json || fail "missing phase7 breach fixtures"
test -f services/simulation-service/simulation_service/effectiveness.py || fail "missing effectiveness module"
test -f services/simulation-service/simulation_service/contagion.py || fail "missing contagion module"
grep -q 'ShortestPath' services/graph-service/internal/graph/store.go || fail "missing graph ShortestPath"
grep -q '/api/v1/graph/paths' services/graph-service/internal/api/handlers.go || fail "missing graph paths route"
bash -n scripts/smoke-test-phase7-effectiveness.sh scripts/smoke-test-phase7-contagion.sh \
  scripts/smoke-test-phase7-reg2policy.sh scripts/smoke-test-phase7-graph.sh \
  scripts/check-cutting-edge-claims.sh scripts/phase7/reg2policy.sh
./scripts/check-cutting-edge-claims.sh
./scripts/smoke-test-phase7-reg2policy.sh

echo "verify: OK (hermetic harness + ADR-011/012/013 + Phase 5–7 scaffolds + claim lint + reg2policy)"
exit 0
