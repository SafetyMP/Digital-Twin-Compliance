#!/usr/bin/env bash
#
# Score a live eval session using git diff + transcript.
#
# Usage:
#   ./scripts/score-eval-session.sh \\
#     --scenario bypass-outbox \\
#     --transcript ~/.cursor/projects/.../chat.jsonl \\
#     [--baseline-ref HEAD] \\
#     [--manifest evals/live-model/manifest.json] \\
#     [--write-result evals/live-model/results/bypass-outbox.json]

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

SCENARIO=""
TRANSCRIPT=""
BASELINE_REF="HEAD"
MANIFEST="$ROOT/evals/live-model/manifest.json"
WRITE_RESULT=""
EXTRA=()

while [[ $# -gt 0 ]]; do
  case "$1" in
    --scenario) SCENARIO="$2"; shift 2 ;;
    --transcript) TRANSCRIPT="$2"; shift 2 ;;
    --baseline-ref) BASELINE_REF="$2"; shift 2 ;;
    --manifest) MANIFEST="$2"; shift 2 ;;
    --write-result) WRITE_RESULT="$2"; shift 2 ;;
    --advisory-judge) EXTRA+=(--advisory-judge); shift ;;
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

if [[ -z "$SCENARIO" || -z "$TRANSCRIPT" ]]; then
  echo "Usage: $0 --scenario <id> --transcript <path.jsonl> [--baseline-ref HEAD]" >&2
  exit 2
fi

DIFF_FILE="$(mktemp)"
trap 'rm -f "$DIFF_FILE"' EXIT

git diff "$BASELINE_REF" >"$DIFF_FILE" || true

CMD=(
  ./scripts/score-agent-transcript.py
  --manifest "$MANIFEST"
  --scenario "$SCENARIO"
  --transcript "$TRANSCRIPT"
  --diff "$DIFF_FILE"
  --repo-root "$ROOT"
  --fail-on-harness-rereads
  --fail-on-efficiency
  "${EXTRA[@]}"
)

if [[ -n "$WRITE_RESULT" ]]; then
  CMD+=(--write-result "$WRITE_RESULT")
fi

echo "Baseline ref: $BASELINE_REF"
echo "Diff bytes: $(wc -c <"$DIFF_FILE" | tr -d ' ')"
"${CMD[@]}"
