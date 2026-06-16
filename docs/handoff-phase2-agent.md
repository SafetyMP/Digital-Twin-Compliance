# Handoff: Phase 2 Agent

> **Status (2026-06):** Phase 2 is **implemented** in the repository. Mechanical smoke and unit tests pass locally ([phase2-exit-checklist.md](./review/phase2-exit-checklist.md)). Use this document for **extensions, fixes, and verification** — not greenfield implementation. Formal Phase 2 completion still requires mechanical nightly evals and the behavior eval pillar ([AGENTS.md](../AGENTS.md#behavior-evals-phase-2)).

**Created**: 2026-06-13  
**Prerequisites**: Phase 1 complete ([phase1-review-checklist.md](./review/phase1-review-checklist.md)).

---

## When to use this handoff

| Task | Start with |
|------|------------|
| Bug fix or extension in Flink, alert-service, alert-console, Grafana | This doc + [phase2-implementation-spec.md](./phase2-implementation-spec.md) |
| Verification / `/verify-phase2` | [handoff-verification-agent.md](./handoff-verification-agent.md) |
| Phase 3+ (Cedar, immudb, Neo4j) | [roadmap.md](./roadmap.md) Phase 3 — do not use this handoff |

---

## Scope reference

Real-time compliance monitoring and alert delivery:

> Flink CEP (3 patterns), Redis feature store, Alert Service + WebSocket, Next.js alert console, Grafana dashboards.

Read and follow:

1. [docs/phase2-implementation-spec.md](./phase2-implementation-spec.md) — executable spec and exit criteria
2. [docs/adr/008-phase2-foundation-decisions.md](./adr/008-phase2-foundation-decisions.md) — D10–D13
3. [AGENTS.md](../AGENTS.md) — current phase contract
4. [docs/data-flow.md](./data-flow.md) — `ComplianceAlertRaised`, `compliance.alerts`
5. [services/alert-service/AGENTS.md](../services/alert-service/AGENTS.md) — for Go alert-service work

Phase 1 services remain; do not break `./scripts/smoke-test.sh`.

---

## Suggested prompt (extensions / fixes)

```
Context budget: read AGENTS.md + phase2-implementation-spec.md + ADR-008.
For Go alert-service work, add services/alert-service/AGENTS.md.
Load data-flow.md for alert envelope fields.

Extend or fix Phase 2 per docs/phase2-implementation-spec.md.
Do NOT implement Phase 3+ (Cedar, immudb, Neo4j, Keycloak, XBRL).

Verify with smoke-test.sh AND smoke-test-phase2.sh before claiming done.
```

---

## Definition of done

```bash
docker compose -f docker-compose.dev.yml up -d --wait
./scripts/seed.sh
./scripts/submit-flink-job.sh   # if Flink job not RUNNING
./scripts/smoke-test.sh
./scripts/smoke-test-phase2.sh
cd services/alert-service && go test ./...
cd services/state-service && go test ./...
# Maven (no local mvn required):
docker run --rm -v "$PWD/jobs/compliance-cep:/app" -w /app maven:3.9-eclipse-temurin-17 mvn -q test
```

Copy [Phase 2 exit criteria checklist](./phase2-implementation-spec.md#14-phase-2-exit-criteria-checklist) or [review/phase2-exit-checklist.md](./review/phase2-exit-checklist.md) into the PR description.

---

## Review

Use [phase2-exit-checklist.md](./review/phase2-exit-checklist.md) plus [phase1-review-checklist.md](./review/phase1-review-checklist.md) for any shared Phase 1 regressions before merge.
