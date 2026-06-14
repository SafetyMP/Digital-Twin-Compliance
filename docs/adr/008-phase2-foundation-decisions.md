# ADR-008: Phase 2 Foundation Decisions (D10–D13)

**Status**: Accepted  
**Date**: 2026-06-13  
**Deciders**: Platform Architecture Team  
**Implements**: [phase2-implementation-spec.md](../phase2-implementation-spec.md)

## Context

Phase 2 adds real-time compliance monitoring (Flink CEP), Redis features, Alert Service, WebSocket delivery, and a Next.js alert console. Several choices affect local dev, CI, and how closely Phase 2 matches the long-term architecture in [ADR-001](./001-kafka-flink-streaming.md).

| ID | Decision | Phase 2 impact |
|----|----------|----------------|
| D10 | Flink runtime (dev vs staging) | Compose vs Kubernetes Operator |
| D11 | Alert durability store | PostgreSQL vs immudb |
| D12 | Threshold evaluation location | Flink inline vs Decision Service |
| D13 | Payment event source | CDC table vs synthetic Kafka producer |

## Decision

### D10 — Flink mini-cluster in Compose for dev/CI; Operator for shared staging

**Decision**: Run Flink JobManager + TaskManager in **Docker Compose** for local development and GitHub Actions. Use **Flink Kubernetes Operator** only for shared staging/production (documented in deployment guide; not required for Phase 2 exit criteria).

**Rationale**: Consistent with ADR-007 D4 (local-first). Avoids K8s dependency for every contributor while preserving ADR-001 production target.

### D11 — PostgreSQL for alert persistence; immudb in Phase 3

**Decision**: Alert Service stores `ComplianceAlert` rows in **PostgreSQL**. Do **not** write to immudb in Phase 2. `evidenceRef` on alert events may be null.

**Rationale**: Phase 3 introduces Audit Service + immudb for tamper-evident records. Dual-write in Phase 2 adds failure modes without audit UI.

### D12 — Inline CEP thresholds in Flink; Zen deferred

**Decision**: Implement `INT-M001`, `INT-M002`, and `BASEL-M001` thresholds as **Flink job configuration** (env vars). Do not call Decision Service / GoRules Zen on the hot path in Phase 2.

**Rationale**: Pattern tuning with compliance officers does not require Zen until rule versioning and CI fixtures land in Phase 3. Reduces latency and service count.

### D13 — Payments ingested via Debezium CDC

**Decision**: Add `payments` table to mock core banking; extend Debezium to publish `domain.events.public.payments`. Flink consumes CDC events for velocity detection.

**Rationale**: Reuses Phase 1 ingestion pattern; avoids a one-off Kafka producer that bypasses schema and ordering guarantees.

## Consequences

### Positive

- Phase 2 can be developed and CI-verified with `docker compose up` only.
- Clear upgrade path: swap Flink deployment to Operator without changing job JAR contract.
- Alert API and UI can ship before audit ledger complexity.

### Negative

- Compose Flink differs from Operator-managed production (savepoint deployment, HA).
- Inline thresholds duplicate logic that Zen will own in Phase 3 — plan migration to `ruleCode` + Decision Service lookup.
- Alerts are not tamper-evident until Phase 3 immudb integration.

## Alternatives Considered

| Decision | Alternative | Why rejected for Phase 2 |
|----------|-------------|--------------------------|
| D10 | K8s Operator only | Blocks local dev and CI without a cluster |
| D11 | immudb from day one | Requires Audit Service and verification UI not in Phase 2 scope |
| D12 | Zen for all thresholds | Adds service dependency and latency before patterns are validated |
| D13 | Script publishes to Kafka directly | Skips CDC; weaker parity with production ingestion |

## References

- [ADR-001: Kafka + Flink](./001-kafka-flink-streaming.md)
- [ADR-007: Phase 1 Foundation](./007-phase1-foundation-decisions.md)
- [phase2-implementation-spec.md](../phase2-implementation-spec.md)
