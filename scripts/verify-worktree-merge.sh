#!/usr/bin/env bash
# Post-merge verification on MAIN repo root — tests, contracts, rebuild, optional smoke.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
# shellcheck source=agent-worktree/lib.sh
source "$ROOT/scripts/agent-worktree/lib.sh"

CONFIG="${AWT_CONFIG:-$ROOT/.cursor/worktrees.config.json}"
CONFIG_PY="$(resolve_agent_worktree_config_py "$ROOT" || true)"
LIB="${CURSOR_AGENT_WORKTREE_LIB:-$HOME/.cursor/scripts/agent-worktree/lib.sh}"

# shellcheck source=/dev/null
[[ -f "$LIB" ]] && source "$LIB" && awt_assert_main_root

COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.dev.yml}"
if [[ -f "$CONFIG" ]]; then
  COMPOSE_FILE="$(python3 -c "import json; print(json.load(open('$CONFIG')).get('compose_file','$COMPOSE_FILE'))" 2>/dev/null || echo "$COMPOSE_FILE")"
fi

usage() {
  cat <<'EOF'
Usage: ./scripts/verify-worktree-merge.sh BRANCH [--base REF] [--rebuild] [--with-smoke] [--skip-tests]

Main checkout only. After merging an agent branch:
  scope → package tests → contracts → [--rebuild] → [--with-smoke]
EOF
}

die() {
  echo "verify-worktree-merge: $*" >&2
  exit 1
}

branch="" base="HEAD" with_smoke=0 skip_tests=0 do_rebuild=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --base) base="${2:-}"; shift 2 ;;
    --rebuild) do_rebuild=1; shift ;;
    --with-smoke) with_smoke=1; shift ;;
    --skip-tests) skip_tests=1; shift ;;
    -h|--help) usage; exit 0 ;;
    -*) die "unknown option: $1" ;;
    *)
      [[ -z "$branch" ]] || die "unexpected argument: $1"
      branch="$1"
      shift
      ;;
  esac
done

[[ -n "$branch" ]] || die "usage: verify-worktree-merge.sh BRANCH"

case "$ROOT" in
  */.worktrees/*) die "run from main repo root, not a worktree ($ROOT)" ;;
esac
[[ "$PWD" != *"/.worktrees/"* ]] || die "run from main repo root, not inside .worktrees/ ($PWD)"

git show-ref --verify --quiet "refs/heads/${branch}" 2>/dev/null || die "branch not found: $branch"

echo "== verify-worktree-merge: $branch vs $base =="

if [[ -x "$ROOT/scripts/check-worktree-scope.sh" ]]; then
  echo "-- scope --"
  "$ROOT/scripts/check-worktree-scope.sh" --branch "$branch" --base "$base" --strict || die "scope check failed"
fi

files=()
while IFS= read -r line; do
  [[ -n "$line" ]] && files+=("$line")
done < <(python3 "$CONFIG_PY" files "$CONFIG" "$branch" "$base" 2>/dev/null || true)
if [[ ${#files[@]} -eq 0 ]]; then
  while IFS= read -r line; do
    [[ -n "$line" ]] && files+=("$line")
  done < <(git diff --name-only "$base...$branch" 2>/dev/null || git diff --name-only "$base" "$branch")
fi

if [[ ${#files[@]} -eq 0 ]]; then
  echo "No file changes vs $base — nothing to verify."
  exit 0
fi

echo "Changed files (${#files[@]}):"
printf '  %s\n' "${files[@]}"

if [[ "$skip_tests" -eq 0 && -f "$CONFIG_PY" && -f "$CONFIG" ]]; then
  echo "-- package tests --"
  while IFS= read -r cmd; do
    [[ -z "$cmd" ]] && continue
    echo ">> $cmd"
    bash -c "$cmd"
  done < <(python3 "$CONFIG_PY" tests "$CONFIG" "${files[@]}")
fi

if [[ -f "$CONFIG_PY" && -f "$CONFIG" ]]; then
  echo "-- contract checks --"
  while IFS= read -r cmd; do
    [[ -z "$cmd" ]] && continue
    if [[ -x "$ROOT/${cmd%% *}" ]] || [[ -f "$ROOT/$cmd" ]]; then
      echo ">> $cmd"
      bash -c "cd '$ROOT' && $cmd"
    fi
  done < <(python3 "$CONFIG_PY" contracts "$CONFIG" "${files[@]}")
fi

unique_services=()
if [[ -f "$CONFIG_PY" && -f "$CONFIG" ]]; then
  while IFS= read -r svc; do
    [[ -z "$svc" ]] && continue
    unique_services+=("$svc")
  done < <(python3 "$CONFIG_PY" services "$CONFIG" "${files[@]}")
fi

if [[ ${#unique_services[@]} -gt 0 ]]; then
  build_cmd="docker compose -f $COMPOSE_FILE build ${unique_services[*]}"
  up_cmd="docker compose -f $COMPOSE_FILE up -d --wait ${unique_services[*]}"
  echo "-- rebuild --"
  echo "$build_cmd"
  echo "$up_cmd"
  if [[ "$do_rebuild" -eq 1 ]]; then
    echo ">> running rebuild"
    docker compose -f "$COMPOSE_FILE" build "${unique_services[@]}"
    docker compose -f "$COMPOSE_FILE" up -d --wait "${unique_services[@]}"
    if printf '%s\n' "${files[@]}" | grep -q '^jobs/compliance-cep/'; then
      echo ">> CEP changed — run: cd jobs/compliance-cep && mvn -q package -DskipTests && ./scripts/submit-flink-job.sh"
    fi
  else
    echo "Pass --rebuild to execute, required before --with-smoke when services changed."
  fi
fi

if [[ "$with_smoke" -eq 1 ]]; then
  if [[ ${#unique_services[@]} -gt 0 && "$do_rebuild" -eq 0 ]]; then
    die "refusing --with-smoke without --rebuild when compose services changed (stale container risk)"
  fi
  echo "-- integration smoke (main root) --"
  ./scripts/smoke-test.sh
  [[ -f scripts/smoke-test-phase2.sh ]] && ./scripts/smoke-test-phase2.sh
  [[ -f scripts/smoke-test-phase3.sh ]] && SMOKE_PHASE3_SKIP_PREREQS=1 ./scripts/smoke-test-phase3.sh
elif [[ "$do_rebuild" -eq 1 ]]; then
  echo "Tip: re-run with --with-smoke for integration verify."
fi

echo "verify-worktree-merge: PASS (branch $branch)"
