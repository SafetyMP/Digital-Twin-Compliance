# ADR-009: Phase 3 Foundation Decisions (D15–D20)

**Status**: Accepted  
**Date**: 2026-06-15  
**Deciders**: Platform Architecture Team  
**Implements**: [phase3-implementation-spec.md](../phase3-implementation-spec.md)

## Context

Phase 3 adds policy evaluation (Cedar + Zen), a tamper-evident audit ledger (immudb), and an Audit Explorer UI. Phase 2 already publishes `ComplianceAlertRaised` events and persists alerts in PostgreSQL with nullable `evidenceRef` ([ADR-008](./008-phase2-foundation-decisions.md) D11).

| ID | Decision | Phase 3 impact |
|----|----------|----------------|
| D15 | immudb runtime (dev vs staging) | Compose vs Kubernetes |
| D16 | Audit write path | Kafka buffer vs direct immudb |
| D17 | Engine deployment shape | Standalone Go services vs embedded only |
| D18 | Flink hot-path migration | Gradual; API/audit first |
| D19 | Evidence artifact store | Local stub vs S3 Object Lock |
| D20 | Authentication | Mock principal vs Keycloak |

## Decision

### D15 — immudb in Compose for dev/CI; clustered immudb for shared staging

**Decision**: Run **immudb** as a single-node Compose service for local development and GitHub Actions. Use **immudb HA / Kubernetes** only for shared staging/production (documented; not required for Phase 3 exit criteria).

**Rationale**: Matches ADR-007 D4 and ADR-008 D10 (local-first). Contributors need `docker compose up` without K8s.

### D16 — Audit writes via `compliance.audit.pending` Kafka topic

**Decision**: Compliance components publish canonical audit payloads to **`compliance.audit.pending`**. **Audit Service** is the sole immudb writer. On immudb failure, Audit Service retries with backoff and may republish to a DLQ after max attempts; producers do not write immudb directly.

**Rationale**: Aligns with [data-flow.md](../data-flow.md) topic registry and ADR-003 buffer-on-failure pattern. Single writer preserves hash-chain ordering.

### D17 — Cedar and Zen as standalone Go services

**Decision**:

- **Cedar Policy Service** (`services/cedar-service`) — HTTP/gRPC evaluate API; policies as `.cedar` files in Git.
- **Decision Service** (`services/decision-service`) — HTTP evaluate API; Zen JDM files in Git with JSON fixtures in CI.

Both return a shared **`RuleDecision`** JSON shape and publish audit intents to `compliance.audit.pending`.

**Rationale**: ADR-002 two-tier separation; independent CI gates (Cedar Analyzer vs Zen fixtures); Flink and Alert Service call via HTTP in later sub-phases.

### D18 — Phase 3a: audit + API evaluation; Phase 3b: optional Flink Zen lookup

**Decision**:

- **Phase 3a (exit criteria)**: Record **all Phase 2 alerts** and **API-triggered rule evaluations** in immudb; populate `evidenceRef` on alerts; Audit Explorer search + chain verification.
- **Phase 3b (stretch)**: Replace one Flink inline threshold (e.g. `BASEL-M001`) with Decision Service call — not required for Phase 3 merge bar.

**Rationale**: Reduces simultaneous changes to Flink hot path and audit ledger. ADR-008 D12 inline thresholds remain until Zen models are CI-stable.

### D19 — Evidence artifacts: filesystem stub in dev; S3 Object Lock in staging

**Decision**: Phase 3 dev/CI stores report-sized artifacts under a **local volume** (`/data/audit-artifacts`) with the same metadata schema as ADR-003. **S3 Object Lock** is documented for staging only.

**Rationale**: Avoids AWS dependency for local smoke tests while preserving the `artifactRef` contract.

### D20 — Mock principal for Cedar; Keycloak deferred to Phase 3+

**Decision**: Dev and CI use **`X-Principal`** / **`X-Roles`** headers (or env-default operator) for Cedar evaluation. **Keycloak** OIDC middleware is out of scope until a dedicated auth phase.

**Rationale**: Roadmap lists Keycloak with full auth middleware as later work; Cedar policies still need a principal model for CI fixtures.

## Consequences

### Positive

- Phase 3 can ship audit trail + policy CI without Flink refactor.
- Clear upgrade path: Compose immudb → HA cluster; filesystem artifacts → S3 Object Lock.
- `evidenceRef` on alerts becomes non-null for all new alerts.

### Negative

- Flink still uses inline thresholds — duplicate logic until Phase 3b/4 migration.
- Mock auth is not production-safe; staging must add OIDC before external exposure.
- Single-node immudb is not HA; staging checklist must call this out.

## Alternatives Considered

| Decision | Alternative | Why rejected for Phase 3 |
|----------|-------------|--------------------------|
| D16 | Direct immudb writes from each service | Breaks hash-chain ordering; harder retry semantics |
| D17 | Embed Cedar/Zen only inside Audit Service | Couples policy deploy to audit deploy; weaker CI boundaries |
| D18 | Migrate all Flink rules to Zen in Phase 3 | Too many moving parts with ledger bring-up |
| D19 | MinIO with Object Lock emulation | Extra Compose service; filesystem stub sufficient for exit criteria |
| D20 | Keycloak in Compose from day one | Expands scope; blocks policy/audit focus |

## References

- [ADR-002: Cedar + Zen](./002-cedar-decision-engine.md)
- [ADR-003: immudb](./003-immudb-audit-ledger.md)
- [ADR-008: Phase 2 Foundation](./008-phase2-foundation-decisions.md)
- [phase3-implementation-spec.md](../phase3-implementation-spec.md)
