#!/usr/bin/env bash
#
# Refresh efficiency baseline from Cursor agent transcripts.
#
# Usage:
#   ./scripts/refresh-efficiency-baseline.sh
#   TRANSCRIPT_DIR=/path/to/transcripts ./scripts/refresh-efficiency-baseline.sh

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

TRANSCRIPT_DIR="${TRANSCRIPT_DIR:-$HOME/.cursor/projects/Users-sagehart-Downloads-Digital-Twin/agent-transcripts}"
OUT="${1:-$ROOT/evals/live-model/results/efficiency-baseline.json}"

if [[ ! -d "$TRANSCRIPT_DIR" ]]; then
  echo "Transcript dir not found: $TRANSCRIPT_DIR" >&2
  exit 2
fi

./scripts/score-agent-transcript.py \
  --write-baseline "$OUT" \
  --transcript-dir "$TRANSCRIPT_DIR"

echo "Baseline written: $OUT"
echo "Run after several clean verification sessions (harness_reread_count: 0)."
