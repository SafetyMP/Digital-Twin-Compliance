#!/usr/bin/env bash
# Hygiene checks for agent git worktrees under .worktrees/
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

WORKTREES_DIR="${AGENT_WORKTREES_DIR:-$ROOT/.worktrees}"

fail=0
ok() { echo "OK  $*"; }
warn() { echo "WARN $*"; }
bad() { echo "FAIL $*"; fail=1; }

echo "== Agent worktree hygiene =="

if [[ ! -d "$WORKTREES_DIR" ]]; then
  ok "no .worktrees directory (nothing to check)"
  exit 0
fi

# Manifest batches
manifest_count=0
while IFS= read -r -d '' mf; do
  manifest_count=$((manifest_count + 1))
  if ! python3 -c "
import json, sys, pathlib
m = json.load(open(sys.argv[1]))
root = pathlib.Path(sys.argv[2])
for a in m.get('attempts', []):
    p = pathlib.Path(a['path'])
    if not p.is_dir():
        print(f\"missing worktree path: {p}\")
        sys.exit(1)
sys.exit(0)
" "$mf" "$ROOT"; then
    bad "broken manifest: $mf"
  else
    ok "manifest valid: $mf"
  fi
done < <(find "$WORKTREES_DIR" -mindepth 2 -maxdepth 2 -name manifest.json -print0 2>/dev/null || true)

if [[ "$manifest_count" -eq 0 ]]; then
  ok "no best-of-N manifests"
fi

# agent/* branches without an active worktree (informational)
while read -r branch; do
  [[ -z "$branch" ]] && continue
  if ! git worktree list | grep -q "\[$branch\]"; then
    warn "agent branch without worktree: $branch (delete after merge)"
  fi
done < <(git branch --list 'agent/*' | sed 's/^[* ]*//')

# Script presence
for f in scripts/agent-worktree.sh scripts/agent-worktree-best-of-n.sh \
  scripts/check-worktree-scope.sh scripts/verify-worktree-merge.sh; do
  if [[ -x "$f" ]]; then ok "$f executable"; else bad "missing or not executable: $f"; fi
done

echo
if [[ "$fail" -eq 0 ]]; then
  echo "Agent worktree hygiene: PASS"
  exit 0
fi
echo "Agent worktree hygiene: FAIL"
exit 1
