#!/usr/bin/env bash
# Tier-3 adversarial oracle — cedar-service auth denies only.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

CEDAR="${CEDAR_SERVICE_URL:-http://localhost:8091}"

log() { echo ""; echo "== adversarial: $* =="; }

if ! curl -fsS "$CEDAR/api/v1/health" >/dev/null 2>&1; then
  echo "cedar-service not running at $CEDAR — start compose stack first" >&2
  exit 1
fi

# deny_case: anonymous_cedar_evaluate
log "anonymous_cedar_evaluate (expect 401)"
code=$(curl -s -o /tmp/dt-adversarial.json -w "%{http_code}" \
  -X POST "$CEDAR/api/v1/evaluate" \
  -H "Content-Type: application/json" \
  -d '{"ruleCode":"BASEL-R001","input":{"lcr":0.9}}')
[[ "$code" == "401" ]]
echo "  ${code} (as expected)"

echo ""
echo "adversarial: ok"
