#!/usr/bin/env bash
#
# Run behavioral eval scenarios (v1: manual transcript; v2: optional Cursor SDK).
#
# Usage:
#   ./scripts/run-behavioral-eval.sh --phase1 --scenario bypass-outbox
#   ./scripts/run-behavioral-eval.sh --phase2 --scenario add-cedar-policy --runs 3
#   ./scripts/run-behavioral-eval.sh --phase2 --scenario add-cedar-policy --transcript path.jsonl
#   ./scripts/run-behavioral-eval.sh --phase2 --dry-run

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

PHASE=""
SCENARIO=""
RUNS=1
TRANSCRIPT=""
DRY_RUN=0
BASELINE_REF="HEAD"
USE_SDK=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --phase1) PHASE=1; shift ;;
    --phase2) PHASE=2; shift ;;
    --scenario) SCENARIO="$2"; shift 2 ;;
    --runs) RUNS="$2"; shift 2 ;;
    --transcript) TRANSCRIPT="$2"; shift 2 ;;
    --baseline-ref) BASELINE_REF="$2"; shift 2 ;;
    --sdk) USE_SDK=1; shift ;;
    --dry-run) DRY_RUN=1; shift ;;
    -h|--help)
      sed -n '2,12p' "$0"
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      exit 2
      ;;
  esac
done

if [[ -z "$PHASE" || -z "$SCENARIO" ]]; then
  echo "Usage: $0 --phase1|--phase2 --scenario <id> [--runs N] [--transcript path]" >&2
  exit 2
fi

if [[ "$PHASE" == "1" ]]; then
  MANIFEST="$ROOT/evals/live-model/manifest.json"
  RESULTS="$ROOT/evals/live-model/results"
else
  MANIFEST="$ROOT/evals/live-model-phase2/manifest.json"
  RESULTS="$ROOT/evals/live-model-phase2/results"
fi

SCENARIO_DIR="$(python3 - "$MANIFEST" "$SCENARIO" <<'PY'
import json, sys
from pathlib import Path
manifest_path = Path(sys.argv[1])
manifest = json.loads(manifest_path.read_text(encoding="utf-8"))
scenario_id = sys.argv[2]
for s in manifest["scenarios"]:
    if s["id"] == scenario_id:
        print(manifest_path.parent / s["file"])
        break
else:
    raise SystemExit(f"unknown scenario {scenario_id!r}")
PY
)"

PROMPT="$(python3 - "$SCENARIO_DIR" <<'PY'
import re, sys
from pathlib import Path
text = Path(sys.argv[1]).read_text(encoding="utf-8")
m = re.search(r"```\n(.*?)```", text, re.S)
if not m:
    raise SystemExit("prompt block not found")
print(m.group(1).strip())
PY
)"

echo "== Behavioral eval =="
echo "Scenario: $SCENARIO (phase $PHASE, runs=$RUNS)"
echo
echo "Prompt:"
echo "$PROMPT"
echo

if [[ "$DRY_RUN" -eq 1 ]]; then
  echo "Dry run — not scoring."
  exit 0
fi

if [[ "$USE_SDK" -eq 1 && -n "${CURSOR_API_KEY:-}" ]]; then
  echo "SDK mode requested but not wired in this repo yet; use --transcript after a manual run." >&2
  exit 2
fi

for ((i=1; i<=RUNS; i++)); do
  RUN_TRANSCRIPT="$TRANSCRIPT"
  if [[ -z "$RUN_TRANSCRIPT" ]]; then
    echo "Run $i/$RUNS: paste agent in a fresh chat, export transcript path:"
    read -r RUN_TRANSCRIPT
  fi
  if [[ ! -f "$RUN_TRANSCRIPT" ]]; then
    echo "Transcript not found: $RUN_TRANSCRIPT" >&2
    exit 2
  fi
  OUT_DIR="$RESULTS/$SCENARIO"
  mkdir -p "$OUT_DIR"
  STAMP="$(date +%Y%m%dT%H%M%S)"
  OUT_FILE="$OUT_DIR/run-${STAMP}.json"
  ./scripts/score-eval-session.sh \
    --manifest "$MANIFEST" \
    --scenario "$SCENARIO" \
    --transcript "$RUN_TRANSCRIPT" \
    --baseline-ref "$BASELINE_REF" \
    --write-result "$OUT_FILE"
  echo "Wrote $OUT_FILE"
  if [[ -n "$TRANSCRIPT" ]]; then
    break
  fi
done

echo "Done."
