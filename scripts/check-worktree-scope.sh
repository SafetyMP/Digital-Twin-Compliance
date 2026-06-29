#!/usr/bin/env bash
# Soft gate: flag files changed on an agent branch outside track scope (reads .cursor/worktrees.config.json).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
# shellcheck source=agent-worktree/lib.sh
source "$ROOT/scripts/agent-worktree/lib.sh"

CONFIG="${AWT_CONFIG:-$ROOT/.cursor/worktrees.config.json}"
CONFIG_PY="$(resolve_agent_worktree_config_py "$ROOT" || true)"

usage() {
  cat <<'EOF'
Usage: ./scripts/check-worktree-scope.sh --branch BRANCH [--base REF] [--track TRACK] [--strict]
EOF
}

die() {
  echo "check-worktree-scope: $*" >&2
  exit 1
}

branch="" base="HEAD" track="" strict=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --branch) branch="${2:-}"; shift 2 ;;
    --base) base="${2:-}"; shift 2 ;;
    --track) track="${2:-}"; shift 2 ;;
    --strict) strict=1; shift ;;
    -h|--help) usage; exit 0 ;;
    *) die "unknown option: $1" ;;
  esac
done

[[ -n "$branch" ]] || die "usage: check-worktree-scope.sh --branch BRANCH"
git show-ref --verify --quiet "refs/heads/${branch}" 2>/dev/null || die "branch not found: $branch"

if [[ -f "$CONFIG" && -f "$CONFIG_PY" ]]; then
  args=(python3 "$CONFIG_PY" scope "$CONFIG" "$branch" "$base")
  [[ -n "$track" ]] && args+=("$track")
  [[ "$strict" -eq 1 ]] && args+=("--strict")
  exec "${args[@]}"
fi

die "missing $CONFIG or scripts/agent-worktree/config.py (or set CURSOR_AGENT_WORKTREE_CONFIG_PY)"
