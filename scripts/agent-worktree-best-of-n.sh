#!/usr/bin/env bash
# Best-of-N wrapper: create N isolated git worktrees + manifest for parallel agent attempts.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

_load_awt_lib() {
  local lib="${CURSOR_AGENT_WORKTREE_LIB:-$HOME/.cursor/scripts/agent-worktree/lib.sh}"
  if [[ -f "$lib" ]]; then
    # shellcheck source=/dev/null
    source "$lib"
    awt_init
    ROOT="$AWT_ROOT"
    WORKTREES_DIR="$AWT_WORKTREES_DIR"
    die() { awt_die "$@"; }
    slugify() { awt_slugify "$@"; }
    normalize_prefix() { awt_normalize_prefix "$@"; }
    return 0
  fi
  return 1
}

if ! _load_awt_lib; then
  WORKTREES_DIR="${AGENT_WORKTREES_DIR:-$ROOT/.worktrees}"

  die() {
    echo "agent-worktree-best-of-n: $*" >&2
    exit 1
  }

  slugify() {
    printf '%s' "$1" | tr '[:upper:]' '[:lower:]' | tr -cs 'a-z0-9._-' '-' | sed 's/^-*//;s/-*$//'
  }

  normalize_prefix() {
    local p
    p="$(slugify "$1")"
    p="${p#bon-}"
    printf '%s' "$p"
  }
fi

WT="$ROOT/scripts/agent-worktree.sh"

usage() {
  cat <<'EOF'
Usage: ./scripts/agent-worktree-best-of-n.sh <command> [options]

Commands:
  create    Create N worktrees + shared manifest and optional task file
  status    Show manifest and worktree health
  compare   Diff stats for each attempt vs --base
  handoffs  Print paste-ready prompts for all attempts
  remove    Remove all worktrees in a batch (--force for dirty trees)

Create options:
  --n N             Number of parallel attempts (default: 3, max: 8)
  --prefix NAME     Batch id (default: bon-YYYYMMDD-HHMMSS)
  --track TRACK     backend | frontend | docs | experiment (default: experiment)
  --base REF        Branch or commit to branch from (default: HEAD)
  --task-file PATH  Copy task spec into batch dir as task.md
  --task TEXT       Inline task spec (mutually exclusive with --task-file)

Shared options:
  --prefix NAME     Batch id (required for status/compare/handoffs/remove)

Examples:
  ./scripts/agent-worktree-best-of-n.sh create --n 3 --prefix audit-refactor --track backend \
    --task "Refactor audit consumer error handling; keep tests green."
  ./scripts/agent-worktree-best-of-n.sh compare --prefix audit-refactor
  ./scripts/agent-worktree-best-of-n.sh handoffs --prefix audit-refactor
  ./scripts/agent-worktree-best-of-n.sh remove --prefix audit-refactor

After create:
  - Open N fresh Cursor chats (or Cursor multi-agent UI) at the printed paths
  - Or delegate each attempt to built-in best-of-n-runner Task subagents
  - Parent compares, picks winner, merges, runs integration smoke from main root
EOF
}

batch_dir_for_prefix() {
  echo "$WORKTREES_DIR/bon-$(normalize_prefix "$1")"
}

manifest_path() {
  echo "$(batch_dir_for_prefix "$1")/manifest.json"
}

require_manifest() {
  local prefix="$1"
  local mp
  mp="$(manifest_path "$prefix")"
  [[ -f "$mp" ]] || die "manifest not found for prefix '$prefix' (expected $mp)"
  printf '%s\n' "$mp"
}

cmd_create() {
  local n=3 prefix="" track="experiment" base="HEAD" task_file="" task_text="" force=0
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --n) n="${2:-}"; shift 2 ;;
      --prefix) prefix="${2:-}"; shift 2 ;;
      --track) track="${2:-}"; shift 2 ;;
      --base) base="${2:-}"; shift 2 ;;
      --task-file) task_file="${2:-}"; shift 2 ;;
      --task) task_text="${2:-}"; shift 2 ;;
      --force) force=1; shift ;;
      -h|--help) usage; exit 0 ;;
      *) die "unknown create option: $1" ;;
    esac
  done

  [[ "$n" =~ ^[0-9]+$ ]] || die "--n must be a positive integer"
  (( n >= 1 && n <= 8 )) || die "--n must be between 1 and 8 (Cursor multi-agent limit)"

  if [[ -n "$task_file" && -n "$task_text" ]]; then
    die "use only one of --task-file or --task"
  fi
  if [[ -n "$task_file" && ! -f "$task_file" ]]; then
    die "task file not found: $task_file"
  fi

  if [[ -z "$prefix" ]]; then
    prefix="bon-$(date +%Y%m%d-%H%M%S)"
  fi
  prefix="$(normalize_prefix "$prefix")"

  local batch_dir mp
  batch_dir="$(batch_dir_for_prefix "$prefix")"
  mp="$batch_dir/manifest.json"

  if [[ -e "$batch_dir" && "$force" -eq 0 ]]; then
    die "batch dir exists: $batch_dir (pass --force to recreate metadata only; remove batch first if worktrees exist)"
  fi
  mkdir -p "$batch_dir"

  if [[ -n "$task_file" ]]; then
    cp "$task_file" "$batch_dir/task.md"
  elif [[ -n "$task_text" ]]; then
    printf '%s\n' "$task_text" >"$batch_dir/task.md"
  fi

  local attempts_json="["
  local i name branch path
  for ((i = 1; i <= n; i++)); do
    name="bon-${prefix}-${i}"
    branch="agent/${track}/${name}"
    path="$("$WT" create --track "$track" --name "$name" --base "$base" --branch "$branch" --print-path)"
    if [[ -f "$batch_dir/task.md" ]]; then
      cp "$batch_dir/task.md" "$path/.agent-bon-task.md"
    fi
    [[ "$i" -gt 1 ]] && attempts_json+=","
    attempts_json+=$(python3 -c "
import json, sys
print(json.dumps({
  'index': int(sys.argv[1]),
  'name': sys.argv[2],
  'branch': sys.argv[3],
  'path': sys.argv[4],
}))
" "$i" "$name" "$branch" "$path")
  done
  attempts_json+="]"

  local created
  created="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
  python3 -c "
import json, pathlib, sys
batch = pathlib.Path(sys.argv[1])
data = {
  'prefix': sys.argv[2],
  'n': int(sys.argv[3]),
  'track': sys.argv[4],
  'base': sys.argv[5],
  'created': sys.argv[6],
  'task_file': 'task.md' if (batch / 'task.md').is_file() else None,
  'attempts': json.loads(sys.argv[7]),
}
(batch / 'manifest.json').write_text(json.dumps(data, indent=2) + '\n')
" "$batch_dir" "$prefix" "$n" "$track" "$base" "$created" "$attempts_json"

  cat <<EOF
Created best-of-$n batch: bon-${prefix}
  manifest: $mp
  base:     $base
  track:    $track

Attempts:
EOF
  python3 -c "
import json, sys
m = json.load(open(sys.argv[1]))
for a in m['attempts']:
    print(f\"  [{a['index']}] {a['path']}  ({a['branch']})\")
" "$mp"

  cat <<'EOF'

Next:
  1. ./scripts/agent-worktree-best-of-n.sh handoffs --prefix PREFIX
  2. Launch N fresh chats (or Cursor multi-agent) at the paths above
  3. Optional: Task subagent best-of-n-runner per attempt (same task, distinct approach)
  4. ./scripts/agent-worktree-best-of-n.sh compare --prefix PREFIX
  5. Merge winner on main root; parent runs integration smoke; then remove batch
EOF
  echo "  (use --prefix $prefix or --prefix bon-$prefix)"
}

cmd_status() {
  local prefix=""
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --prefix) prefix="${2:-}"; shift 2 ;;
      *) die "unknown status option: $1" ;;
    esac
  done
  [[ -n "$prefix" ]] || die "usage: agent-worktree-best-of-n.sh status --prefix NAME"
  prefix="$(normalize_prefix "$prefix")"
  local mp
  mp="$(require_manifest "$prefix")"
  python3 -c "
import json, pathlib, subprocess, sys
m = json.load(open(sys.argv[1]))
print(f\"Batch bon-{m['prefix']}: n={m['n']} track={m['track']} base={m['base']} created={m['created']}\")
if m.get('task_file'):
    print(f\"Task: {(pathlib.Path(sys.argv[1]).parent / m['task_file'])}\")
for a in m['attempts']:
    p = pathlib.Path(a['path'])
    exists = p.is_dir()
    dirty = 'unknown'
    if exists:
        r = subprocess.run(['git', '-C', str(p), 'status', '--porcelain'], capture_output=True, text=True)
        dirty = 'dirty' if r.stdout.strip() else 'clean'
    print(f\"  [{a['index']}] exists={exists} {dirty} {a['branch']} {a['path']}\")
" "$mp"
}

cmd_compare() {
  local prefix="" base_override=""
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --prefix) prefix="${2:-}"; shift 2 ;;
      --base) base_override="${2:-}"; shift 2 ;;
      *) die "unknown compare option: $1" ;;
    esac
  done
  [[ -n "$prefix" ]] || die "usage: agent-worktree-best-of-n.sh compare --prefix NAME [--base REF]"
  prefix="$(normalize_prefix "$prefix")"
  local mp scope_fail=0 compare_base
  mp="$(require_manifest "$prefix")"
  local cfg="$ROOT/.cursor/worktrees.config.json"
  compare_base="${base_override:-$(python3 -c "import json; print(json.load(open('$mp'))['base'])")}"
  python3 -c "
import json, subprocess, sys
m = json.load(open(sys.argv[1]))
base = sys.argv[2] or m['base']
print(f'Compare vs base: {base}')
for a in m['attempts']:
    branch = a['branch']
    print(f\"\\n=== attempt {a['index']}: {branch} ===\")
    r = subprocess.run(['git', 'diff', '--stat', f'{base}...{branch}'], capture_output=True, text=True)
    out = (r.stdout or r.stderr).strip()
    print(out if out else '(no diff vs base)')
    r2 = subprocess.run(['git', '-C', a['path'], 'log', '--oneline', '-3'], capture_output=True, text=True)
    if r2.stdout.strip():
        print('Recent commits:')
        print(r2.stdout.rstrip())
" "$mp" "$base_override"
  if [[ -x "$ROOT/scripts/check-worktree-scope.sh" && -f "$cfg" ]]; then
    echo ""
    echo "== scope (all attempts) =="
    while IFS= read -r branch; do
      [[ -z "$branch" ]] && continue
      echo "-- $branch --"
      if ! "$ROOT/scripts/check-worktree-scope.sh" --branch "$branch" --base "$compare_base" --strict; then
        scope_fail=1
      fi
    done < <(python3 -c "import json,sys; [print(a['branch']) for a in json.load(open(sys.argv[1]))['attempts']]" "$mp")
    [[ "$scope_fail" -eq 0 ]] || die "one or more attempts failed scope check"
  fi
}

cmd_handoffs() {
  local prefix=""
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --prefix) prefix="${2:-}"; shift 2 ;;
      *) die "unknown handoffs option: $1" ;;
    esac
  done
  [[ -n "$prefix" ]] || die "usage: agent-worktree-best-of-n.sh handoffs --prefix NAME"
  prefix="$(normalize_prefix "$prefix")"
  local mp
  mp="$(require_manifest "$prefix")"

  python3 -c "
import json, pathlib, sys
m = json.load(open(sys.argv[1]))
batch = pathlib.Path(sys.argv[1]).parent
task = (batch / 'task.md').read_text() if m.get('task_file') else None
n = m['n']
for a in m['attempts']:
    print('=' * 72)
    print(f\"HANDOFF attempt {a['index']} of {n} — bon-{m['prefix']}\")
    print('=' * 72)
    print(f\"Workspace: {a['path']}\")
    print(f\"Branch: {a['branch']}\")
    print(f\"Track: {m['track']}\")
    print()
    print('Best-of-N: pursue a **distinct** approach from the other attempts. Do not read other worktrees.')
    if task:
        print()
        print('## Task')
        print(task.rstrip())
    print()
    print('Run from parent for full handoff block:')
    print(f\"  ./scripts/agent-worktree.sh handoff {a['name']}\")
    print()
" "$mp"
}

cmd_remove() {
  local prefix="" force=0
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --prefix) prefix="${2:-}"; shift 2 ;;
      --force) force=1; shift ;;
      *) die "unknown remove option: $1" ;;
    esac
  done
  [[ -n "$prefix" ]] || die "usage: agent-worktree-best-of-n.sh remove --prefix NAME [--force]"
  prefix="$(normalize_prefix "$prefix")"
  local mp
  mp="$(require_manifest "$prefix")"

  python3 -c "
import json, subprocess, sys
m = json.load(open(sys.argv[1]))
force = sys.argv[2] == '1'
wt = sys.argv[3]
for a in m['attempts']:
    cmd = [wt, 'remove', a['name']]
    if force:
        cmd.append('--force')
    print('Removing', a['name'], '...')
    subprocess.run(cmd, check=True)
" "$mp" "$force" "$WT"

  local batch_dir
  batch_dir="$(batch_dir_for_prefix "$prefix")"
  rm -f "$mp"
  rmdir "$batch_dir" 2>/dev/null || true
  echo "Removed best-of-N batch bon-${prefix}"
}

main() {
  [[ -x "$WT" ]] || die "missing executable: $WT"
  local cmd="${1:-}"
  shift || true
  case "$cmd" in
    create) cmd_create "$@" ;;
    status) cmd_status "$@" ;;
    compare) cmd_compare "$@" ;;
    handoffs) cmd_handoffs "$@" ;;
    remove) cmd_remove "$@" ;;
    -h|--help|help|"") usage ;;
    *) die "unknown command: $cmd (try --help)" ;;
  esac
}

main "$@"
