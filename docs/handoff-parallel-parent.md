# Handoff: Parallel parent agent

Use when **you orchestrate** child agents (worktrees, best-of-N, or in-session subagents). You stay on the **main repo root**.

**Companion**: [handoff-worktree-agent.md](./handoff-worktree-agent.md) · [handoff-parallel-agent.md](./handoff-parallel-agent.md) · [handoff-verification-agent.md](./handoff-verification-agent.md)

Cursor slash command: `/parallel-parent`

---

## Before spawning children

0. Run `./scripts/check-subagent-preflight.sh` (parent root only — advisory, not CI).

1. Write or point to a **single spec** (`specs/<task>.md`, bon `task.md`, or handoff block).
2. Confirm **non-overlapping** file boundaries per track.
3. Create worktrees if needed:

```bash
# Single track
./scripts/agent-worktree.sh create --track backend --name <name>

# Best-of-N (same task, distinct approaches)
./scripts/agent-worktree-best-of-n.sh create --n 3 --prefix <id> --track experiment --task "..."
./scripts/agent-worktree-best-of-n.sh handoffs --prefix <id>
```

4. Dispatch fresh chats, multi-agent UI, or `best-of-n-runner` subagents — one per track/attempt.

For **multi-layer tasks** (schemas → services → Flink), use dependency waves:

```bash
./scripts/check-dependency-waves.sh init --task <id>
./scripts/check-dependency-waves.sh plan
./scripts/check-dependency-waves.sh ready --task <id> --wave backend-services
./scripts/check-dependency-waves.sh handoff --task <id> --wave backend-services
```

See [handoff-dependency-waves.md](./handoff-dependency-waves.md) · `/dependency-waves`

---

## While children run

- Do **not** edit the same paths from main root unless coordinating a conflict fix.
- Do **not** start docker compose for integration debug in parallel with children.
- Checkpoint when a child returns: confirm branch exists, tests claimed, files match scope.

---

## When children return

### Best-of-N batch

```bash
./scripts/agent-worktree-best-of-n.sh status --prefix <id>
./scripts/agent-worktree-best-of-n.sh compare --prefix <id>
```

Pick winner by: tests pass, smallest diff, clearest design, acceptance criteria in spec.

### Merge

```bash
git checkout <integration-branch>
git merge agent/<track>/<name>   # or cherry-pick
# resolve conflicts
```

### Post-merge verify (main root only — required)

```bash
./scripts/check-worktree-scope.sh --branch agent/<track>/<name> --strict
./scripts/verify-worktree-merge.sh agent/<track>/<name>
```

When the merge touches Compose-backed services, **rebuild before smoke**:

```bash
./scripts/verify-worktree-merge.sh agent/<track>/<name> --rebuild
./scripts/verify-worktree-merge.sh agent/<track>/<name> --rebuild --with-smoke
```

`--with-smoke` is refused without `--rebuild` when changed paths map to compose services (stale container risk). Review printed rebuild hints if you prefer manual `docker compose build/up` lines instead of `--rebuild`.

Or run smoke steps manually after rebuild:

```bash
docker compose -f docker-compose.dev.yml up -d --wait   # if stack needed
./scripts/seed.sh                                          # if cold
./scripts/smoke-test.sh
./scripts/smoke-test-phase2.sh                             # when Phase 2 applies
./scripts/smoke-test-phase3.sh                             # when Phase 3 applies
```

**Do not** run smoke or `./scripts/verify-worktree-merge.sh` from a worktree — they target the main Docker stack; worktree runs produce **false greens** (blocked by guard-shell).

### Cleanup

```bash
./scripts/agent-worktree-best-of-n.sh remove --prefix <id>   # or --force
# or per worktree:
./scripts/agent-worktree.sh remove <name>
./scripts/check-agent-worktrees.sh
```

---

## Paste template (parent session)

```
Parallel parent — Digital Twin. Main root: <absolute path>. Do NOT read agent-transcripts.

## Batch / tracks
- best-of-N prefix: <id> OR worktree names: <list>

## Child status
- attempt 1: branch … tests … 
- attempt 2: …

## Next (parent)
- [ ] compare / pick winner
- [ ] merge to <branch>
- [ ] verify-worktree-merge.sh on main root (+ rebuild hints)
- [ ] --with-smoke after rebuild
- [ ] remove worktrees

Do not trust child smoke/compose claims. Parent owns integration verify.
```

---

## Anti-patterns

- Merging without `compare` on best-of-N
- Claiming Phase 2/3 done after child package tests only
- Leaving stale `.worktrees/` directories after merge
- Running `docker compose` or `./scripts/smoke-test*.sh` from a worktree (blocked by guard-shell; smoke would hit the main stack anyway)
