#!/usr/bin/env bash
# Shared helpers for agent worktree scripts (repo-local config.py for CI).
resolve_agent_worktree_config_py() {
  local root="$1"
  if [[ -n "${CURSOR_AGENT_WORKTREE_CONFIG_PY:-}" && -f "${CURSOR_AGENT_WORKTREE_CONFIG_PY}" ]]; then
    printf '%s\n' "${CURSOR_AGENT_WORKTREE_CONFIG_PY}"
    return 0
  fi
  if [[ -f "$root/scripts/agent-worktree/config.py" ]]; then
    printf '%s\n' "$root/scripts/agent-worktree/config.py"
    return 0
  fi
  if [[ -f "$HOME/.cursor/scripts/agent-worktree/config.py" ]]; then
    printf '%s\n' "$HOME/.cursor/scripts/agent-worktree/config.py"
    return 0
  fi
  return 1
}
