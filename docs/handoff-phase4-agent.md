# Handoff: Phase 4 Implementation Agent

Use this document when starting Phase 4 implementation in a **new** Cursor chat.

**Created**: 2026-06-29  
**Planning deliverables**: ADR-010, Phase 4 spec, readiness checklist, AGENTS.md phase pointer.

---

## Your mission

Implement **Phase 4 only** — exposure graph and deterministic stress simulation:

> Graph Service (Go + Neo4j), Graph Explorer UI, Simulation Service (Python), Simulation Console UI, smoke-test-phase4.sh, CI extension.

Read and follow:

1. [docs/phase4-implementation-spec.md](./phase4-implementation-spec.md) — **primary executable spec**
2. [AGENTS.md](../AGENTS.md) — commands, layout, definition of done
3. [docs/adr/010-phase4-foundation-decisions.md](./adr/010-phase4-foundation-decisions.md) — D21–D24
4. [docs/review/phase4-readiness.md](./review/phase4-readiness.md) — confirm prerequisites before coding

---

## Decisions already made (do not re-open)

| ID | Decision | Implementation hint |
|----|----------|---------------------|
| D21 | Neo4j Community in Compose | Single node; not Aura for exit criteria |
| D22 | Ingest from Kafka twin/domain events | No batch-only ETL for smoke path |
| D23 | Deterministic ECB Adverse scenario only | Agent-based contagion is Phase 6+ |
| D24 | Simulation audit via `compliance.audit.pending` | Audit Service sole immudb writer |

Phase 3 decisions (ADR-009) and Phase 2 Kafka envelope patterns still apply.

---

## In scope

| Deliverable | Spec section |
|-------------|--------------|
| Neo4j + Compose | §4 |
| Graph Service | §6.1 |
| Simulation Service | §6.2 |
| Graph Explorer + Simulation Console | §6.3 |
| `smoke-test-phase4.sh`, `wait-graph-seeded.sh` | §7 |
| CI extend | §8 |

---

## Explicitly out of scope

- XBRL/SDMX regulatory reporting (Phase 5)
- Agent-based contagion models
- Keycloak / OIDC middleware
- ClickHouse analytics store
- Neo4j HA / Aura (document only)

---

## Suggested prompt for implementation agent

```
Context budget: read AGENTS.md + phase4-implementation-spec.md only.
For touched services, read services/<svc>/AGENTS.md once created.
Load docs/data-flow.md for Kafka envelope / audit tasks.

Implement Phase 4 per docs/phase4-implementation-spec.md and ADR-010.
Do NOT implement Phase 5+ (XBRL, regulatory reporting pipeline).

Order: neo4j Compose → graph-service consumer → wait-graph-seeded.sh →
simulation-service → graph-explorer → simulation-console →
smoke-test-phase4.sh → CI.

Mark done only when smoke-test-phase4.sh, smoke-test-phase3.sh (regression),
go test (graph-service), and pytest (simulation-service) pass.
```

---

## Verification handoff

When implementation claims done, start a **fresh** verification chat with [handoff-verification-agent.md](./handoff-verification-agent.md) adapted for Phase 4:

```bash
./scripts/smoke-test.sh
./scripts/smoke-test-phase2.sh
./scripts/smoke-test-phase3.sh
./scripts/smoke-test-phase4.sh
./scripts/verify-audit-chain.sh
```

---

## Parallel tracks (optional)

Only if file boundaries are explicit:

| Track | Paths | Notes |
|-------|-------|-------|
| Backend graph | `services/graph-service/`, Neo4j compose | Parent runs Kafka smoke |
| Backend simulation | `services/simulation-service/` | Depends on graph REST |
| Frontend | `apps/graph-explorer/`, `apps/simulation-console/` | Proxy pattern only |

Run `./scripts/check-subagent-preflight.sh` before dispatching subagents ([handoff-parallel-parent.md](./handoff-parallel-parent.md)).
