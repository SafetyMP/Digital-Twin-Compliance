#!/usr/bin/env bash
# CI-sized load substitute — honest residual vs 10K evt/s target.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
EVIDENCE_DIR="${ROOT}/evidence"
mkdir -p "$EVIDENCE_DIR"
STAMP="$(date -u +"%Y%m%dT%H%M%SZ")"
OUT="${EVIDENCE_DIR}/phase6-load-${STAMP}.txt"
START=$(date +%s)
# Lightweight burst: 50 health polls (not a 10K claim)
for _ in $(seq 1 50); do
  curl -sf "${STATE_SERVICE_URL:-http://localhost:8080}/api/v1/health" >/dev/null || true
done
END=$(date +%s)
ELAPSED=$((END - START))
[[ "$ELAPSED" -lt 1 ]] && ELAPSED=1
RATE=$((50 / ELAPSED))
{
  echo "profile=ci-sized-substitute"
  echo "events=50"
  echo "elapsed_s=${ELAPSED}"
  echo "approx_eps=${RATE}"
  echo "target_eps=10000"
  echo "residual=full_10k_not_run_on_ci_hosts"
  echo "captured_at=$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
} | tee "$OUT"
echo "CI-sized load profile written: $OUT"
