# ADR-010: Phase 4 Foundation Decisions (D21–D24)

**Status**: Accepted (planning)  
**Date**: 2026-06-29  
**Deciders**: Platform Architecture Team  
**Implements**: [phase4-implementation-spec.md](../phase4-implementation-spec.md)

## Context

Phase 4 adds an exposure graph (Neo4j) and deterministic stress simulation (Python) on top of the Phase 3 rules + audit stack. Phase 2 already computes counterparty exposure features in Redis/Flink; Phase 3 records alert and rule outcomes in immudb.

| ID | Decision | Phase 4 impact |
|----|----------|----------------|
| D21 | Graph database | Neo4j vs Memgraph |
| D22 | Graph ingestion source | Domain events vs batch ETL |
| D23 | Simulation scope | Deterministic vs agent-based |
| D24 | Simulation ↔ audit linkage | Kafka audit pending vs direct HTTP |

Roadmap open items D5 (simulation scope) and D7 (graph DB) are resolved here for Phase 4 implementation.

## Decision

### D21 — Neo4j Community Edition in Compose for dev/CI

**Decision**: Use **Neo4j 5.x Community** in `docker-compose.dev.yml` for local dev and CI. Evaluate **Memgraph** or **Neo4j Aura** only if Cypher query p99 on seed graph exceeds budget during Phase 4 exit benchmarking.

**Rationale**: ADR-006 polyglot strategy and roadmap D7; mature GDS library for centrality metrics later.

### D22 — Graph ingestion from Kafka domain events (incremental)

**Decision**: **Graph Service** consumes **`twin.state.updated`** and **`domain.events.public.*`** (instrument / legal-entity changes) to upsert nodes and `Exposure` edges. No separate batch ETL for exit criteria — seed graph must be reachable from existing Phase 2 CDC path within smoke-test window.

**Rationale**: Aligns with event-driven architecture (ADR-001); reuses state-service twin mirror as source of truth.

### D23 — Deterministic stress scenario only for Phase 4 exit criteria

**Decision**: Ship **one ECB Adverse–style deterministic scenario** (NetworkX or internal graph walk) completing in **< 60s** on seed data. **Agent-based contagion** is Phase 6+ stretch (roadmap D5).

**Rationale**: Reduces simultaneous risk (new DB + new ML/agent runtime) while delivering explainable simulation output linked to audit.

### D24 — Simulation results audit via `compliance.audit.pending`

**Decision**: Simulation Service publishes **`SimulationRunCompleted`** audit intents to **`compliance.audit.pending`** (same envelope family as Phase 3). Audit Service remains sole immudb writer.

**Rationale**: Consistent with ADR-009 D16; Audit Explorer can search simulation runs without a second ledger.

## Consequences

### Positive

- Phase 4 builds on existing Kafka + audit patterns.
- Graph UI and simulation UI can reuse Next.js proxy conventions from alert-console / audit-explorer.
- Zen **COREP-R001/R002** can evaluate simulation-derived capital metrics without new engine work.

### Negative

- Neo4j adds memory/Compose footprint; CI job time may increase.
- Python simulation service introduces gRPC/HTTP contract maintenance (ADR-006).
- Incremental graph ingestion may lag twin mirror — smoke tests must wait on graph node counts like Phase 2 waits on Redis keys.

## References

- [roadmap.md](../roadmap.md) — Phase 4 scope
- [ADR-006](./006-polyglot-language-strategy.md) — Go + Python split
- [ADR-009](./009-phase3-foundation-decisions.md) — audit write path
