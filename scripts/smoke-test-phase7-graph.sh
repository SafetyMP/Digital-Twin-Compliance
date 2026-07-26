#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GRAPH="${GRAPH_SERVICE_URL:-http://localhost:8093}"
EVIDENCE_DIR="${ROOT}/evidence"
mkdir -p "$EVIDENCE_DIR"
STAMP="$(date -u +"%Y%m%dT%H%M%SZ")"
OUT="${EVIDENCE_DIR}/phase7-graph-${STAMP}.txt"
: >"$OUT"

"$ROOT/scripts/wait-graph-seeded.sh"
nodes="$(curl -sf "${GRAPH}/api/v1/graph/nodes?limit=20")"
echo "$nodes" | tee -a "$OUT" >/dev/null
from="$(echo "$nodes" | jq -r '.[0].entityId')"
to="$(echo "$nodes" | jq -r '.[1].entityId // .[0].entityId')"
[[ -n "$from" && "$from" != "null" ]]

path="$(curl -sf "${GRAPH}/api/v1/graph/paths?from=${from}&to=${to}")"
echo "$path" | tee -a "$OUT" >/dev/null
echo "$path" | jq -e '.hops >= 0 and (.nodeIds|length) >= 1' >/dev/null

cent="$(curl -sf "${GRAPH}/api/v1/graph/centrality?limit=10")"
echo "$cent" | tee -a "$OUT" >/dev/null
echo "$cent" | jq -e '.metric=="degree" and (.rows|length) >= 1 and .rows[0].degree >= 0' >/dev/null

echo "Phase 7 graph smoke passed"
