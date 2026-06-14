#!/usr/bin/env bash
#
# Mechanical live-eval checks for Digital Twin Phase 1.
# Scores repo state and (optionally) Definition of Done commands.
#
# Usage:
#   ./scripts/run-live-evals.sh           # fast, no Docker
#   ./scripts/run-live-evals.sh --full    # also runs go test, coverage gate, smoke-test

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

FULL=0
if [[ "${1:-}" == "--full" ]]; then
  FULL=1
fi

PASS=0
FAIL=0

check() {
  local name="$1"
  shift
  if "$@"; then
    echo "  ok  - $name"
    PASS=$((PASS + 1))
  else
    echo "  FAIL - $name"
    FAIL=$((FAIL + 1))
  fi
}

check_scope_boundary() {
  local hits
  hits=$(rg -l -i \
    'apache/flink|org\.apache\.flink|immudb|neo4j|keycloak|cedar-policy|gorules|next\.js' \
    services mocks schemas \
    docker-compose.dev.yml \
    .github/workflows \
    --glob '*.{go,sql,yml,yaml,avsc,sh}' \
    --glob '!scripts/run-live-evals.sh' \
    --glob '!scripts/score-agent-transcript.py' \
    2>/dev/null || true)
  [[ -z "$hits" ]]
}

check_tenant_id_columns() {
  local migration="$ROOT/services/state-service/migrations/001_init.sql"
  [[ -f "$migration" ]] || return 1
  rg -q 'tenant_id' "$migration" && \
    rg -q '00000000-0000-0000-0000-000000000001' "$migration"
}

check_outbox_only_kafka_writer() {
  local writers
  writers=$(rg -l 'kafka\.Writer' services/state-service --glob '*.go' 2>/dev/null || true)
  if [[ -z "$writers" ]]; then
    return 0
  fi
  local extra
  extra=$(echo "$writers" | rg -v 'internal/outbox/' || true)
  [[ -z "$extra" ]]
}

check_required_scripts() {
  [[ -x "$ROOT/scripts/seed.sh" ]] && \
    [[ -x "$ROOT/scripts/smoke-test.sh" ]] && \
    [[ -x "$ROOT/scripts/register-schemas.sh" ]] && \
    [[ -f "$ROOT/docker-compose.dev.yml" ]]
}

check_avro_schemas() {
  [[ -f "$ROOT/schemas/avro/event-envelope.avsc" ]] && \
    [[ -f "$ROOT/schemas/avro/entity-state-updated.avsc" ]] && \
    [[ -f "$ROOT/schemas/avro/twin-state-updated.avsc" ]]
}

check_go_tests() {
  (cd "$ROOT/services/state-service" && go test ./...)
}

check_state_service_package_tests() {
  local pkg
  for pkg in api config consumer events outbox store; do
    if ! compgen -G "$ROOT/services/state-service/internal/$pkg/*_test.go" > /dev/null; then
      return 1
    fi
  done
  [[ -f "$ROOT/services/state-service/cmd/server/migrations_test.go" ]]
}

check_state_service_coverage() {
  local cover_out pct
  cover_out="$(mktemp)"
  if ! (cd "$ROOT/services/state-service" && go test ./... -coverprofile="$cover_out") >/dev/null; then
    rm -f "$cover_out"
    return 1
  fi
  pct="$(cd "$ROOT/services/state-service" && go tool cover -func="$cover_out" | awk '/^total:/ {gsub(/%/,"",$3); print $3}')"
  rm -f "$cover_out"
  if [[ -z "$pct" ]]; then
    return 1
  fi
  python3 - "$pct" <<'PY'
import sys
pct = float(sys.argv[1])
min_pct = 35.0
if pct < min_pct:
    print(f"state-service coverage {pct:.1f}% < {min_pct:.1f}%", file=sys.stderr)
    sys.exit(1)
print(f"state-service coverage {pct:.1f}%")
PY
}

check_smoke_test() {
  "$ROOT/scripts/smoke-test.sh"
}

echo "== Digital Twin Phase 1 live evals (mechanical) =="
echo "Repo: $(basename "$ROOT")"
echo

echo "== Scope & contracts =="
check "scope-boundary (no Phase 2+ stack in code)" check_scope_boundary
check "tenant-id-columns in state migrations" check_tenant_id_columns
check "outbox-only-kafka-writer" check_outbox_only_kafka_writer
check "required-scripts-present" check_required_scripts
check "avro-schemas-present" check_avro_schemas
check "state-service-package-tests" check_state_service_package_tests

if [[ "$FULL" -eq 1 ]]; then
  echo
  echo "== Definition of Done (requires stack up) =="
  check "go test ./... in state-service" check_go_tests
  check "state-service-coverage-floor (>=35%)" check_state_service_coverage
  check "scripts/smoke-test.sh exits 0" check_smoke_test
else
  echo
  echo "== Skipped (--full not set) =="
  echo "  ... go test + coverage + smoke-test (run: ./scripts/run-live-evals.sh --full)"
fi

echo
TOTAL=$((PASS + FAIL))
echo "RESULT: $PASS passed, $FAIL failed (mechanical)"

if [[ "$FAIL" -gt 0 ]]; then
  exit 1
fi
