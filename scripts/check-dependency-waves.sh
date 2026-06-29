#!/usr/bin/env bash
# Dependency-wave orchestration for parallel agent parents (reads .cursor/worktrees.config.json).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
# shellcheck source=agent-worktree/lib.sh
source "$ROOT/scripts/agent-worktree/lib.sh"

CONFIG="${AWT_CONFIG:-$ROOT/.cursor/worktrees.config.json}"
CONFIG_PY="$(resolve_agent_worktree_config_py "$ROOT" || true)"
WAVES_DIR="${AGENT_WORKTREES_DIR:-$ROOT/.worktrees}/waves"

usage() {
  cat <<'EOF'
Usage: ./scripts/check-dependency-waves.sh <command> [options]

Commands:
  validate          Validate waves DAG in worktrees.config.json
  plan              Print ordered waves and dependencies
  init --task ID    Create wave state file for a parent-orchestrated task
  status --task ID  Show wave completion / blocked status
  ready --task ID --wave WAVE   Exit 0 when dependencies for WAVE are satisfied
  complete --task ID --wave WAVE [--branch BRANCH] [--note TEXT]
  skip --task ID --wave WAVE [--note TEXT]   Mark optional wave skipped (parent only)
  handoff --task ID --wave WAVE   Print paste-ready child spawn block

State files live under .worktrees/waves/<task>/state.json (gitignored).

Examples:
  ./scripts/check-dependency-waves.sh plan
  ./scripts/check-dependency-waves.sh init --task lcr-alert-refactor
  ./scripts/check-dependency-waves.sh ready --task lcr-alert-refactor --wave backend
  ./scripts/check-dependency-waves.sh complete --task lcr-alert-refactor --wave contracts --branch agent/docs/contracts
  ./scripts/check-dependency-waves.sh handoff --task lcr-alert-refactor --wave backend
EOF
}

die() {
  echo "check-dependency-waves: $*" >&2
  exit 1
}

require_config() {
  [[ -f "$CONFIG" ]] || die "missing $CONFIG"
  [[ -n "$CONFIG_PY" && -f "$CONFIG_PY" ]] || die "missing scripts/agent-worktree/config.py (or set CURSOR_AGENT_WORKTREE_CONFIG_PY)"
}

state_file() {
  local task="$1"
  echo "$WAVES_DIR/$(printf '%s' "$task" | tr '[:upper:]' '[:lower:]' | tr -cs 'a-z0-9._-' '-')/state.json"
}

cmd_validate() {
  require_config
  python3 "$CONFIG_PY" waves validate "$CONFIG"
}

cmd_plan() {
  require_config
  python3 "$CONFIG_PY" waves plan "$CONFIG"
}

cmd_init() {
  local task="" base="HEAD"
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --task) task="${2:-}"; shift 2 ;;
      --base) base="${2:-}"; shift 2 ;;
      *) die "unknown option: $1" ;;
    esac
  done
  [[ -n "$task" ]] || die "usage: init --task ID [--base REF]"
  require_config
  local sf
  sf="$(state_file "$task")"
  mkdir -p "$(dirname "$sf")"
  if [[ -f "$sf" ]]; then
    echo "Wave state already exists: $sf"
    exit 0
  fi
  python3 -c "
import json, datetime, pathlib, sys
sf = pathlib.Path(sys.argv[1])
task = sys.argv[2]
base = sys.argv[3]
now = datetime.datetime.now(datetime.timezone.utc).replace(microsecond=0).isoformat().replace('+00:00', 'Z')
sf.parent.mkdir(parents=True, exist_ok=True)
sf.write_text(json.dumps({
  'version': 1,
  'task': task,
  'base': base,
  'created_at': now,
  'completed': {},
}, indent=2) + '\n')
print(f'Initialized wave state: {sf}')
" "$sf" "$task" "$base"
}

_load_state() {
  local sf="$1"
  [[ -f "$sf" ]] || die "wave state not found: $sf (run init --task first)"
}

cmd_status() {
  local task=""
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --task) task="${2:-}"; shift 2 ;;
      *) die "unknown option: $1" ;;
    esac
  done
  [[ -n "$task" ]] || die "usage: status --task ID"
  require_config
  _load_state "$(state_file "$task")"
  python3 "$CONFIG_PY" waves status "$CONFIG" --task "$task"
}

cmd_ready() {
  local task="" wave=""
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --task) task="${2:-}"; shift 2 ;;
      --wave) wave="${2:-}"; shift 2 ;;
      *) die "unknown option: $1" ;;
    esac
  done
  [[ -n "$task" && -n "$wave" ]] || die "usage: ready --task ID --wave WAVE"
  require_config
  _load_state "$(state_file "$task")"
  python3 "$CONFIG_PY" waves ready "$wave" "$CONFIG" --task "$task"
}

cmd_complete() {
  local task="" wave="" branch="" note=""
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --task) task="${2:-}"; shift 2 ;;
      --wave) wave="${2:-}"; shift 2 ;;
      --branch) branch="${2:-}"; shift 2 ;;
      --note) note="${2:-}"; shift 2 ;;
      *) die "unknown option: $1" ;;
    esac
  done
  [[ -n "$task" && -n "$wave" ]] || die "usage: complete --task ID --wave WAVE [--branch BRANCH]"
  require_config
  local sf
  sf="$(state_file "$task")"
  _load_state "$sf"
  python3 -c "
import json, datetime, pathlib, sys
sf = pathlib.Path(sys.argv[1])
wave, branch, note = sys.argv[2], sys.argv[3], sys.argv[4]
now = datetime.datetime.now(datetime.timezone.utc).replace(microsecond=0).isoformat().replace('+00:00', 'Z')
state = json.loads(sf.read_text())
completed = state.setdefault('completed', {})
entry = {
  'merged': True,
  'at': now,
}
if branch:
  entry['branch'] = branch
if note:
  entry['note'] = note
completed[wave] = entry
sf.write_text(json.dumps(state, indent=2) + '\n')
print(f'Marked wave complete: {wave}')
" "$sf" "$wave" "$branch" "$note"
}

cmd_skip() {
  local task="" wave="" note=""
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --task) task="${2:-}"; shift 2 ;;
      --wave) wave="${2:-}"; shift 2 ;;
      --note) note="${2:-}"; shift 2 ;;
      *) die "unknown option: $1" ;;
    esac
  done
  [[ -n "$task" && -n "$wave" ]] || die "usage: skip --task ID --wave WAVE"
  require_config
  local sf
  sf="$(state_file "$task")"
  _load_state "$sf"
  python3 -c "
import json, datetime, pathlib, sys
sf = pathlib.Path(sys.argv[1])
wave, note = sys.argv[2], sys.argv[3]
now = datetime.datetime.now(datetime.timezone.utc).replace(microsecond=0).isoformat().replace('+00:00', 'Z')
state = json.loads(sf.read_text())
state.setdefault('completed', {})[wave] = {
  'skipped': True,
  'at': now,
  **({'note': note} if note else {}),
}
sf.write_text(json.dumps(state, indent=2) + '\n')
print(f'Marked wave skipped: {wave}')
" "$sf" "$wave" "$note"
}

cmd_handoff() {
  local task="" wave=""
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --task) task="${2:-}"; shift 2 ;;
      --wave) wave="${2:-}"; shift 2 ;;
      *) die "unknown option: $1" ;;
    esac
  done
  [[ -n "$task" && -n "$wave" ]] || die "usage: handoff --task ID --wave WAVE"
  require_config
  _load_state "$(state_file "$task")"
  if ! python3 "$CONFIG_PY" waves ready "$wave" "$CONFIG" --task "$task"; then
    die "wave $wave not ready — complete or skip prior dependencies first"
  fi
  python3 "$CONFIG_PY" waves handoff "$wave" "$CONFIG" "$task"
}

main() {
  local cmd="${1:-}"
  shift || true
  case "$cmd" in
    validate) cmd_validate "$@" ;;
    plan) cmd_plan "$@" ;;
    init) cmd_init "$@" ;;
    status) cmd_status "$@" ;;
    ready) cmd_ready "$@" ;;
    complete) cmd_complete "$@" ;;
    skip) cmd_skip "$@" ;;
    handoff) cmd_handoff "$@" ;;
    -h|--help|"") usage ;;
    *) die "unknown command: $cmd (see --help)" ;;
  esac
}

main "$@"
