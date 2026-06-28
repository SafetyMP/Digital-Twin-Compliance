# Phase 3 Exit Criteria Checklist

Maps [phase3-implementation-spec.md §13](../phase3-implementation-spec.md#13-phase-3-exit-criteria-checklist) to fresh local evidence.

Verified: 2026-06-28 — local cold-start stack (`phase3b-flink-zen-lcr`) and CI on `main` after [PR #19](https://github.com/SafetyMP/Digital-Twin-Compliance/pull/19) merge ([CI run](https://github.com/SafetyMP/Digital-Twin-Compliance/actions/runs/28335248362), `6240a80`).

Merged to `main` 2026-06-28.

## Checklist

- [x] Compose starts Phase 1–2 services **plus** immudb, audit, cedar, decision, audit-explorer
  **Evidence**: `docker compose ps` — all 22 services up; healthy: `core-banking-db`, `state-db`, `kafka`, `schema-registry`, `debezium`, `state-service`, `redis`, `alert-db`, `alert-service`, `flink-jobmanager`, `grafana`, plus Phase 3 `immudb`, `audit-db`, `audit-service`, `cedar-service`, `decision-service` (all `healthy`) and `audit-explorer` (up). `seed.sh` created Phase 3 topics `compliance.audit.pending`, `compliance.audit.pending.dlq`, `compliance.audit.recorded`.

- [x] 5 Cedar policies pass policy gate in CI
  **Evidence**: `policies/cedar/*.cedar` = 5 (`int-r003`, `int-r004`, `corep-r005`, `emir-r004`, `dora-r001`). `cedar-service` `/api/v1/health` → `policiesLoaded: 5`, `schemaLoaded: true`, `status: ok`, ruleCodes `[DORA-R001, INT-R003, INT-R004, COREP-R005, EMIR-R004]`. `.github/workflows/policy-gates.yml` installs the Cedar CLI (`install-cedar-cli.sh`) and runs `run-policy-ci.sh`.
  **Caveat**: the CI gate runs `cedar check-parse` per policy (the spec wording says "analyze"); the Cedar CLI is **not installed on this host**, so the parse gate was verified by reading the workflow + confirming all 5 policies load at runtime, **not** by a live local `cedar` invocation.

- [x] 5 Zen models pass fixture tests in CI
  **Evidence**: `policies/zen/*.json` = 5 models + 5 `policies/zen/fixtures/*.json`. `decision-service` `/api/v1/rules` lists `BASEL-R001, COREP-R001, COREP-R002, INT-R001, INT-R002` (all `v1.0.0`). `cd services/decision-service && go test ./...` → exit 0 (includes `TestZenFixtures` in `internal/engine`).

- [x] `compliance.audit.pending` → immudb → `compliance.audit.recorded` pipeline works
  **Evidence**: Phase 3 smoke step 4 produced a fresh `Alert` audit entry (`evidenceRef=e8a03c91-6fdd-41a5-900e-2865dc653fd7`) consumed from `compliance.audit.pending` and written to immudb; `audit-service` reports 30 entries on `/api/v1/audit/verify`. `audit-service` is `healthy` (health includes immudb connectivity).

- [x] Hash chain verification API returns valid for 100% of entries in smoke run
  **Evidence**: `GET /api/v1/audit/verify` → `{ "valid": true, "checkedCount": 30 }`; `./scripts/verify-audit-chain.sh` → `Audit chain valid (30 entries checked)` (exit 0).

- [x] New alerts have non-null `evidenceRef` within 10s
  **Evidence**: Phase 3 smoke step 4 — after an INT-M001 burst, `evidenceRef` was set within the smoke poll window (≤10s + burst). Alert DB query: `with_ref=3 / total=3` (100% non-null).

- [x] Audit Explorer search returns alert audit entries with integrity badge
  **Evidence**: Phase 3 smoke step 6 — `audit-explorer` `/api/audit/entries?limit=5` returned ≥1 entry. UI renders integrity badge + chain status: `apps/audit-explorer/src/app/entries/[entryId]/page.tsx` ("integrity OK", payloadHash/previousHash) and `src/app/page.tsx` (`Chain valid/broken — N entries`).

- [x] `./scripts/smoke-test-phase3.sh` exits 0
  **Evidence**: `SMOKE_PHASE3_SKIP_PREREQS=1 ./scripts/smoke-test-phase3.sh` → `Phase 3 smoke test passed` (exit 0): health, Cedar INT-R003 deny→allow, Zen BASEL-R001 LCR 0.90 → Deny, evidenceRef, chain verify, Audit Explorer.

- [x] `./scripts/smoke-test-phase2.sh` still passes (regression)
  **Evidence**: exit 0 — `==> Phase 2 smoke test passed`. INT-M001 (`consume_latency_ms=1403`), INT-M002, BASEL-M001 alerts fired; WebSocket raise + acknowledge; Redis `vel:*`/`exp:*`/`lcr:*` keys; Grafana healthy. Phase 1 base: `./scripts/smoke-test.sh` → `==> Smoke test passed.` (exit 0).

- [x] `go test ./...` passes for audit, cedar, decision services
  **Evidence**:
  - `cd services/cedar-service && go test ./...` → exit 0 (`api`, `audit`, `config`, `engine`)
  - `cd services/decision-service && go test ./...` → exit 0 (`api`, `audit`, `config`, `decision`, `engine`)
  - `cd services/audit-service && go test ./...` → exit 0 (`cmd/server`, `api`, `chain`, `config`, `consumer`, `events`, `store`)

- [x] No Phase 4+ components added (Neo4j, simulation, XBRL)
  **Evidence**: repo grep for `neo4j|simulation-service|networkx|xbrl|sdmx|clickhouse` outside docs/evals/md — only hit is a scope-guard function name in `scripts/run-live-evals-phase2.sh` (`check_phase3_scope_boundary`).

- [x] Keycloak not added (mock principal only per D20)
  **Evidence**: same grep — no `keycloak` usage in services/jobs/apps. `cedar-service` uses dev `X-Principal`/`X-Roles` headers (smoke step 2 sends `X-Roles: Reporter`).

## Not verified / deferred

- **Alert latency p99 < 2s and Flink checkpoint success rate > 99% over soak**: not measured. Phase 2 smoke observed single-shot `consume_latency_ms=1403` (< 2000ms budget); Flink reported `no checkpoint history yet (non-fatal during cold start)` on cold start.
- **Cedar CLI gate**: confirmed in CI on PR #19 and post-merge `main` (`Policy CI` step — `cedar check-parse` on all 5 policies + Zen fixtures).
- **Rule evaluation latency < 5ms p99** (roadmap success metric): not benchmarked.

## Commands to re-verify

```bash
docker compose -f docker-compose.dev.yml up -d --wait
./scripts/seed.sh
docker compose -f docker-compose.dev.yml restart alert-service state-service \
  && docker compose -f docker-compose.dev.yml up -d --wait alert-service state-service
./scripts/wait-outbox-drained.sh
./scripts/submit-flink-job.sh            # if compliance-cep not RUNNING
./scripts/smoke-test.sh
./scripts/smoke-test-phase2.sh
SMOKE_PHASE3_SKIP_PREREQS=1 ./scripts/smoke-test-phase3.sh
./scripts/verify-audit-chain.sh
for d in audit cedar decision; do (cd services/$d-service && go test ./...); done
```
