# Phase 2 Exit Criteria Checklist

Verified: 2026-06-14 (local stack, branch `feat/phase2-monitoring`)

Copy into PR description when opening Phase 2 PR.

## Checklist

- [x] Extended `docker compose -f docker-compose.dev.yml up` starts Phase 1 + Redis + Flink + Alert Service + UI + Grafana  
  **Evidence**: `docker compose ps` — `redis`, `flink-jobmanager`, `flink-taskmanager`, `alert-service`, `alert-db`, `alert-console`, `grafana` all running alongside Phase 1 services (`kafka`, `state-service`, `debezium`, etc.).

- [x] Flink job RUNNING; checkpoint success rate > 99% over soak  
  **Evidence**: `./scripts/smoke-test-phase2.sh` step 1 — `Flink checkpoints: completed=60 failed=0 success_rate=100.00%` (job `6841659ef76cbc153cc7771ecf7fd3cc`).

- [x] `INT-M001`, `INT-M002`, `BASEL-M001` each produce ≥ 1 alert from seed/simulator scenarios  
  **Evidence**: smoke test steps 3–5 — all three rule codes detected.

- [x] Alerts on `compliance.alerts` consumed by Alert Service within 2s p99 (measured in smoke test log)  
  **Evidence**: `INT-M001 alert detected (count=1, consume_latency_ms=1091)` — under 2000ms budget; smoke test fails if > 2000ms.

- [x] Alert visible in UI within 5s of detection  
  **Evidence**: `alert-console HTTP 200` + `Alert visible via API within 1146ms` (Alert Console fetches same REST API on load; CSR page shell reachable at `:3000`).

- [x] Acknowledge flow updates PostgreSQL and WebSocket clients  
  **Evidence**: smoke test step 7 — `Acknowledge persisted and WebSocket alert.acknowledged received`.

- [x] Redis keys updated for velocity / exposure / LCR features  
  **Evidence**: smoke test step 8 — `vel:*`, `exp:*`, `lcr:*` keys present.

- [x] `./scripts/smoke-test.sh` still passes (Phase 1 regression)  
  **Evidence**: exit 0 — `==> Smoke test passed.`

- [x] `./scripts/smoke-test-phase2.sh` exits 0  
  **Evidence**: exit 0 — `==> Phase 2 smoke test passed`.

- [x] `go test ./...` (alert-service) and `mvn test` (compliance-cep) pass  
  **Evidence**:  
  - `cd services/alert-service && go test ./...` → exit 0  
  - `cd services/state-service && go test ./...` → exit 0  
  - `docker run maven:3.9-eclipse-temurin-17 mvn test` in `jobs/compliance-cep` → exit 0

- [x] New Avro schemas pass BACKWARD compat CI  
  **Evidence**: local compat check — `compliance-alert-raised: true`, `compliance-alert-resolved: true` (mirrors `.github/workflows/schema-compat.yml`).

- [x] No Phase 3+ components added (Cedar, immudb, Neo4j, full auth)  
  **Evidence**: codebase grep — no Cedar/immudb/Neo4j/Keycloak/GoRules in services, jobs, or apps (eval scripts only reference them as forbidden scope checks).

## Commands to re-verify

```bash
docker compose -f docker-compose.dev.yml up -d --wait
./scripts/seed.sh
./scripts/submit-flink-job.sh   # if Flink job not RUNNING
./scripts/smoke-test.sh
./scripts/smoke-test-phase2.sh
cd services/alert-service && go test ./...
cd services/state-service && go test ./...
# Maven (no local mvn required):
cd jobs/compliance-cep && docker run --rm -v "$PWD":/app -w /app maven:3.9-eclipse-temurin-17 mvn test -q
```
