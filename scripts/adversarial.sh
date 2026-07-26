#!/usr/bin/env bash
# Site adversarial wrapper placeholder for G-CECT-HARNESS / AC-HARNESS-001.
# Full corporate adversary remains corporate-side after conformance; this stub
# proves the site argv is wired and exits 0 safely without attacking systems.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
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

echo "adversarial: OK (claim lint + invalid reg2policy shape + no auto-deploy; corporate adversary is separate)"
exit 0
