# Handoff: Parallel Implementation Agent

Use this document when starting Phase 1 implementation in a separate Cursor chat or agent session.

**Created**: 2026-06-13  
**Planning agent deliverables**: ADR-007, Phase 1 spec, AGENTS.md, this handoff, review checklist.  
**Companion**: [handoff-worktree-agent.md](./handoff-worktree-agent.md) when the track needs an isolated git worktree (branch + directory).

---

## Your mission

Implement **Phase 1 only** of the Financial Digital Twin + Compliance Platform:

> Ingestion backbone and minimal twin — Kafka, Debezium CDC, Go State Service, PostgreSQL entity store, REST API, schema CI.

Read and follow:

1. [docs/phase1-implementation-spec.md](./phase1-implementation-spec.md) — **primary executable spec**
2. [AGENTS.md](../AGENTS.md) — commands, layout, definition of done
3. [docs/adr/007-phase1-foundation-decisions.md](./adr/007-phase1-foundation-decisions.md) — D1, D4, D9 decisions

---

## Decisions already made (do not re-open)

| ID | Decision | Implementation hint |
|----|----------|---------------------|
| D1 | Single institution; multi-tenant-ready schema | `tenant_id` column, default UUID on all tables |
| D4 | Local KRaft Kafka in Docker Compose for dev/CI | No Confluent Cloud creds required locally |
| D9 | 3-level consolidation hierarchy | Validate parent chain depth ≤ 3 in seed + consumer |

Other ADRs (Kafka+Flink, Cedar, immudb, Zen, datastores) apply to later phases — reference only.

---

## In scope

| Deliverable | Spec section |
|-------------|--------------|
| Repo scaffold | Spec §2 |
| `docker-compose.dev.yml` | Spec §3 |
| Mock core-banking Postgres + seed | Spec §4.1, §6 |
| State store migrations | Spec §4.2 |
| Avro schemas + registration script | Spec §5 |
| Go State Service (REST + consumer + outbox) | Spec §7 |
| `scripts/seed.sh`, `scripts/smoke-test.sh` | Spec §6, §9 |
| GitHub Actions CI + schema compat | Spec §8 |

---

## Explicitly out of scope

See [AGENTS.md § Scope by phase](../AGENTS.md#scope-by-phase) for the canonical list. Phase/deferral rationale remains in [roadmap.md](./roadmap.md).

If you need a stub interface for a future service, use a TODO comment referencing the roadmap phase — do not build the service.

---

## Suggested prompt for implementation agent

Copy into the parallel Cursor chat:

```
Context budget: read AGENTS.md + phase1-implementation-spec.md only.
For Go work, read services/state-service/AGENTS.md.
Do not load architecture/domain-model/ADRs unless this task requires them.

Implement Phase 1 of the Financial Digital Twin platform per:
- docs/phase1-implementation-spec.md
- AGENTS.md

Follow ADR-007 (single tenant, local Kafka, 3-level hierarchy).
Do NOT implement Phase 2+ components (see AGENTS.md § Out of scope).

Order: scaffold → docker-compose → migrations → seed → schemas →
state-service (store, REST, consumer, outbox) → debezium → smoke-test → CI.

Mark done only when scripts/smoke-test.sh and go test ./... pass.
```

---

## Verification before handback to planning agent

When implementation is complete, use a **separate fresh chat** and [handoff-verification-agent.md](./handoff-verification-agent.md) (or `/verify-phase2`) before handback. Do not verify in the same long implementation thread.

## Three-chat workflow

| Chat | Load | Run |
|------|------|-----|
| **Implement** (this handoff) | Phase spec + service `AGENTS.md` | Code + `go test` |
| **Verify** | `AGENTS.md` + smoke scripts | `./scripts/smoke-test.sh` (+ phase2) + `token-efficiency.sh --strict` |
| **Eval / metrics** | Scripts only | `./scripts/report-eval-scorecard.sh` |

Ask the planning agent to review using [docs/review/phase1-review-checklist.md](./review/phase1-review-checklist.md).

Provide:

- Branch name or `git diff` summary
- Output of `./scripts/smoke-test.sh`
- Output of `cd services/state-service && go test ./...`
- PR link if opened

---

## Questions → planning agent

Ask the planning agent (not the user) for clarification on:

- Event schema field additions (must stay BACKWARD compatible)
- Repo layout deviations
- Whether a feature belongs in Phase 1 vs Phase 2

Do **not** ask the planning agent to write application code — only specs, ADRs, and reviews.

---

## File index (created by planning agent)

| File | Purpose |
|------|---------|
| [docs/adr/007-phase1-foundation-decisions.md](./adr/007-phase1-foundation-decisions.md) | D1, D4, D9 ADR |
| [docs/phase1-implementation-spec.md](./phase1-implementation-spec.md) | Full implementation spec |
| [AGENTS.md](../AGENTS.md) | Repo contract |
| [docs/handoff-parallel-agent.md](./handoff-parallel-agent.md) | This document |
| [docs/handoff-verification-agent.md](./handoff-verification-agent.md) | Phase 2 verification handoff |
| [docs/review/phase1-review-checklist.md](./review/phase1-review-checklist.md) | Post-implementation review |
