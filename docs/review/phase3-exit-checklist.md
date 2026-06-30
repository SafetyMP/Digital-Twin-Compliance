# Phase 3 Exit Criteria Checklist

**Phase 3: COMPLETE** — release [v0.1.0](https://github.com/SafetyMP/Digital-Twin-Compliance/releases/tag/v0.1.0) on `main` (2026-06-29).

Maps [phase3-implementation-spec.md §13](../phase3-implementation-spec.md#13-phase-3-exit-criteria-checklist) to fresh local evidence.

Verified: 2026-06-29 — `main` at `7ee99c7` after [PR #69](https://github.com/SafetyMP/Digital-Twin-Compliance/pull/69) (Flink 1.20, CHANGELOG, deploy docs). Prior verification: 2026-06-28 — [PR #19](https://github.com/SafetyMP/Digital-Twin-Compliance/pull/19) merge ([CI run](https://github.com/SafetyMP/Digital-Twin-Compliance/actions/runs/28335248362)).

---

## Mechanical checklist

- [x] Compose starts Phase 1–2 services **plus** immudb, audit, cedar, decision, audit-explorer
  **Evidence**: `docker compose ps` — all services up; Phase 3 `immudb`, `audit-db`, `audit-service`, `cedar-service`, `decision-service` healthy; `audit-explorer` up. `seed.sh` creates `compliance.audit.pending`, `compliance.audit.pending.dlq`, `compliance.audit.recorded`.

- [x] 5 Cedar policies pass policy gate in CI
  **Evidence**: `policies/cedar/*.cedar` = 5. `cedar-service` `/api/v1/health` → `policiesLoaded: 5`. `.github/workflows/policy-gates.yml` runs `run-policy-ci.sh` (`cedar check-parse` per policy).

- [x] 5 Zen models pass fixture tests in CI
  **Evidence**: `policies/zen/*.json` = 5 models + fixtures. `decision-service` `/api/v1/rules` lists all five rule codes. `cd services/decision-service && go test ./...` → exit 0.

- [x] `compliance.audit.pending` → immudb → `compliance.audit.recorded` pipeline works
  **Evidence**: Phase 3 smoke step 4 — alert audit entry consumed and written; `/api/v1/audit/verify` valid.

- [x] Hash chain verification API returns valid for 100% of entries in smoke run
  **Evidence**: `GET /api/v1/audit/verify` → `{ "valid": true }`; `./scripts/verify-audit-chain.sh` exit 0.

- [x] New alerts have non-null `evidenceRef` within 10s
  **Evidence**: Phase 3 smoke — `evidenceRef` set within poll window; alert DB `with_ref = total`.

- [x] Audit Explorer search returns alert audit entries with integrity badge
  **Evidence**: smoke step 6; UI shows chain status and integrity badges.

- [x] `./scripts/smoke-test-phase3.sh` exits 0
  **Evidence**: `SMOKE_PHASE3_SKIP_PREREQS=1 ./scripts/smoke-test-phase3.sh` → `Phase 3 smoke test passed`.

- [x] `./scripts/smoke-test-phase2.sh` still passes (regression)
  **Evidence**: INT-M001/M002/BASEL-M001 alerts, WebSocket, Redis keys; `./scripts/smoke-test.sh` passes.

- [x] `go test ./...` passes for audit, cedar, decision services
  **Evidence**: all three services `go test ./...` exit 0.

- [x] No Phase 4+ components added (Neo4j, simulation, XBRL)
  **Evidence**: no `services/graph-service`, `services/simulation-service`, or Neo4j in Compose (Phase 4 planning docs only).

- [x] Keycloak not added (mock principal only per D20)
  **Evidence**: Cedar uses `X-Principal` / `X-Roles` headers only.

---

## Phase 3b (stretch — complete)

- [x] Flink CEP → Decision Service for **INT-M001** (`INT-R001`) when `CEP_DECISION_SERVICE_URL` set
- [x] Flink CEP → Decision Service for **INT-M002** (`INT-R002`) when URL set
- [x] Flink CEP → Decision Service for **BASEL-M001** (`BASEL-R001`) when URL set
  **Evidence**: `jobs/compliance-cep` DecisionServiceClient; inline thresholds remain fallback on HTTP failure. Rebuild jar + `./scripts/submit-flink-job.sh` after Java changes (AGENTS.md gotcha).

---

## Release v0.1.0

- [x] Semver tag `v0.1.0` on `main` with [CHANGELOG.md](../../CHANGELOG.md) entry
- [x] Flink runtime aligned to **1.20** (`docker-compose.dev.yml`, CEP `pom.xml` kafka connector `3.4.0-1.20`)
- [x] GHCR publish workflow + [docs/deployment.md](../deployment.md) release validation section
- [x] Eight service images in `docker-compose.deploy.yml` for Phase 1–3 runtime

---

## Agent / parallel workflow (optional)

- [x] `./scripts/check-subagent-preflight.sh` — parent-only advisory before Task subagents ([handoff-parallel-parent.md](../handoff-parallel-parent.md))

---

## Stretch metrics (roadmap — sampled, not soak gates)

| Metric | Target | Status |
|--------|--------|--------|
| Alert end-to-end p99 | < 2s | **Deferred** — single-shot Phase 2 smoke `consume_latency_ms≈1403` (< budget); no multi-hour soak |
| Flink checkpoint success | > 99% | **Deferred** — cold start may show empty checkpoint history; monitor in staging |
| Rule evaluate p99 (Cedar/Zen) | < 5ms | **Sampled** — run `./scripts/measure-phase3-latency.sh` on warm stack; dev Compose often exceeds 5ms (WARN, not exit 1) |

```bash
# Optional benchmark (requires running stack)
./scripts/measure-phase3-latency.sh
```

---

## Phase 4 handoff

Phase 3 is closed for implementation. Next work:

- [phase4-readiness.md](./phase4-readiness.md) — prerequisites (satisfied)
- [phase4-implementation-spec.md](../phase4-implementation-spec.md)
- [handoff-phase4-agent.md](../handoff-phase4-agent.md)

---

## Commands to re-verify

```bash
docker compose -f docker-compose.dev.yml up -d --wait
./scripts/seed.sh
docker compose -f docker-compose.dev.yml restart alert-service state-service \
  && docker compose -f docker-compose.dev.yml up -d --wait alert-service state-service
./scripts/wait-outbox-drained.sh
./scripts/submit-flink-job.sh
./scripts/smoke-test.sh
./scripts/smoke-test-phase2.sh
SMOKE_PHASE3_SKIP_PREREQS=1 ./scripts/smoke-test-phase3.sh
./scripts/verify-audit-chain.sh
for d in audit cedar decision; do (cd services/$d-service && go test ./...); done
./scripts/check-subagent-preflight.sh
./scripts/measure-phase3-latency.sh   # optional
```
