# Handoff: Dependency waves (parent orchestrator)

Use with [handoff-parallel-parent.md](./handoff-parallel-parent.md) when a task spans **ordered layers** (schemas → services → Flink → integration).

Cursor slash command: `/dependency-waves`

**Global reference:** `~/.cursor/skills/agent-worktrees/references/dependency-waves.md`

---

## Setup (once per task)

```bash
./scripts/check-dependency-waves.sh validate
./scripts/check-dependency-waves.sh plan
./scripts/check-dependency-waves.sh init --task <task-id> --base main
./scripts/check-dependency-waves.sh complete --task <task-id> --wave spec --note specs/<task-id>.md
```

---

## Per wave (repeat)

```bash
# 1. Gate
./scripts/check-dependency-waves.sh ready --task <task-id> --wave <wave-id>

# 2. Handoff block for children
./scripts/check-dependency-waves.sh handoff --task <task-id> --wave <wave-id>

# 3. Spawn worktrees (example)
./scripts/agent-worktree.sh create --track backend --name <task-id>-state

# 4. After child returns + parent merges
./scripts/check-dependency-waves.sh complete --task <task-id> --wave <wave-id> --branch agent/backend/...

# Optional wave skipped
./scripts/check-dependency-waves.sh skip --task <task-id> --wave frontend --note "no UI change"
```

---

## Digital Twin wave order

| Wave | Who | Notes |
|------|-----|-------|
| `spec` | Parent | Written spec only |
| `contracts` | Child (`docs` track) | `check-kafka-contracts.sh` before return |
| `backend-services` | Child (parallel, ≤3) | Disjoint `services/*` paths |
| `flink` | Child (`backend`) | After contracts + services; `mvn test` |
| `frontend` | Child (optional) | After backend-services |
| `integration` | Parent | merge + `verify-worktree-merge --rebuild --with-smoke` |

Check status anytime:

```bash
./scripts/check-dependency-waves.sh status --task <task-id>
```

---

## Paste template (parent)

```
Dependency-wave parent — Digital Twin. Task: <task-id>. Main root: <path>.

## State
./scripts/check-dependency-waves.sh status --task <task-id>

## Current wave
<wave-id> — ready: yes/no

## Next
- [ ] handoff / spawn children OR parent integration wave
- [ ] complete wave after merge
- [ ] do not run smoke until integration wave on main root
```

---

## Anti-patterns

- Starting `backend-services` before `contracts` complete
- Child running `./scripts/smoke-test*.sh` (blocked; false green)
- Forgetting `complete` after merge — blocks downstream `ready`
