#!/usr/bin/env bash
#
# Coverage floor gates (no smoke re-run).
#
# Usage: ./scripts/check-coverage-gates.sh

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

check_service_coverage() {
  local service="$1"
  local min_pct="$2"
  local cover_out pct
  cover_out="$(mktemp)"
  if ! (cd "$ROOT/services/$service" && go test ./... -coverprofile="$cover_out") >/dev/null; then
    rm -f "$cover_out"
    return 1
  fi
  pct="$(cd "$ROOT/services/$service" && go tool cover -func="$cover_out" | awk '/^total:/ {gsub(/%/,"",$3); print $3}')"
  rm -f "$cover_out"
  if [[ -z "$pct" ]]; then
    return 1
  fi
  python3 - "$service" "$pct" "$min_pct" <<'PY'
import sys
svc, pct, min_pct = sys.argv[1], float(sys.argv[2]), float(sys.argv[3])
if pct < min_pct:
    print(f"{svc} coverage {pct:.1f}% < {min_pct:.1f}%", file=sys.stderr)
    sys.exit(1)
print(f"{svc} coverage {pct:.1f}%")
PY
}

echo "== Coverage gates =="
check_service_coverage state-service 35
check_service_coverage alert-service 25

check_cep_coverage() {
  local min_pct="$1"
  local csv pct
  if command -v mvn >/dev/null 2>&1; then
    if ! (cd "$ROOT/jobs/compliance-cep" && mvn -q test jacoco:report) >/dev/null; then
      return 1
    fi
  else
    if ! docker run --rm -v "$ROOT:/repo" -w /repo/jobs/compliance-cep maven:3.9-eclipse-temurin-17 mvn -q test jacoco:report >/dev/null; then
      return 1
    fi
  fi
  csv="$ROOT/jobs/compliance-cep/target/site/jacoco/jacoco.csv"
  if [[ ! -f "$csv" ]]; then
    echo "compliance-cep jacoco report missing" >&2
    return 1
  fi
  pct="$(awk -F, 'NR>1 {miss+=$4; cov+=$5} END {if (miss+cov==0) print 0; else printf "%.1f", (cov/(miss+cov))*100}' "$csv")"
  python3 - "compliance-cep" "$pct" "$min_pct" <<'PY'
import sys
svc, pct, min_pct = sys.argv[1], float(sys.argv[2]), float(sys.argv[3])
if pct < min_pct:
    print(f"{svc} coverage {pct:.1f}% < {min_pct:.1f}%", file=sys.stderr)
    sys.exit(1)
print(f"{svc} coverage {pct:.1f}%")
PY
}

check_cep_coverage 15
