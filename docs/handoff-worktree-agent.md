# Handoff: Git worktree agent session

Use when parallel work needs **filesystem isolation** (separate branch + directory) instead of only a separate chat on the same root.

**Companion**: [handoff-parallel-agent.md](./handoff-parallel-agent.md) (separate-chat, same root) · [handoff-continuation.md](./handoff-continuation.md) · [AGENTS.md](../AGENTS.md) § Parallel work

---

## When to use worktrees vs separate chats

| Pattern | Use when |
|---------|----------|
| **Separate chat** (same repo root) | Default; planning → implementation → verification |
| **In-session subagents** | Max 3 tracks, parent synthesizes, short-lived |
| **Git worktree** (`./scripts/agent-worktree.sh`) | Independent file boundaries, long-running track, or Cursor multi-agent / best-of-N isolation |

**Do not use worktrees** for integration debugging, shared `docker-compose.dev.yml` edits, or anything that needs `./scripts/smoke-test-phase2.sh` from the agent — parent owns Compose and smoke from the main checkout.

---

## Quick start

From the **main** repo root:

```bash
# Backend-only track (state-service example)
./scripts/agent-worktree.sh create --track backend --name state-outbox

# Print paste-ready prompt for a fresh chat at the worktree
./scripts/agent-worktree.sh handoff state-outbox

# List / clean up
./scripts/agent-worktree.sh list
./scripts/agent-worktree.sh remove state-outbox
```

Worktrees live under `.worktrees/` (gitignored). Each create also makes branch `agent/<track>/<name>`.

---

## Best-of-N (parallel attempts)

When you need **multiple independent solutions** to the same task (not just parallel file tracks):

```bash
./scripts/agent-worktree-best-of-n.sh create --n 3 --prefix audit-refactor --track backend \
  --task "Refactor audit consumer retries; keep go test green."
./scripts/agent-worktree-best-of-n.sh handoffs --prefix audit-refactor
# ... run N agents (fresh chats, multi-agent UI, or best-of-n-runner subagents) ...
./scripts/agent-worktree-best-of-n.sh compare --prefix audit-refactor
./scripts/agent-worktree-best-of-n.sh remove --prefix audit-refactor
```

Manifest: `.worktrees/bon-<prefix>/manifest.json`. Each attempt gets `.agent-bon-task.md` in its worktree.

Cursor slash command: `/best-of-n-worktrees`. Global skill: `agent-worktrees` in `~/.cursor/skills/`.

---

## Cursor integration

1. Run `create`, note the absolute **path** in the output.
2. **Fresh chat** with that folder as workspace root, **or** call `cursor-app-control` **`move_agent_to_root`** with `rootPath` set to that path.
3. Paste output of `./scripts/agent-worktree.sh handoff <name>`.
4. When done: push branch or open PR; parent merges after unit tests; parent runs smoke from main root.
5. `remove` when the branch is merged or abandoned.

For clones under `~/.cursor/cursorfs-clone/...`, use **`move_agent_to_cloned_root`** instead of `move_agent_to_root` (local-only branches).

---

## Track scopes (non-overlapping)

| Track | May edit | Must not edit |
|-------|----------|----------------|
| `backend` | `services/*`, `jobs/*` | `apps/*`, `docker-compose.dev.yml`, smoke scripts |
| `frontend` | `apps/*` | `services/*`, Compose, smoke scripts |
| `docs` | `docs/`, `policies/`, eval docs | Service code unless explicit |
| `experiment` | Confirm with parent first | Compose without coordination |

---

## Suggested prompt (after handoff block)

```
Implement <task> in this worktree only.

Read AGENTS.md + scoped services/*/AGENTS.md. Do not load architecture docs unless required.
Do not run docker compose or smoke tests — parent owns integration verify.
Mark done only when package tests for touched paths exit 0.
Return: branch, files changed, test commands + exit codes.
```

---

## Parent checklist (main root)

After worktree agent returns:

1. `./scripts/check-worktree-scope.sh --branch agent/<track>/<name> --strict`
2. Review diff on `agent/<track>/<name>`
3. Merge or cherry-pick into your integration branch
4. `./scripts/verify-worktree-merge.sh agent/<track>/<name>` → run rebuild hints
5. `./scripts/verify-worktree-merge.sh agent/<track>/<name> --with-smoke` (after rebuild)
6. `./scripts/agent-worktree.sh remove <name>` when merged

Do **not** run step 4–5 from the worktree directory — guard-shell blocks integration scripts there.

Parent orchestration: [handoff-parallel-parent.md](./handoff-parallel-parent.md) · `/parallel-parent`
