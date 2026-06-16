# Handoff: Phase 3 Implementation Agent

Use this document when starting Phase 3 implementation in a **new** Cursor chat.

**Created**: 2026-06-15  
**Planning deliverables**: ADR-009, Phase 3 spec, AGENTS.md phase pointer, this handoff.

---

## Your mission

Implement **Phase 3 only** — rules engine and audit ledger:

> Cedar Policy Service, Decision Service (Zen), Audit Service (immudb), alert `evidenceRef` wiring, Audit Explorer UI, policy CI gates.

Read and follow:

1. [docs/phase3-implementation-spec.md](./phase3-implementation-spec.md) — **primary executable spec**
2. [AGENTS.md](../AGENTS.md) — commands, layout, definition of done
3. [docs/adr/009-phase3-foundation-decisions.md](./adr/009-phase3-foundation-decisions.md) — D15–D20

---

## Decisions already made (do not re-open)

| ID | Decision | Implementation hint |
|----|----------|---------------------|
| D15 | immudb in Compose for dev/CI | Single node; not K8s HA for exit criteria |
| D16 | Kafka `compliance.audit.pending` → Audit Service only writer | No direct immudb from alert/cedar/decision |
| D17 | Standalone cedar-service + decision-service | HTTP evaluate APIs |
| D18 | Phase 3a: audit + API eval; Flink stays inline | Do not refactor CEP job unless explicitly Phase 3b |
| D19 | Filesystem artifact stub | No S3/MinIO required locally |
| D20 | Mock principal (`X-Principal`, `X-Roles`) | No Keycloak in Phase 3 |

Phase 2 decisions (ADR-008) still apply — PostgreSQL alerts, JSON Kafka envelopes in dev.

---

## In scope

| Deliverable | Spec section |
|-------------|--------------|
| immudb + Compose wiring | §4 |
| Audit Service | §6.1 |
| Cedar Service + `policies/cedar/` | §6.2 |
| Decision Service + `policies/zen/` | §6.3 |
| Alert Service `evidenceRef` | §6.4 |
| Audit Explorer UI | §7 |
| `smoke-test-phase3.sh`, `verify-audit-chain.sh` | §8 |
| `policy-gates.yml` + CI extend | §9 |

---

## Explicitly out of scope

- Neo4j, Simulation, XBRL reporting (Phase 4–5)
- Keycloak / OIDC middleware
- Flink hot-path Zen migration (Phase 3b stretch)
- S3 Object Lock in dev

---

## Suggested prompt for implementation agent

```
Context budget: read AGENTS.md + phase3-implementation-spec.md only.
For touched services, read services/<svc>/AGENTS.md once created.
Load docs/data-flow.md only for audit envelope / Kafka tasks.

Implement Phase 3 per docs/phase3-implementation-spec.md and ADR-009.
Do NOT implement Phase 4+ (Neo4j, simulation, XBRL).
Do NOT migrate Flink CEP to Zen unless task is explicitly Phase 3b.

Order: immudb Compose → audit-service → policies → cedar-service →
decision-service → alert-service evidenceRef → audit-explorer →
smoke-test-phase3.sh → CI.

Mark done only when smoke-test-phase3.sh, smoke-test-phase2.sh (regression),
and go test for all three new services pass.
```

---

## Verification handoff

When implementation claims done, start a **fresh** verification chat with [handoff-verification-agent.md](./handoff-verification-agent.md) adapted for Phase 3:

```bash
./scripts/smoke-test.sh
./scripts/smoke-test-phase2.sh
./scripts/smoke-test-phase3.sh
./scripts/verify-audit-chain.sh
./scripts/run-policy-ci.sh
```

---

## Parallel tracks (optional)

Only if file boundaries are explicit:

| Track | Paths | Owner verifies |
|-------|-------|----------------|
| Integration | `docker-compose.dev.yml`, `scripts/smoke-test-phase3.sh`, `.env.example` | Parent smoke |
| Policy engines | `services/cedar-service/`, `services/decision-service/`, `policies/` | `run-policy-ci.sh` |
| Audit path | `services/audit-service/`, alert-service audit wiring | `verify-audit-chain.sh` |
| UI | `apps/audit-explorer/` | Smoke step 6 |

Parent owns conflict resolution and full smoke sequence.
