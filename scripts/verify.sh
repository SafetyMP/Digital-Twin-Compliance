#!/usr/bin/env bash
# Definition of Done — static checks without Docker (full CI needs compose stack).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

echo "==> publish-check"
chmod +x scripts/publish-check.sh
./scripts/publish-check.sh

echo "==> agent-worktree scripts"
chmod +x scripts/check-agent-worktrees.sh
./scripts/check-agent-worktrees.sh

if command -v go >/dev/null 2>&1; then
  echo "==> go vet (state-service)"
  (cd services/state-service && go vet ./...)
  echo "==> go vet (alert-service)"
  (cd services/alert-service && go vet ./...)
else
  echo "skip go vet (go not installed)"
fi

echo "verify: ok (static parity; full smoke requires Docker — see AGENTS.md)"
