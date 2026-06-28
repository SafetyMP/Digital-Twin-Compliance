#!/usr/bin/env bash
# Create and manage git worktrees for isolated Cursor agent sessions.
# Outputs absolute paths for cursor-app-control move_agent_to_root / new chats.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

WORKTREES_DIR="${AGENT_WORKTREES_DIR:-$ROOT/.worktrees}"

usage() {
  cat <<'EOF'
Usage: ./scripts/agent-worktree.sh <command> [options]

Commands:
  create   Create a worktree + branch for an agent session
  list     List agent worktrees under .worktrees/
  path     Print absolute path for a worktree name
  handoff  Print a paste-ready handoff block (includes path + verify steps)
  remove   Remove a worktree (refuses if dirty unless --force)

Create options:
  --name NAME       Worktree directory name (default: slug from --track)
  --track TRACK     backend | frontend | docs | experiment (default: experiment)
  --base REF        Branch or commit to branch from (default: current HEAD)
  --branch NAME     Git branch name (default: agent/<track>/<name>)
  --print-path      Print only the absolute worktree path (for scripting)

Remove options:
  --force           Remove even when the worktree has uncommitted changes

Examples:
  ./scripts/agent-worktree.sh create --track backend --name state-outbox
  ./scripts/agent-worktree.sh handoff state-outbox
  ./scripts/agent-worktree.sh list
  ./scripts/agent-worktree.sh remove state-outbox

After create/handoff, open a fresh Cursor chat at the printed path or call
cursor-app-control move_agent_to_root with the absolute path.
EOF
}

die() {
  echo "agent-worktree: $*" >&2
  exit 1
}

slugify() {
  printf '%s' "$1" | tr '[:upper:]' '[:lower:]' | tr -cs 'a-z0-9._-' '-' | sed 's/^-*//;s/-*$//'
}

require_git() {
  git rev-parse --is-inside-work-tree >/dev/null 2>&1 || die "not inside a git repository"
}

worktree_path_for_name() {
  local name="$1"
  echo "$WORKTREES_DIR/$(slugify "$name")"
}

resolve_existing_path() {
  local name="$1"
  local path
  path="$(worktree_path_for_name "$name")"
  if [[ -d "$path" ]]; then
    echo "$path"
    return 0
  fi
  while IFS= read -r line; do
    [[ -z "$line" ]] && continue
    if [[ "$(basename "$line")" == "$(slugify "$name")" ]]; then
      echo "$line"
      return 0
    fi
  done < <(git worktree list --porcelain | awk '/^worktree / {print substr($0, 10)}')
  return 1
}

cmd_create() {
  local name="" track="experiment" base="" branch="" print_path=0
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --name) name="${2:-}"; shift 2 ;;
      --track) track="${2:-}"; shift 2 ;;
      --base) base="${2:-}"; shift 2 ;;
      --branch) branch="${2:-}"; shift 2 ;;
      --print-path) print_path=1; shift ;;
      -h|--help) usage; exit 0 ;;
      *) die "unknown create option: $1" ;;
    esac
  done

  case "$track" in
    backend|frontend|docs|experiment) ;;
    *) die "invalid --track '$track' (use backend, frontend, docs, or experiment)" ;;
  esac

  if [[ "$track" == "integration" ]]; then
    die "integration track is blocked — use serial work on docker-compose.dev.yml (see AGENTS.md § Parallel work)"
  fi

  if [[ -z "$name" ]]; then
    name="${track}-$(date +%Y%m%d-%H%M%S)"
  fi
  name="$(slugify "$name")"

  if [[ -z "$branch" ]]; then
    branch="agent/${track}/${name}"
  fi

  local path
  path="$(worktree_path_for_name "$name")"
  mkdir -p "$WORKTREES_DIR"

  if [[ -e "$path" ]]; then
    die "worktree path already exists: $path"
  fi

  local base_ref="${base:-HEAD}"
  git show-ref --verify --quiet "refs/heads/${branch}" 2>/dev/null && die "branch already exists: $branch"

  git worktree add -b "$branch" "$path" "$base_ref" >/dev/null

  if [[ "$print_path" -eq 1 ]]; then
    printf '%s\n' "$path"
    return 0
  fi

  cat <<EOF
Created agent worktree
  path:   $path
  branch: $branch
  base:   $base_ref
  track:  $track

Next steps:
  1. Open a fresh Cursor chat with workspace root: $path
     (or cursor-app-control move_agent_to_root rootPath=$path)
  2. Paste handoff: ./scripts/agent-worktree.sh handoff $name

Do not run docker compose from multiple worktrees on the same host ports.
Parent merges branch after unit tests; parent owns smoke-test-phase2.sh.
EOF
}

cmd_list() {
  if [[ ! -d "$WORKTREES_DIR" ]]; then
    echo "No agent worktrees directory ($WORKTREES_DIR)."
    exit 0
  fi
  echo "Agent worktrees under $WORKTREES_DIR:"
  git worktree list | while read -r line; do
    case "$line" in
      *"$WORKTREES_DIR"*) echo "  $line" ;;
    esac
  done
}

cmd_path() {
  local name="${1:-}"
  [[ -n "$name" ]] || die "usage: agent-worktree.sh path NAME"
  local path
  path="$(resolve_existing_path "$name")" || die "worktree not found: $name"
  printf '%s\n' "$path"
}

track_scope() {
  case "$1" in
    backend)
      cat <<'EOF'
- Scope: services/* and jobs/* only (no apps/, no docker-compose.dev.yml)
- Do not start docker compose from this worktree
EOF
      ;;
    frontend)
      cat <<'EOF'
- Scope: apps/* only (use Next.js /api proxies; no direct :8085/:8090 from browser)
- Do not start docker compose from this worktree
EOF
      ;;
    docs)
      cat <<'EOF'
- Scope: docs/, policies/, evals/ docs only — no service code unless explicitly requested
EOF
      ;;
    *)
      cat <<'EOF'
- Scope: confirm file boundaries with parent before editing
- Do not start docker compose from this worktree unless parent owns port coordination
EOF
      ;;
  esac
}

infer_track_from_branch() {
  local branch="$1"
  case "$branch" in
    agent/backend/*) echo backend ;;
    agent/frontend/*) echo frontend ;;
    agent/docs/*) echo docs ;;
    *) echo experiment ;;
  esac
}

cmd_handoff() {
  local name="${1:-}"
  [[ -n "$name" ]] || die "usage: agent-worktree.sh handoff NAME"
  local path branch track
  path="$(resolve_existing_path "$name")" || die "worktree not found: $name"
  branch="$(git -C "$path" branch --show-current 2>/dev/null || true)"
  track="$(infer_track_from_branch "${branch:-}")"

  cat <<EOF
Worktree agent session — Digital Twin

Workspace root (absolute): $path
Branch: ${branch:-unknown}
Track: $track

Do NOT read agent-transcripts. Context budget: AGENTS.md + scoped service AGENTS.md + phase spec only.

## Scope (non-overlapping)
$(track_scope "$track")

## Parent owns
- Merge/rebase of branch $branch after your unit tests pass
- docker compose, seed, smoke-test.sh, smoke-test-phase2.sh from the main repo root
- Conflict resolution if two tracks touch the same path

## Verify in this worktree before handoff back
EOF

  case "$track" in
    backend)
      echo "- cd services/state-service && go test ./...  (when touching state-service)"
      echo "- cd services/alert-service && go test ./...   (when touching alert-service)"
      echo "- cd jobs/compliance-cep && mvn test           (when touching CEP)"
      ;;
    frontend)
      echo "- cd apps/alert-console && npm test            (when touching alert-console)"
      echo "- cd apps/audit-explorer && npm test           (when touching audit-explorer)"
      ;;
    *)
      echo "- Package tests for every touched code path (see AGENTS.md § Verification floor)"
      ;;
  esac

  cat <<EOF

## Done / Blocked / Next
- (fill in before returning to parent)

Return: branch name, files touched, test commands run + exit codes. Before handoff back, parent may run:
  ./scripts/check-worktree-scope.sh --branch $branch --strict
Parent runs verify-worktree-merge + smoke on main root.
EOF
}

cmd_remove() {
  local name="" force=0
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --force) force=1; shift ;;
      -*) die "unknown remove option: $1" ;;
      *)
        [[ -z "$name" ]] || die "unexpected argument: $1"
        name="$1"
        shift
        ;;
    esac
  done
  [[ -n "$name" ]] || die "usage: agent-worktree.sh remove NAME [--force]"

  local path
  path="$(resolve_existing_path "$name")" || die "worktree not found: $name"

  branch="$(git -C "$path" branch --show-current 2>/dev/null || true)"

  if [[ "$force" -eq 0 ]] && [[ -n "$branch" ]]; then
    if ! git merge-base --is-ancestor "$branch" HEAD 2>/dev/null; then
      die "branch $branch is not merged into HEAD; merge first or pass --force"
    fi
  fi

  if [[ "$force" -eq 0 ]]; then
    if ! git -C "$path" diff --quiet 2>/dev/null || ! git -C "$path" diff --cached --quiet 2>/dev/null; then
      die "worktree has uncommitted changes; commit/stash or pass --force"
    fi
  fi

  local branch
  branch="$(git -C "$path" branch --show-current 2>/dev/null || true)"
  git worktree remove "$path" --force
  if [[ -n "$branch" ]] && [[ "$branch" == agent/* ]]; then
    git branch -D "$branch" 2>/dev/null || true
  fi
  echo "Removed worktree $path${branch:+ (branch $branch deleted)}"
}

main() {
  require_git
  local cmd="${1:-}"
  shift || true
  case "$cmd" in
    create) cmd_create "$@" ;;
    list) cmd_list "$@" ;;
    path) cmd_path "$@" ;;
    handoff) cmd_handoff "$@" ;;
    remove) cmd_remove "$@" ;;
    -h|--help|help|"") usage ;;
    *) die "unknown command: $cmd (try --help)" ;;
  esac
}

main "$@"
