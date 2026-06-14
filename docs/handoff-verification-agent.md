# Handoff: Verification Agent (Phase 2)

Use this document when verifying Phase 2 integration in a **fresh** Cursor chat — especially after analysis or implementation in another session.

**Created**: 2026-06-13  
**Companion**: [handoff-parallel-agent.md](./handoff-parallel-agent.md) (implementation handoff)

---

## Your mission

Confirm Phase 2 integration is green with **minimal diffs**:

> `./scripts/smoke-test.sh` AND `./scripts/smoke-test-phase2.sh` must exit 0.

Read and follow:

1. [AGENTS.md](../AGENTS.md) — commands, session hygiene, definition of done
2. [services/alert-service/AGENTS.md](../services/alert-service/AGENTS.md) — if alert pipeline fails

Do **not** read prior `agent-transcripts/`, `evals/live-model/README.md`, or superpowers skills.

---

## Suggested prompt for verification agent

Copy into a **new** Cursor chat:

```
Context budget: AGENTS.md + services/alert-service/AGENTS.md only.
Do NOT read agent-transcripts, evals README, or prior chat history.

Verify Phase 2 per docs/handoff-verification-agent.md:

1. docker compose -f docker-compose.dev.yml up -d --wait
2. ./scripts/seed.sh
3. Ensure Flink job RUNNING (http://localhost:8082/jobs); resubmit via ./scripts/submit-flink-job.sh if needed
4. ./scripts/smoke-test.sh
5. ./scripts/smoke-test-phase2.sh
6. ./scripts/token-efficiency.sh --strict
7. Fix failures with minimal focused diffs; re-run until steps 4–6 exit 0

Return: what failed, what you fixed, command output evidence (exit codes), and efficiency metrics (harness_reread_count: 0, duplicate_read_count ≤ 3).
```

---

## Ordered verification steps

| Step | Command / check |
|------|-----------------|
| Stack | `docker compose -f docker-compose.dev.yml up -d --wait` |
| Seed | `./scripts/seed.sh` |
| Flink | Job status `RUNNING` on `:8082`; `./scripts/submit-flink-job.sh` if missing |
| Phase 1 regression | `./scripts/smoke-test.sh` |
| Phase 2 E2E | `./scripts/smoke-test-phase2.sh` |
| Unit tests (if code changed) | `cd services/alert-service && go test ./...` |
| Context efficiency (required) | `./scripts/token-efficiency.sh --strict` — `harness_reread_count: 0`, `duplicate_read_count ≤ 3` |

---

## Three-chat workflow

| Chat | Load | Run |
|------|------|-----|
| Implement | Phase spec + service `AGENTS.md` | Code + `go test` |
| **Verify** (this handoff) | `AGENTS.md` + smoke scripts | Smoke tests + `token-efficiency.sh --strict` |
| Eval / metrics | Scripts only | `./scripts/report-eval-scorecard.sh` |

---

## Common failure areas

- Flink job not `RUNNING` (missing `domain.events.public.payments` topic, Jackson classpath, job not submitted)
- `INT-M001` not firing (Debezium payments CDC, Flink payment parser, alert-service consumer lag)
- `BASEL-M001` not firing (state-service liquidity enrichment, twin.state.updated, consumer persistence)
- Alert service empty API but Kafka has messages (stale consumer offsets — restart `alert-service`)

---

## Return to user

Provide:

- Output of both smoke tests (exit 0 evidence)
- `./scripts/token-efficiency.sh --strict` output (`harness_reread_count: 0`, `duplicate_read_count ≤ 3`)
- List of files changed (if any)
- Anything still blocked

Do **not** claim done without fresh command output.

---

## File index

| File | Purpose |
|------|---------|
| [docs/handoff-verification-agent.md](./handoff-verification-agent.md) | This document |
| [docs/phase2-implementation-spec.md](./phase2-implementation-spec.md) | Phase 2 spec §14 exit criteria |
| [AGENTS.md](../AGENTS.md) | Repo contract |
