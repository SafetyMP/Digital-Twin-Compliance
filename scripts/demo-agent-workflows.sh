#!/usr/bin/env bash
# Dry-run dependency-wave workflow for hygiene checks and onboarding (no worktrees created).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

TASK="${AWT_DEMO_TASK:-_hygiene-demo}"
CLEANUP=1

usage() {
  cat <<'EOF'
Usage: ./scripts/demo-agent-workflows.sh [--keep-state]

Runs a non-destructive dependency-wave dry run:
  validate → plan → init → complete spec → ready contracts → status

Removes .worktrees/waves/_hygiene-demo/ unless --keep-state.
Does not create git worktrees or merge branches.
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --keep-state) CLEANUP=0; shift ;;
    -h|--help) usage; exit 0 ;;
    *) echo "demo-agent-workflows: unknown option: $1" >&2; usage; exit 1 ;;
  esac
done

WAVES="$ROOT/scripts/check-dependency-waves.sh"
[[ -x "$WAVES" ]] || { echo "demo-agent-workflows: missing $WAVES" >&2; exit 1; }

echo "== Agent workflow demo (task: $TASK) =="
"$WAVES" validate
"$WAVES" plan
"$WAVES" init --task "$TASK" --base HEAD
"$WAVES" complete --task "$TASK" --wave spec --note "demo-agent-workflows.sh dry run"
"$WAVES" ready --task "$TASK" --wave contracts
"$WAVES" status --task "$TASK"

if [[ "$CLEANUP" -eq 1 ]]; then
  rm -rf "${AGENT_WORKTREES_DIR:-$ROOT/.worktrees}/waves/${TASK}"
  echo "Cleaned demo wave state for $TASK"
fi

echo "Agent workflow demo: PASS"
