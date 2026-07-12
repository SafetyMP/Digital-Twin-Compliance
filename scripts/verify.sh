#!/usr/bin/env bash
# Definition of Done — unit/policy gates (no Docker). Full stack: CURSOR_VERIFY_STACK=1.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

echo "== verify: script hygiene =="
chmod +x scripts/run-policy-ci.sh scripts/publish-check.sh scripts/check-agent-worktrees.sh \
  scripts/check-kafka-contracts.sh scripts/smoke-test-phase4.sh scripts/wait-graph-seeded.sh 2>/dev/null || true
bash -n scripts/smoke-test-phase4.sh scripts/wait-graph-seeded.sh

echo "== verify: go unit tests =="
for svc in state-service alert-service audit-service cedar-service decision-service graph-service; do
  echo "-- services/$svc"
  (cd "services/$svc" && go test ./...)
done

echo "== verify: simulation-service =="
python3 -m pip install -q -r services/simulation-service/requirements.txt
(cd services/simulation-service && python3 -m pytest -q)

echo "== verify: policy + contracts =="
./scripts/run-policy-ci.sh
./scripts/check-kafka-contracts.sh
./scripts/check-agent-worktrees.sh
./scripts/publish-check.sh

if [[ "${CURSOR_VERIFY_STACK:-}" == "1" ]]; then
  echo "== verify: full stack smokes (Docker required) =="
  docker compose -f docker-compose.dev.yml up -d --wait
  ./scripts/seed.sh
  docker compose -f docker-compose.dev.yml restart state-service
  docker compose -f docker-compose.dev.yml up -d --wait state-service
  SMOKE_PHASE4_SKIP_PREREQS=1 ./scripts/smoke-test-phase4.sh
fi

echo "verify: ok"
