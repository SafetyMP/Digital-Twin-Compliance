# Phase 1 Review Checklist

Use this checklist when reviewing implementation agent output against [roadmap.md](../roadmap.md) Phase 1 and [phase1-implementation-spec.md](../phase1-implementation-spec.md).

**Reviewer**: Planning agent or human  
**When**: Before merge / after parallel agent claims Phase 1 complete

---

## 1. Scope compliance

| Check | Pass | Notes |
|-------|------|-------|
| No Flink, Cedar, immudb, Neo4j, UI, or auth code added | ☐ | |
| Only Phase 1 paths touched (`services/state-service`, `mocks/`, `schemas/`, `scripts/`, Compose, CI) | ☐ | |
| No unrelated refactors or doc rewrites outside Phase 1 | ☐ | |

---

## 2. Architecture alignment (ADRs)

| Check | Pass | Notes |
|-------|------|-------|
| Kafka as event backbone ([ADR-001](../adr/001-kafka-flink-streaming.md)) | ☐ | |
| PostgreSQL entity store ([ADR-004](../adr/004-datastore-selection.md)) | ☐ | |
| Go for State Service ([ADR-006](../adr/006-polyglot-language-strategy.md)) | ☐ | |
| Single tenant + `tenant_id` column ([ADR-007](../adr/007-phase1-foundation-decisions.md) D1) | ☐ | |
| Local KRaft in Compose for dev ([ADR-007](../adr/007-phase1-foundation-decisions.md) D4) | ☐ | |
| Max 3-level entity hierarchy ([ADR-007](../adr/007-phase1-foundation-decisions.md) D9) | ☐ | |

---

## 3. Data contract ([data-flow.md](../data-flow.md))

| Check | Pass | Notes |
|-------|------|-------|
| Event envelope fields: eventId, eventType, eventVersion, source, correlationId, timestamp, idempotencyKey, payload | ☐ | |
| Idempotency via `processed_events` / idempotencyKey dedup | ☐ | |
| Outbox pattern for `twin.state.updated` (not fire-and-forget Kafka) | ☐ | |
| Avro schemas in `schemas/avro/` with BACKWARD compatibility | ☐ | |
| Topics: `domain.events`, `twin.state.updated` at minimum | ☐ | |

---

## 4. API contract (spec §7)

| Check | Pass | Notes |
|-------|------|-------|
| `GET /api/v1/health` returns 200 | ☐ | |
| `GET /api/v1/personas/{id}` returns persona or 404 | ☐ | |
| `GET /api/v1/personas?personaType=` filters correctly | ☐ | |
| Response JSON matches TwinPersona shape | ☐ | |

---

## 5. Seed data (roadmap Phase 1)

| Check | Pass | Notes |
|-------|------|-------|
| 10 institutions (legal entities) | ☐ | |
| 100 accounts | ☐ | |
| 500 instruments | ☐ | |
| Hierarchy demonstrates 3 consolidation levels | ☐ | |

---

## 6. Exit criteria (roadmap + spec §10)

| Check | Pass | Notes |
|-------|------|-------|
| `docker compose -f docker-compose.dev.yml up` succeeds | ☐ | |
| Core-banking UPDATE → `domain.events` within 5s | ☐ | |
| State Service upserts persona; state_version increments | ☐ | |
| Outbox publishes to `twin.state.updated` | ☐ | |
| `./scripts/smoke-test.sh` exits 0 | ☐ | |
| `go test ./...` in state-service exits 0 | ☐ | |
| Schema compat CI fails on breaking Avro change | ☐ | |

---

## 7. Security and hygiene

| Check | Pass | Notes |
|-------|------|-------|
| No secrets in git (.env, API keys, passwords) | ☐ | |
| `.env.example` documents vars without real values | ☐ | |
| No `any` types without justification in Go public APIs | ☐ | |
| SQL uses parameterized queries (no string concat) | ☐ | |

---

## 8. P0/P1 blockers

Flag as **BLOCK** if any apply:

- [ ] Direct Kafka publish without outbox (data loss risk on crash)
- [ ] No idempotency → duplicate personas on replay
- [ ] Breaking Avro change without version bump / compat plan
- [ ] Credentials committed
- [ ] Phase 2+ component implemented (scope creep)

---

## 9. Review history

| Review | Date | Verdict | Notes |
|--------|------|---------|-------|
| [Baseline (planning)](./phase1-review-2026-06-13-baseline.md) | 2026-06-13 | APPROVE handoff | Docs-only; artifacts ready for implementation agent |
| [Implementation](./phase1-review-2026-06-13-implementation.md) | 2026-06-13 | APPROVE | Exit criteria verified; smoke test + go test pass |

**Current status**: Phase 1 implementation complete. See implementation review for Section 1–8 results.

---

## Review outcome template

```markdown
## Phase 1 Review — [date]

**Branch/PR**: 
**Reviewer**: 

### Summary
[1–2 sentences]

### Exit criteria
- [ ] All Section 6 checks pass (evidence: smoke-test output attached)

### Blockers (P0/P1)
- None / [list]

### Non-blocking suggestions
- [list]

### Verdict
APPROVE / REQUEST CHANGES
```
