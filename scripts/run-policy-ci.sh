#!/usr/bin/env bash
# Cedar Analyzer + Zen fixture regression gate.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

if command -v cedar >/dev/null 2>&1; then
  cedar analyze policies/cedar/
else
  echo "cedar CLI not installed; skipping analyze (install for local policy gate)"
fi

cd services/decision-service
go test ./internal/engine -run TestZenFixtures -count=1
