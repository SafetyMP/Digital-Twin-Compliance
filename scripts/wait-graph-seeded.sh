#!/usr/bin/env bash
# Poll graph-service until seed graph meets Phase 4 thresholds.
set -euo pipefail

GRAPH_URL="${GRAPH_SERVICE_URL:-http://localhost:8093}"
MIN_NODES="${GRAPH_SEED_MIN_NODES:-10}"
MIN_EDGES="${GRAPH_SEED_MIN_EDGES:-50}"
TIMEOUT_SEC="${GRAPH_SEED_WAIT_SEC:-120}"

echo "Waiting for graph seed (nodes>=${MIN_NODES}, edges>=${MIN_EDGES}) at ${GRAPH_URL}..."

for _ in $(seq 1 "$TIMEOUT_SEC"); do
  if summary=$(curl -sf "${GRAPH_URL}/api/v1/graph/summary" 2>/dev/null); then
    nodes=$(echo "$summary" | jq -r '.nodeCount // 0')
    edges=$(echo "$summary" | jq -r '.edgeCount // 0')
    if [[ "$nodes" -ge "$MIN_NODES" && "$edges" -ge "$MIN_EDGES" ]]; then
      echo "Graph seeded: nodes=${nodes} edges=${edges}"
      exit 0
    fi
  fi
  sleep 1
done

echo "Graph seed timeout after ${TIMEOUT_SEC}s" >&2
curl -sf "${GRAPH_URL}/api/v1/graph/summary" | jq . 2>/dev/null || true
curl -sf "${GRAPH_URL}/api/v1/health" | jq . 2>/dev/null || true
exit 1
