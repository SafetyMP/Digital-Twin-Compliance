# Handoff: Phase 2 Implementation Agent

Use this document when starting Phase 2 implementation in a separate Cursor chat or agent session.

**Created**: 2026-06-13  
**Prerequisites**: Phase 1 complete ([phase1-review-checklist.md](./review/phase1-review-checklist.md)).

---

## Your mission

Implement **Phase 2 only** — real-time compliance monitoring and alert delivery:

> Flink CEP (3 patterns), Redis feature store, Alert Service + WebSocket, Next.js alert console, Grafana dashboards.

Read and follow:

1. [docs/phase2-implementation-spec.md](./phase2-implementation-spec.md) — **primary executable spec**
2. [docs/adr/008-phase2-foundation-decisions.md](./adr/008-phase2-foundation-decisions.md) — D10–D13
3. [AGENTS.md](../AGENTS.md) — update § Current phase when starting; extend out-of-scope list
4. [docs/data-flow.md](./data-flow.md) — `ComplianceAlertRaised`, `compliance.alerts`

Phase 1 services remain; do not break `./scripts/smoke-test.sh`.

---

## Suggested prompt

```
Context budget: read AGENTS.md + phase2-implementation-spec.md + ADR-008.
For Go alert-service work, add services/alert-service/AGENTS.md.
Load data-flow.md for alert envelope fields.

Implement Phase 2 per docs/phase2-implementation-spec.md.
Do NOT implement Phase 3+ (Cedar, immudb, Neo4j, Keycloak, XBRL).

Order: ADR-008 → schemas → payments migration → Redis/alert-db →
alert-service → State Service liquidity fields → Flink job →
Compose Flink → alert-console → Grafana → smoke-test-phase2 → CI.

Mark done only when smoke-test.sh AND smoke-test-phase2.sh pass.
```

---

## Definition of done

```bash
docker compose -f docker-compose.dev.yml up -d --wait
./scripts/seed.sh
# ... submit Flink job, register schemas/connectors as needed ...
./scripts/smoke-test.sh
./scripts/smoke-test-phase2.sh
cd services/alert-service && go test ./...
cd jobs/compliance-cep && mvn test
```

Copy [Phase 2 exit criteria checklist](./phase2-implementation-spec.md#14-phase-2-exit-criteria-checklist) into the PR description.

---

## Review

Use an extended checklist derived from [phase1-review-checklist.md](./review/phase1-review-checklist.md) plus Phase 2 exit criteria before merge.
