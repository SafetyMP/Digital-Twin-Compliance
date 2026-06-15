#!/usr/bin/env bash
#
# Mechanical live-eval checks for Digital Twin Phase 2.
# Scores repo state and (optionally) Phase 2 Definition of Done commands.
#
# Usage:
#   ./scripts/run-live-evals-phase2.sh           # fast, no Docker
#   ./scripts/run-live-evals-phase2.sh --full    # also runs tests + smoke tests

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

invariant_check() {
  python3 -m evals.lib.invariants --check "$1" --repo-root "$ROOT"
}

check_phase3_scope_boundary() {
  invariant_check phase3-scope-boundary
}

check_phase2_stack_in_compose() {
  local compose="$ROOT/docker-compose.dev.yml"
  [[ -f "$compose" ]] || return 1
  rg -q '^\s+alert-service:' "$compose" && \
    rg -q '^\s+redis:' "$compose" && \
    rg -q '^\s+flink-jobmanager:' "$compose" && \
    rg -q '^\s+grafana:' "$compose" && \
    rg -q '^\s+alert-console:' "$compose"
}

check_phase2_deliverables() {
  [[ -d "$ROOT/services/alert-service" ]] && \
    [[ -f "$ROOT/services/alert-service/migrations/001_alerts.sql" ]] && \
    [[ -f "$ROOT/jobs/compliance-cep/pom.xml" ]] && \
    [[ -f "$ROOT/jobs/compliance-cep/src/main/java/com/digitaltwin/jobs/cep/ComplianceCepJob.java" ]] && \
    [[ -d "$ROOT/apps/alert-console" ]]
}

check_alert_avro_schemas() {
  [[ -f "$ROOT/schemas/avro/compliance-alert-raised.avsc" ]] && \
    [[ -f "$ROOT/schemas/avro/compliance-alert-resolved.avsc" ]]
}

check_tenant_id_alert_migrations() {
  invariant_check tenant-id-alert-migrations
}

check_tenant_id_state_migrations() {
  invariant_check tenant-id-state-migrations
}

check_outbox_only_kafka_writer() {
  invariant_check outbox-only-kafka-writer
}

check_required_phase2_scripts() {
  [[ -x "$ROOT/scripts/smoke-test-phase2.sh" ]] && \
    [[ -x "$ROOT/scripts/submit-flink-job.sh" ]] && \
    [[ -x "$ROOT/scripts/smoke-test.sh" ]] && \
    [[ -x "$ROOT/scripts/seed.sh" ]]
}

check_alert_idempotency_key() {
  local migration="$ROOT/services/alert-service/migrations/001_alerts.sql"
  [[ -f "$migration" ]] || return 1
  rg -q 'idempotency_key' "$migration"
}

check_go_tests_alert_service() {
  (cd "$ROOT/services/alert-service" && go test ./...)
}

check_alert_service_package_tests() {
  local pkg
  for pkg in api config consumer events hub store; do
    if ! compgen -G "$ROOT/services/alert-service/internal/$pkg/*_test.go" > /dev/null; then
      return 1
    fi
  done
  [[ -f "$ROOT/services/alert-service/cmd/server/migrations_test.go" ]]
}

check_alert_service_coverage() {
  local cover_out pct
  cover_out="$(mktemp)"
  if ! (cd "$ROOT/services/alert-service" && go test ./... -coverprofile="$cover_out") >/dev/null; then
    rm -f "$cover_out"
    return 1
  fi
  pct="$(cd "$ROOT/services/alert-service" && go tool cover -func="$cover_out" | awk '/^total:/ {gsub(/%/,"",$3); print $3}')"
  rm -f "$cover_out"
  if [[ -z "$pct" ]]; then
    return 1
  fi
  python3 - "$pct" <<'PY'
import sys
pct = float(sys.argv[1])
min_pct = 25.0
if pct < min_pct:
    print(f"alert-service coverage {pct:.1f}% < {min_pct:.1f}%", file=sys.stderr)
    sys.exit(1)
print(f"alert-service coverage {pct:.1f}%")
PY
}

check_maven_tests_compliance_cep() {
  (cd "$ROOT/jobs/compliance-cep" && mvn -q test)
}

check_smoke_test_phase1() {
  "$ROOT/scripts/smoke-test.sh"
}

check_smoke_test_phase2() {
  "$ROOT/scripts/smoke-test-phase2.sh"
}

echo "== Digital Twin Phase 2 live evals (mechanical) =="
echo "Repo: $(basename "$ROOT")"
echo

echo "== Phase 2 deliverables =="
check "phase2-stack-in-compose" check_phase2_stack_in_compose
check "phase2-deliverables-present" check_phase2_deliverables
check "alert-avro-schemas-present" check_alert_avro_schemas
check "required-phase2-scripts-present" check_required_phase2_scripts
check "alert-idempotency-key-column" check_alert_idempotency_key

echo
echo "== Contracts (Phase 1 + Phase 2) =="
check "tenant-id-columns in alert migrations" check_tenant_id_alert_migrations
check "tenant-id-columns in state migrations" check_tenant_id_state_migrations
check "outbox-only-kafka-writer (state-service)" check_outbox_only_kafka_writer
check "phase3-scope-boundary (no Cedar/immudb/Neo4j/auth)" check_phase3_scope_boundary
check "alert-service-package-tests" check_alert_service_package_tests

if [[ "$FULL" -eq 1 ]]; then
  echo
  echo "== Definition of Done (requires stack up) =="
  check "go test ./... in alert-service" check_go_tests_alert_service
  check "alert-service-coverage-floor (>=25%)" check_alert_service_coverage
  check "mvn test in compliance-cep" check_maven_tests_compliance_cep
  check "scripts/smoke-test.sh exits 0 (Phase 1 regression)" check_smoke_test_phase1
  check "scripts/smoke-test-phase2.sh exits 0" check_smoke_test_phase2
else
  echo
  echo "== Skipped (--full not set) =="
  echo "  ... go test + mvn test + smoke-test.sh + smoke-test-phase2.sh"
  echo "  ... (run: ./scripts/run-live-evals-phase2.sh --full)"
fi

echo
TOTAL=$((PASS + FAIL))
echo "RESULT: $PASS passed, $FAIL failed (mechanical)"

if [[ "$FAIL" -gt 0 ]]; then
  exit 1
fi
