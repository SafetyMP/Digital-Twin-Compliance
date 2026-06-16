#!/usr/bin/env bash
# Cedar CLI + Zen fixture regression gate.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

if command -v cedar >/dev/null 2>&1; then
  shopt -s nullglob
  cedar_policies=(policies/cedar/*.cedar)
  if [[ ${#cedar_policies[@]} -eq 0 ]]; then
    echo "no Cedar policies found under policies/cedar/" >&2
    exit 1
  fi
  for policy in "${cedar_policies[@]}"; do
    echo "cedar check-parse -p ${policy}"
    cedar check-parse -p "${policy}"
  done
else
  if [[ "${CI:-}" == "true" || "${GITHUB_ACTIONS:-}" == "true" ]]; then
    echo "cedar CLI not installed (run scripts/install-cedar-cli.sh)" >&2
    exit 1
  fi
  echo "cedar CLI not installed; skipping policy parse gate (install for local policy gate)"
fi

cd services/decision-service
go test ./internal/engine -run TestZenFixtures -count=1
