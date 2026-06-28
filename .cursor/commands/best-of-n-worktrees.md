---
description: Create N git worktrees for best-of-N parallel agent attempts with compare/handoff helpers.
---
Run a **best-of-N** batch per [docs/handoff-worktree-agent.md](../../docs/handoff-worktree-agent.md) § Best-of-N.

Load the **`agent-worktrees`** skill (`~/.cursor/skills/agent-worktrees/SKILL.md`) for global patterns.

## Create batch

Ask for `--n` (1–8), `--prefix`, `--track`, and task text or `--task-file`, then:

```bash
./scripts/agent-worktree-best-of-n.sh create --n 3 --prefix <prefix> --track experiment \
  --task "<single shared task spec>"
./scripts/agent-worktree-best-of-n.sh handoffs --prefix <prefix>
```

## Run attempts

For each printed path, either:

1. **Fresh Cursor chat** at that workspace root (or `move_agent_to_root`), paste that attempt's handoff, **or**
2. **Task subagent** `best-of-n-runner` with the same task and "attempt K of N — distinct approach".

Parent must not merge until `compare` and integration verify on main root.

## Compare and cleanup

```bash
./scripts/agent-worktree-best-of-n.sh compare --prefix <prefix>
./scripts/agent-worktree-best-of-n.sh status --prefix <prefix>
./scripts/agent-worktree-best-of-n.sh remove --prefix <prefix>
```

Return: manifest path, all attempt paths/branches, compare summary, and recommended winner selection criteria.
