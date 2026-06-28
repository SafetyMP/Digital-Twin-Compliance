# Handoff: Multi-session continuation

Use when stopping mid-task or starting a **new** chat to continue work. Prefer this over reading prior `agent-transcripts/`.

**Companion**: [handoff-verification-agent.md](./handoff-verification-agent.md) · [handoff-parallel-agent.md](./handoff-parallel-agent.md) · [handoff-worktree-agent.md](./handoff-worktree-agent.md) · [handoff-parallel-parent.md](./handoff-parallel-parent.md)

---

## Paste template (copy into new chat)

```
Continue work on Digital Twin (Phase 2). Do NOT read agent-transcripts.

## Workspace
- Root: <main repo absolute path | worktree absolute path>
- Branch: <branch name, e.g. agent/backend/my-track>
- Parent batch: <bon-<prefix> attempt K/N | none>

## Done
- <bullet>

## Blocked
- <bullet or "none">

## Next
- <ordered steps>

## Gotchas hit
- <one line each; link AGENTS.md § Repo gotchas if already documented>

## Verify before claiming done
- Worktree/child: package tests only (go test / mvn test / npm test for touched paths)
- Main root / parent: ./scripts/smoke-test.sh, smoke-test-phase2.sh (integration — parent only)

Context budget: AGENTS.md + relevant service AGENTS.md + phase spec only.
```

---

## First-read contract (every new session)

Before debugging or editing:

1. [AGENTS.md](../AGENTS.md) — especially **Repo gotchas**
2. [docs/phase2-implementation-spec.md](./phase2-implementation-spec.md) when Phase 2 applies
3. Scoped [services/*/AGENTS.md](../services/) for the service you touch

Then `grep` before `Read` on large files.

---

## Capture before close-out

If the session involved a correction or non-obvious fix:

| Learning | Where |
|----------|--------|
| Repo-specific gotcha | `AGENTS.md` → Repo gotchas (one line) |
| Service pattern | `services/<svc>/AGENTS.md` |
| Cross-project preference | `~/.cursor/memory/MEMORY.md` or user rule |
| Repeatable workflow | Cursor skill |

Ask: *"Should I add this to AGENTS.md gotchas?"* before ending a painful debug session.

Do **not** capture one-off task state here — use the paste template above.

---

## Session boundaries

| From | To | Action |
|------|-----|--------|
| Implementation | Verification | New chat → [handoff-verification-agent.md](./handoff-verification-agent.md) |
| Parallel children | Parent merge/smoke | Parent chat → [handoff-parallel-parent.md](./handoff-parallel-parent.md) or `/parallel-parent` |
| Long debug | Fresh attempt | New chat + paste template (outcomes only) |
| Any session | Eval scoring | New chat; run `./scripts/score-eval-session.sh` only |

---

## References

- [AGENTS.md](../AGENTS.md) § Agent learning · § Session hygiene
- [evals/live-model-phase2/scenarios/06-debug-int-m001-retention.md](../evals/live-model-phase2/scenarios/06-debug-int-m001-retention.md) — retention behavioral scenario
