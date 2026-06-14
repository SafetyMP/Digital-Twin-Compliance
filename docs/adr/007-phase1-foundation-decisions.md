# ADR-007: Phase 1 Foundation Decisions (D1, D4, D9)

**Status**: Accepted  
**Date**: 2026-06-13  
**Deciders**: Platform Architecture Team  
**Supersedes**: Open items D1, D4, D9 in [roadmap.md](../roadmap.md)

## Context

Phase 1 implementation (ingestion backbone + minimal twin) requires three architectural choices that were left open after Phase 0. Each choice affects database schema, local dev setup, and seed data shape.

| ID | Decision | Phase 1 impact |
|----|----------|----------------|
| D1 | Deployment model | Auth, tenancy columns, API scope |
| D4 | Kafka hosting (dev) | `docker-compose` vs cloud credentials |
| D9 | Consolidation hierarchy depth | `LegalEntity` parent links, seed data |

## Decision

### D1 — Single institution first; multi-tenant-ready schema

**Decision**: Deploy as a **single-institution** platform for Phase 1–5. Design PostgreSQL schemas and APIs so multi-tenant support can be added in Phase 6 without breaking changes.

**Phase 1 implications**:

- No auth/IdP in Phase 1 (deferred to Phase 3 per roadmap).
- Add optional `tenant_id UUID` column on all entity tables with a single default tenant (`00000000-0000-0000-0000-000000000001`) used in seed data.
- REST API serves one institution's data; no tenant routing middleware yet.
- Do **not** build supervisory multi-tenant UI or cross-tenant isolation tests in Phase 1.

**Rationale**: Matches roadmap recommendation; avoids Phase 1 scope creep while preventing a costly schema migration later.

### D4 — Local KRaft Kafka for dev; managed Kafka for staging+

**Decision**: Use **self-hosted KRaft Kafka + Schema Registry in Docker Compose** for local development and CI integration tests. Use **Confluent Cloud** (or equivalent managed Kafka) for shared staging when the team needs it (Phase 2+).

**Phase 1 implications**:

- `docker-compose.dev.yml` runs: PostgreSQL (mock core banking), PostgreSQL (state store), Kafka (KRaft, 1 broker for dev), Schema Registry, Debezium Connect.
- No cloud credentials required to start Phase 1 locally.
- CI uses the same Compose stack (GitHub Actions service containers or `docker compose up -d`).
- Document optional Confluent Cloud env vars in `docs/phase1-implementation-spec.md` for teams that prefer managed dev Kafka.

**Rationale**: Local-first dev matches Phase 1 exit criteria (events flow end-to-end without external deps). Managed Kafka adds cost and credential friction for initial scaffolding. ADR-001 already specifies KRaft for production; local KRaft keeps dev/prod parity.

### D9 — Three-level consolidation hierarchy

**Decision**: Support **three levels** of group consolidation: **parent → subsidiary → sub-subsidiary**. No unlimited-depth hierarchy in Phase 1–5.

**Phase 1 implications**:

- `legal_entities.parent_entity_id UUID NULL` with check constraint: depth ≤ 3 (enforced in application layer or trigger).
- Seed data includes a small hierarchy: 3 parent groups, each with 2–3 subsidiaries, some with 1 sub-subsidiary (10 institutions total per roadmap).
- `consolidationScope` enum unchanged (`Solo`, `Group`, `Excluded`).
- Graph Service (Phase 4) can traverse up to 3 hops; no recursive CTE requirement beyond that.

**Rationale**: Covers FINREP/COREP group reporting for typical banking groups without the complexity of arbitrary-depth trees. Matches roadmap recommendation.

## Consequences

### Positive

- Phase 1 parallel agent can scaffold without waiting on auth, cloud Kafka, or deep hierarchy design.
- Local `docker compose up` is sufficient for all Phase 1 exit criteria.
- `tenant_id` column is cheap insurance for Phase 6 multi-tenant.

### Negative

- Single default tenant may be forgotten when adding real auth — document in AGENTS.md.
- One-broker KRaft dev differs from 3-broker production — document production topology separately.
- Three-level cap may require ADR revision if a jurisdiction mandates deeper consolidation trees.

## Alternatives Considered

| Decision | Alternative | Why rejected for Phase 1 |
|----------|-------------|--------------------------|
| D1 | Supervisory multi-tenant from day one | Requires auth, row-level security, and tenant isolation tests — out of Phase 1 scope |
| D4 | Confluent Cloud only for dev | Adds credential setup and network dependency before first code runs |
| D9 | Unlimited hierarchy depth | Complicates seed data, validation, and graph queries with no Phase 1 benefit |

## References

- [roadmap.md](../roadmap.md) — Phase 1 scope and open decisions
- [ADR-001](./001-kafka-flink-streaming.md) — Kafka + Flink
- [ADR-004](./004-datastore-selection.md) — PostgreSQL entity store
- [phase1-implementation-spec.md](../phase1-implementation-spec.md) — Executable handoff
