#!/usr/bin/env bash
# Sample Cedar/Zen evaluate latency and optional Flink checkpoint stats (warm stack).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

SAMPLES="${PHASE3_BENCH_SAMPLES:-50}"
CEDAR_URL="${CEDAR_SERVICE_URL:-http://localhost:8091}"
DECISION_URL="${DECISION_SERVICE_URL:-http://localhost:8092}"
FLINK_URL="${FLINK_JOBMANAGER_URL:-http://localhost:8082}"
RULE_P99_BUDGET_MS="${PHASE3_RULE_EVAL_P99_MS:-5}"
ALERT_P99_BUDGET_MS="${PHASE3_ALERT_P99_MS:-2000}"

fail=0
warn=0

ok() { echo "OK  $*"; }
warn_line() { echo "WARN $*"; warn=1; }
bad() { echo "FAIL $*"; fail=1; }

require_health() {
  local url="$1" name="$2"
  if curl -sf "${url}/api/v1/health" >/dev/null 2>&1; then
    ok "${name} healthy at ${url}"
    return 0
  fi
  bad "${name} unreachable at ${url} — start stack before benchmarking"
  return 1
}

bench_post_ms() {
  local url="$1" payload="$2" samples="$3"
  python3 - "$url" "$samples" "$payload" <<'PY'
import json, sys, time, urllib.request

url, n, payload = sys.argv[1], int(sys.argv[2]), sys.argv[3]
data = payload.encode()
req = urllib.request.Request(
    url, data=data, method="POST", headers={"Content-Type": "application/json"}
)
times = []
for _ in range(n):
    t0 = time.perf_counter()
    with urllib.request.urlopen(req, timeout=10) as resp:
        resp.read()
    times.append((time.perf_counter() - t0) * 1000.0)
times.sort()
def pct(p):
    if not times:
        return 0.0
    idx = min(len(times) - 1, int(round((p / 100.0) * (len(times) - 1))))
    return times[idx]
print(json.dumps({"n": n, "p50_ms": round(pct(50), 3), "p99_ms": round(pct(99), 3), "max_ms": round(times[-1], 3)}))
PY
}

echo "== Phase 3 latency sample (n=${SAMPLES}) =="

if ! require_health "$CEDAR_URL" "cedar-service"; then
  echo "Phase 3 latency: SKIP (stack down)"
  exit 1
fi
require_health "$DECISION_URL" "decision-service" || exit 1

cedar_payload='{"ruleCode":"INT-R003","principal":{"id":"bench-user","roles":["Reporter"]},"resource":{"type":"TwinData","id":"t1","attributes":{"sensitivity":"high"}}}'
zen_payload='{"ruleCode":"BASEL-R001","input":{"lcr":0.9,"personaId":"44444444-4444-4444-4444-444444444401"}}'

cedar_stats="$(bench_post_ms "${CEDAR_URL}/api/v1/evaluate" "$cedar_payload" "$SAMPLES")"
zen_stats="$(bench_post_ms "${DECISION_URL}/api/v1/evaluate" "$zen_payload" "$SAMPLES")"

echo "Cedar INT-R003: $cedar_stats"
echo "Zen BASEL-R001: $zen_stats"

cedar_p99="$(echo "$cedar_stats" | python3 -c "import json,sys; print(json.load(sys.stdin)['p99_ms'])")"
zen_p99="$(echo "$zen_stats" | python3 -c "import json,sys; print(json.load(sys.stdin)['p99_ms'])")"

rule_p99="$(python3 -c "print(max(float('$cedar_p99'), float('$zen_p99')))")"
if python3 -c "import sys; sys.exit(0 if float('$rule_p99') <= float('$RULE_P99_BUDGET_MS') else 1)"; then
  ok "rule evaluate p99 ${rule_p99}ms <= ${RULE_P99_BUDGET_MS}ms budget"
else
  warn_line "rule evaluate p99 ${rule_p99}ms exceeds ${RULE_P99_BUDGET_MS}ms roadmap target (dev Compose; not a soak gate)"
fi

if curl -sf "${FLINK_URL}/overview" >/dev/null 2>&1; then
  running="$(curl -sf "${FLINK_URL}/jobs/overview" | python3 -c "import json,sys; d=json.load(sys.stdin); print(sum(1 for j in d.get('jobs',[]) if j.get('state')=='RUNNING'))" 2>/dev/null || echo 0)"
  ok "Flink JobManager reachable; RUNNING jobs=${running}"
  if [[ "$running" -eq 0 ]]; then
    warn_line "no RUNNING Flink job — submit with ./scripts/submit-flink-job.sh before checkpoint stats matter"
  fi
else
  warn_line "Flink JobManager not reachable at ${FLINK_URL} — skip checkpoint stats"
fi

echo
if [[ "$fail" -eq 0 ]]; then
  echo "Phase 3 latency sample: PASS${warn:+ with warnings}"
  exit 0
fi
echo "Phase 3 latency sample: FAIL"
exit 1
