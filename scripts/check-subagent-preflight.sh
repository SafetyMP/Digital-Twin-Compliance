#!/usr/bin/env bash
# Preflight checks before dispatching Cursor Task subagents in this repo.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

word_count() {
  wc -w < "$1" | tr -d ' '
}

fail=0
warn=0

ok() { echo "OK  $*"; }
note() { echo "NOTE $*"; }
bad() { echo "FAIL $*"; fail=1; }
w() { echo "WARN $*"; warn=1; }

echo "== Subagent preflight =="

agents_words="$(word_count "$ROOT/AGENTS.md")"
if [[ "$agents_words" -gt 2500 ]]; then
  w "AGENTS.md is ${agents_words} words — subagents inherit this always-on; keep Task prompts narrow and avoid preloading phase specs"
else
  ok "AGENTS.md word count ${agents_words} (<=2500 advisory)"
fi

note "Do not assign cold docker compose / smoke as a subagent's first action — parent runs integration"

invalid_models="$(rg -n 'model:\s*\"?(fast|default|inherit-only)\"?' "$ROOT" --glob '*.md' --glob '*.mdc' 2>/dev/null || true)"
if [[ -n "$invalid_models" ]]; then
  w "Possible invalid Task model references in repo docs — omit model or use an allowed slug only"
  echo "$invalid_models" | sed 's/^/      /'
else
  ok "no obvious invalid Task model slugs in repo markdown"
fi

if [[ -f "$HOME/.cursor/hooks/scan-prompt.py" ]]; then
  handoff_sample="Task subagent probe: explore services/state-service only. No secrets."
  scan_result="$(printf '%s' "$handoff_sample" | python3 -c "
import json, sys, os
sys.path.insert(0, os.path.expanduser('~/.cursor/hooks'))
from _common import find_secret
label = find_secret(sys.stdin.read())
print(json.dumps({'blocked': bool(label), 'label': label or ''}))
" 2>/dev/null || echo '{"blocked":false,"label":""}')"
  blocked_flag="$(echo "$scan_result" | python3 -c "import json,sys; print('yes' if json.load(sys.stdin).get('blocked') else 'no')" 2>/dev/null || echo "no")"
  if [[ "$blocked_flag" == "yes" ]]; then
    w "scan-prompt would block a sample handoff — redact secrets in Task prompts"
  else
    ok "scan-prompt dry-run passed on sample handoff text"
  fi
else
  note "scan-prompt hook not found — skip secret scan dry-run"
fi

if [[ -x "$ROOT/scripts/check-agent-worktrees.sh" ]]; then
  if ./scripts/check-agent-worktrees.sh >/dev/null 2>&1; then
    ok "agent worktree hygiene"
  else
    bad "agent worktree hygiene failed — run ./scripts/check-agent-worktrees.sh"
  fi
fi

if [[ -f "$HOME/.cursor/agents/best-of-n-runner.md" ]]; then
  note "best-of-n-runner has custom agent file (unexpected — usually Cursor built-in)"
elif rg -q 'best-of-n-runner' "$ROOT/docs" "$ROOT/AGENTS.md" 2>/dev/null; then
  ok "best-of-n-runner referenced — Cursor built-in Task type (no ~/.cursor/agents file required)"
fi

echo
if [[ "$fail" -eq 0 && "$warn" -eq 0 ]]; then
  echo "Subagent preflight: PASS"
  exit 0
fi
if [[ "$fail" -eq 0 ]]; then
  echo "Subagent preflight: PASS with warnings"
  exit 0
fi
echo "Subagent preflight: FAIL"
exit 1
