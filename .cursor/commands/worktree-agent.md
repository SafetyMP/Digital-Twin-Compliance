---
description: Create an isolated git worktree for a parallel agent track (backend, frontend, docs).
---
Start or continue an **isolated agent session** on a git worktree per [docs/handoff-worktree-agent.md](../../docs/handoff-worktree-agent.md).

**Default remains serial** — use worktrees only when tracks have non-overlapping file boundaries (see [AGENTS.md](../../AGENTS.md) § Parallel work). Parent owns docker compose and smoke tests from the main repo root.

## Create a worktree

Ask the user for track (`backend` | `frontend` | `docs` | `experiment`) and a short name, then run:

```bash
./scripts/agent-worktree.sh create --track <track> --name <name>
./scripts/agent-worktree.sh handoff <name>
```

## Move this agent to the worktree (optional)

If continuing in **this** chat, call **cursor-app-control** `move_agent_to_root` with the absolute path from `create` output. Re-read `AGENTS.md` after the move.

For `~/.cursor/cursorfs-clone/...` targets, use `move_agent_to_cloned_root` instead.

## List / remove

```bash
./scripts/agent-worktree.sh list
./scripts/agent-worktree.sh remove <name>
```

Return: worktree path, branch name, handoff block, and whether the user should open a fresh chat or move this agent.

For **multiple parallel attempts** on the same task, use `/best-of-n-worktrees` instead.
