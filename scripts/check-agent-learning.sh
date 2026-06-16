#!/usr/bin/env bash
# Agent learning hygiene checks — verifies externalized-memory artifacts exist and
# retention fixtures score correctly. Run after adding gotchas or behavioral scenarios.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

fail=0
ok() { echo "OK  $*"; }
bad() { echo "FAIL $*"; fail=1; }

echo "== Agent learning hygiene =="

# Required docs
for f in AGENTS.md docs/handoff-continuation.md; do
  if [[ -f "$f" ]]; then ok "$f present"; else bad "missing $f"; fi
done

# AGENTS.md learning section
if grep -q "## Agent learning" AGENTS.md; then
  ok "AGENTS.md has Agent learning section"
else
  bad "AGENTS.md missing ## Agent learning"
fi

if grep -q "### Capture checklist" AGENTS.md; then
  ok "AGENTS.md has capture checklist"
else
  bad "AGENTS.md missing capture checklist"
fi

# Retention scenario wired
SCENARIO="evals/live-model-phase2/scenarios/06-debug-int-m001-retention.md"
if [[ -f "$SCENARIO" ]]; then
  ok "retention scenario file present"
else
  bad "missing $SCENARIO"
fi

if python3 -c "
import json, sys
m = json.load(open('evals/live-model-phase2/manifest.json'))
ids = [s['id'] for s in m['scenarios']]
sys.exit(0 if 'debug-int-m001-retention' in ids else 1)
"; then
  ok "manifest lists debug-int-m001-retention"
else
  bad "manifest missing debug-int-m001-retention"
fi

if python3 -c "
import json, sys
g = json.load(open('evals/harness/gates.json'))['gates']
sys.exit(0 if g.get('debug-int-m001-retention', {}).get('type') == 'contract_retention' else 1)
"; then
  ok "gates.json has contract_retention gate"
else
  bad "gates.json missing contract_retention gate"
fi

# Gotcha inventory (informational)
gotcha_count="$(grep -c '^\-\s\+\*\*' AGENTS.md || true)"
echo "INFO repo gotcha bullets: ${gotcha_count}"

# Fixture regression for retention scenario only
PASS_FIX="evals/fixtures/transcripts/scenario-debug-int-m001-retention-pass.jsonl"
FAIL_FIX="evals/fixtures/transcripts/scenario-debug-int-m001-retention-fail.jsonl"
for f in "$PASS_FIX" "$FAIL_FIX"; do
  if [[ -f "$f" ]]; then ok "fixture $(basename "$f")"; else bad "missing $f"; fi
done

if [[ -f "$PASS_FIX" && -f "$FAIL_FIX" ]]; then
  if ./scripts/score-agent-transcript.py \
    --manifest evals/live-model-phase2/manifest.json \
    --scenario debug-int-m001-retention \
    --transcript "$PASS_FIX" >/dev/null 2>&1; then
    ok "retention pass fixture scores pass"
  else
    bad "retention pass fixture expected pass"
  fi

  if ./scripts/score-agent-transcript.py \
    --manifest evals/live-model-phase2/manifest.json \
    --scenario debug-int-m001-retention \
    --transcript "$FAIL_FIX" >/dev/null 2>&1; then
    bad "retention fail fixture expected fail"
  else
    ok "retention fail fixture scores fail"
  fi
fi

echo
if [[ "$fail" -eq 0 ]]; then
  echo "Agent learning hygiene: PASS"
  exit 0
fi
echo "Agent learning hygiene: FAIL"
exit 1
