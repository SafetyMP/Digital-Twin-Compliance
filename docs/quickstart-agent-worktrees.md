# Quickstart: Agent worktrees and dependency waves

Minimal path to use parallel agent tooling on this repo after PR #27.

## Prerequisites

- Global harness: `~/.cursor/scripts/agent-worktree/config.py`, `~/.cursor/skills/agent-worktrees/`, `22-parallel-agents.mdc`
- On `main` with worktree scripts present

## 1. Dry run (no worktrees)

```bash
./scripts/demo-agent-workflows.sh
./scripts/check-agent-worktrees.sh
```

## 2. Single isolated track

Parent on **main repo root**:

```bash
./scripts/agent-worktree.sh create --track backend --name my-fix
./scripts/agent-worktree.sh handoff my-fix
```

Open a fresh chat at the printed path (or `move_agent_to_root`). Child runs **package tests only** — not smoke.

Parent after child returns:

```bash
git merge agent/backend/my-fix
./scripts/check-worktree-scope.sh --branch agent/backend/my-fix --strict
./scripts/verify-worktree-merge.sh agent/backend/my-fix --rebuild --with-smoke
./scripts/agent-worktree.sh remove my-fix
```

See [handoff-worktree-agent.md](./handoff-worktree-agent.md) · [handoff-parallel-parent.md](./handoff-parallel-parent.md).

## 3. Multi-layer task (dependency waves)

For changes touching schemas → services → Flink:

```bash
./scripts/check-dependency-waves.sh init --task my-feature
./scripts/check-dependency-waves.sh complete --task my-feature --wave spec --note specs/my-feature.md

# Each child wave:
./scripts/check-dependency-waves.sh ready --task my-feature --wave contracts
./scripts/check-dependency-waves.sh handoff --task my-feature --wave contracts
# spawn worktree → merge → complete wave

./scripts/check-dependency-waves.sh status --task my-feature
```

Wave order (default): `spec` → `contracts` → `backend-services` → `flink` → `frontend` (optional) → `integration`.

Integration wave is **parent only** — `verify-worktree-merge` + smoke on main root.

See [handoff-dependency-waves.md](./handoff-dependency-waves.md) · slash `/dependency-waves`.

## 4. Best-of-N (same task, different approaches)

```bash
./scripts/agent-worktree-best-of-n.sh create --n 3 --prefix my-id --track experiment --task "..."
./scripts/agent-worktree-best-of-n.sh handoffs --prefix my-id
# after attempts:
./scripts/agent-worktree-best-of-n.sh compare --prefix my-id
```

Parent merges winner; see [handoff-parallel-parent.md](./handoff-parallel-parent.md).

## 5. Port to another repo

Copy from `~/.cursor/templates/worktrees/`:

- `worktrees.config.json.example` → `.cursor/worktrees.config.json`
- `check-dependency-waves.sh.example` → `scripts/check-dependency-waves.sh`

Guide: `~/.cursor/skills/agent-worktrees/references/repo-setup.md`

## Anti-patterns

- Running `./scripts/smoke-test*.sh` from a `.worktrees/` cwd (guard-shell blocks; false green)
- Spawning backend children before `contracts` wave completes
- Skipping parent `verify-worktree-merge` after merge

## Related

- [handoff-worktree-agent.md](./handoff-worktree-agent.md)
- [handoff-parallel-parent.md](./handoff-parallel-parent.md)
- [handoff-dependency-waves.md](./handoff-dependency-waves.md)
