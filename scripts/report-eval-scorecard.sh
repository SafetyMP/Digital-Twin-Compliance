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


def load_manifest(results_dir: Path) -> dict:
    if "phase2" in results_dir.parts:
        path = root / "evals/live-model-phase2/manifest.json"
    else:
        path = root / "evals/live-model/manifest.json"
    return json.loads(path.read_text(encoding="utf-8"))


def collect_runs(results_dir: Path) -> dict[str, list[dict]]:
    by_scenario: dict[str, list[dict]] = {}
    if not results_dir.is_dir():
        return by_scenario
    for path in sorted(results_dir.rglob("*.json")):
        if path.name == "efficiency-baseline.json":
            continue
        try:
            data = json.loads(path.read_text(encoding="utf-8"))
        except json.JSONDecodeError:
            continue
        scenario = data.get("scenario")
        if not scenario:
            continue
        by_scenario.setdefault(str(scenario), []).append(data)
    return by_scenario


def summarize(results_dir: Path, label: str) -> None:
    manifest = load_manifest(results_dir)
    threshold = manifest.get("pass_threshold", {}).get("live_scenarios", {})
    if isinstance(threshold, (int, float)):
        min_rate = float(threshold)
        min_runs = 1
    else:
        min_rate = float(threshold.get("min_pass_rate", 0.8))
        min_runs = int(threshold.get("runs_per_scenario", 3))

    by_scenario = collect_runs(results_dir)
    if not by_scenario:
        print(f"  {label}: no stored results under {results_dir}")
        return

    scenarios_ok = 0
    scenarios_total = len(manifest.get("scenarios") or [])
    for scenario in manifest.get("scenarios") or []:
        sid = scenario["id"]
        runs = by_scenario.get(sid, [])
        if not runs:
            print(f"  {label} {sid}: no runs")
            continue
        passed = sum(1 for r in runs if r.get("passed"))
        rate = passed / len(runs)
        ok = len(runs) >= min_runs and rate >= min_rate
        flag = "ok" if ok else "FAIL"
        print(
            f"  {label} {sid}: pass_rate={rate:.0%} ({passed}/{len(runs)} runs, need >={min_rate:.0%} over >={min_runs}) {flag}"
        )
        if ok:
            scenarios_ok += 1

    need = max(1, int(scenarios_total * min_rate))
    print(f"  {label} summary: {scenarios_ok}/{scenarios_total} scenarios meet pass-rate bar (need >={need})")
    if scenarios_ok < need:
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
  ./scripts/run-eval-fixtures.sh || efficiency_ok=0
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
