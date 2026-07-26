#!/usr/bin/env bash
# Lightweight UI health + Next.js /api proxy checks (stack must be running).
set -euo pipefail

ALERT_UI="${ALERT_CONSOLE_URL:-http://localhost:3000}"
AUDIT_UI="${AUDIT_EXPLORER_URL:-http://localhost:3002}"
GRAPH_UI="${GRAPH_EXPLORER_URL:-http://localhost:3003}"
SIM_UI="${SIMULATION_CONSOLE_URL:-http://localhost:3004}"

check_root() {
  local name="$1"
  local url="$2"
  curl -sf "$url/" >/dev/null || {
    echo "UI root unreachable: $name ($url/)" >&2
    return 1
  }
  echo "OK $name /"
}

check_proxy() {
  local name="$1"
  local url="$2"
  curl -sf "$url" >/dev/null || {
    echo "UI proxy unreachable: $name ($url)" >&2
    return 1
  }
  echo "OK $name proxy"
}

echo "== UI proxy smoke =="
check_root "alert-console" "$ALERT_UI"
check_root "audit-explorer" "$AUDIT_UI"
check_root "graph-explorer" "$GRAPH_UI"
check_root "simulation-console" "$SIM_UI"

check_proxy "alert-console" "${ALERT_UI}/api/alerts?limit=1"
check_proxy "audit-explorer" "${AUDIT_UI}/api/audit/verify"
check_proxy "graph-explorer" "${GRAPH_UI}/api/graph/summary"

echo "UI proxy smoke passed"
