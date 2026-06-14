#!/usr/bin/env bash
#
# Three-pillar eval scorecard: Product, Behavior, Efficiency.
#
# Usage:
#   ./scripts/report-eval-scorecard.sh --phase1
#   ./scripts/report-eval-scorecard.sh --phase2
#   ./scripts/report-eval-scorecard.sh --all [--full] [--transcript path.jsonl]

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

PHASE1=0
PHASE2=0
FULL=0
TRANSCRIPT=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --phase1) PHASE1=1; shift ;;
    --phase2) PHASE2=1; shift ;;
    --all) PHASE1=1; PHASE2=1; shift ;;
    --full) FULL=1; shift ;;
    --transcript) TRANSCRIPT="$2"; shift 2 ;;
    -h|--help)
      echo "Usage: $0 [--phase1|--phase2|--all] [--full] [--transcript path.jsonl]"
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      exit 2
      ;;
  esac
done

if [[ "$PHASE1" -eq 0 && "$PHASE2" -eq 0 ]]; then
  PHASE1=1
fi

product_ok=1
behavior_ok=1
efficiency_ok=1

echo "== Digital Twin eval scorecard =="
echo

echo "== Product (mechanical + DoD) =="
if [[ "$PHASE1" -eq 1 ]]; then
  echo "--- Phase 1 ---"
  if [[ "$FULL" -eq 1 ]]; then
    ./scripts/run-live-evals.sh --full || product_ok=0
  else
    ./scripts/run-live-evals.sh || product_ok=0
  fi
fi
if [[ "$PHASE2" -eq 1 ]]; then
  echo "--- Phase 2 ---"
  if [[ "$FULL" -eq 1 ]]; then
    ./scripts/run-live-evals-phase2.sh --full || product_ok=0
  else
    ./scripts/run-live-evals-phase2.sh || product_ok=0
  fi
fi

echo
echo "== Behavior (stored live scenario results) =="
python3 - "$ROOT" "$PHASE1" "$PHASE2" <<'PY' || behavior_ok=0
import json
import sys
from pathlib import Path

root = Path(sys.argv[1])
phase1 = int(sys.argv[2])
phase2 = int(sys.argv[3])

def summarize(results_dir: Path, label: str) -> None:
    files = sorted(results_dir.glob("*.json"))
    files = [f for f in files if f.name != "efficiency-baseline.json"]
    if not files:
        print(f"  {label}: no stored results under {results_dir}")
        return
    passed = 0
    total = 0
    weighted = 0.0
    weight_sum = 0
    for path in files:
        data = json.loads(path.read_text())
        if "scenario" not in data:
            continue
        total += 1
        if data.get("passed"):
            passed += 1
        score = float(data.get("score", 0))
        weighted += score
        weight_sum += 1
    avg = (weighted / weight_sum) if weight_sum else 0.0
    print(f"  {label}: {passed}/{total} scenarios passed, avg score={avg:.2f}")
    if total and passed < total:
        sys.exit(1)

if phase1:
    summarize(root / "evals/live-model/results", "Phase 1")
if phase2:
    summarize(root / "evals/live-model-phase2/results", "Phase 2")
PY

echo
echo "== Efficiency =="
if [[ -n "$TRANSCRIPT" ]]; then
  echo "Transcript: $TRANSCRIPT"
  ./scripts/token-efficiency.sh --strict "$TRANSCRIPT" || efficiency_ok=0
else
  echo "Fixture regression (no live transcript supplied):"
  ./scripts/run-efficiency-fixtures.sh || efficiency_ok=0
fi

echo
echo "== Summary =="
printf "  Product:     %s\n" "$([[ "$product_ok" -eq 1 ]] && echo PASS || echo FAIL)"
printf "  Behavior:    %s\n" "$([[ "$behavior_ok" -eq 1 ]] && echo PASS || echo FAIL)"
printf "  Efficiency:  %s\n" "$([[ "$efficiency_ok" -eq 1 ]] && echo PASS || echo FAIL)"

if [[ "$product_ok" -eq 1 && "$behavior_ok" -eq 1 && "$efficiency_ok" -eq 1 ]]; then
  exit 0
fi
exit 1
