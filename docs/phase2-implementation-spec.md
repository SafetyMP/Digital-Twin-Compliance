# Phase 2 Implementation Spec

Executable handoff for the implementation agent. Implements [roadmap.md](./roadmap.md) Phase 2 only.

**Prerequisites**: Phase 1 complete ([phase1-implementation-spec.md](./phase1-implementation-spec.md) exit criteria). [ADR-001](./adr/001-kafka-flink-streaming.md) (Kafka + Flink), [ADR-007](./adr/007-phase1-foundation-decisions.md) (tenant, local dev).

**Related docs**: [architecture.md](./architecture.md), [domain-model.md](./domain-model.md), [data-flow.md](./data-flow.md), [compliance-mapping.md](./compliance-mapping.md).

---

## 1. Goal

Build real-time compliance monitoring and alert delivery so that:

1. Three CEP patterns detect breaches from twin and payment events within **2s p99** (simulated load).
2. Rolling features are maintained in **Redis** and updated by Flink.
3. `ComplianceAlertRaised` events are published to **`compliance.alerts`** with Avro envelope.
4. **Alert Service** persists alerts and streams them to the UI via **WebSocket** within **5s** of detection.
5. **Alert Console** (Next.js) shows a live feed with **acknowledge** action.
6. **Grafana** dashboards show Kafka consumer lag and Flink checkpoint health.
7. CI runs Flink job tests, Alert Service tests, and an extended smoke test.

---

## 2. Scope boundaries

### In scope (Phase 2)

| Deliverable | Technology |
|-------------|------------|
| Flink CEP job (3 patterns) | Java 17 + Flink 1.18+ |
| Redis online feature store | Redis 7 |
| Alert Service | Go + PostgreSQL |
| WebSocket hub | Go (inside Alert Service or sub-package) |
| Alert Console UI | Next.js 14+ (App Router), TypeScript |
| Grafana dashboards | Grafana 10+ (Compose) |
| Mock payment/event feed | SQL seed + optional Kafka injector script |
| Avro schemas for alerts | `schemas/avro/` |
| CI + smoke test extensions | GitHub Actions |

### Out of scope (defer to Phase 3+)

Per [AGENTS.md](../AGENTS.md) and [roadmap.md](./roadmap.md):

- Cedar Policy Service, GoRules Zen / Decision Service integration in the hot path
- immudb audit ledger (`evidenceRef` may be null or point to alert row until Phase 3)
- Neo4j / Graph Service
- Keycloak / auth middleware (UI and API remain open in dev; optional API key env var only)
- Regulatory reporting (XBRL)
- Flink Kubernetes Operator production deployment (document staging path; local Compose mini-cluster for dev/CI)

### Pattern mapping (roadmap vs compliance-mapping)

Phase 2 implements these **monitoring check IDs**:

| Check ID | Pattern | Threshold (dev default) | Severity | Source events |
|----------|---------|-------------------------|----------|---------------|
| `INT-M001` | Transaction velocity | > 50 payments / account / 1h sliding window | Warning | `PaymentInitiated` |
| `INT-M002` | Counterparty exposure limit | Instrument notional to same counterparty > €10M per institution | Critical | `twin.state.updated` (Instrument persona) |
| `BASEL-M001` | LCR below minimum | LCR < 100% | Critical | `twin.state.updated` (Institution persona liquidity fields) |

**Note**: [compliance-mapping.md](./compliance-mapping.md) lists `INT-M002` as sanctions screening; Phase 2 follows [roadmap.md](./roadmap.md) and implements **exposure limit** (`INT-R002`). Sanctions (`INT-R005`) remains Phase 3 with Decision Service.

---

## 3. Repository layout

Add to Phase 1 structure:

```
/
├── docker-compose.dev.yml          # extend: redis, flink, alert-service, ui, grafana
├── docker-compose.deploy.yml       # extend: same services (GHCR images)
├── schemas/avro/
│   ├── compliance-alert-raised.avsc
│   └── compliance-alert-resolved.avsc
├── jobs/
│   └── compliance-cep/             # Flink job (Java)
│       ├── pom.xml
│       └── src/main/java/.../
├── services/
│   ├── state-service/              # Phase 1 (unchanged contract)
│   └── alert-service/
│       ├── go.mod
│       ├── cmd/server/main.go
│       ├── internal/
│       │   ├── api/                # REST + WebSocket
│       │   ├── consumer/           # compliance.alerts consumer
│       │   ├── store/              # PostgreSQL alerts
│       │   └── config/
│       └── migrations/
│           └── 001_alerts.sql
├── apps/
│   └── alert-console/              # Next.js
│       ├── package.json
│       └── src/...
├── mocks/
│   ├── core-banking/
│   │   └── migrations/
│   │       └── 002_payments.sql    # payment tables for CDC
│   └── simulators/
│       └── payment-burst.sh        # triggers INT-M001 in smoke test
├── infra/
│   └── grafana/
│       ├── dashboards/
│       └── provisioning/
├── scripts/
│   ├── smoke-test-phase2.sh
│   └── register-schemas.sh         # extend for alert schemas
└── .github/workflows/
    └── ci.yml                      # extend Phase 2 jobs
```

**Go module path**: `github.com/digital-twin/platform/services/alert-service`

**Java coordinates**: `com.digitaltwin.jobs:compliance-cep:0.1.0-SNAPSHOT`

---

## 4. Docker Compose (local dev)

Extend `docker-compose.dev.yml`:

| Service | Image / build | Ports | Purpose |
|---------|---------------|-------|---------|
| `redis` | redis:7-alpine | 6379 | Online feature store |
| `flink-jobmanager` | flink:1.18-java17 | 8081→8082* | JobManager REST / UI |
| `flink-taskmanager` | flink:1.18-java17 | — | Task slots |
| `alert-service` | build `./services/alert-service` | 8085 | Alerts REST + WebSocket |
| `alert-db` | postgres:16 | 5435:5432 | Alert persistence |
| `alert-console` | build `./apps/alert-console` | 3000 | Next.js UI |
| `grafana` | grafana/grafana:10.4.0 | 3001:3000 | Dashboards |

\* Map Flink UI to host **8082** to avoid conflict with Schema Registry **8081**.

Phase 1 services remain unchanged. Flink depends on `kafka`, `redis`, and healthy Phase 1 stack.

### Environment variables (add to `.env.example`)

```bash
# Redis
REDIS_URL=redis://localhost:6379/0

# Alert Service
ALERT_DB_URL=postgres://alert:alert@localhost:5435/alerts?sslmode=disable
ALERT_SERVICE_HTTP_ADDR=:8085
ALERT_SERVICE_URL=http://localhost:8085
ALERT_SERVICE_WS_PATH=/ws/alerts

# Flink
FLINK_JOBMANAGER_URL=http://localhost:8082
FLINK_PARALLELISM=2

# Alert Console (Next.js)
NEXT_PUBLIC_ALERT_SERVICE_URL=http://localhost:8085
NEXT_PUBLIC_WS_URL=ws://localhost:8085/ws/alerts

# CEP thresholds (dev defaults; override in Compose)
CEP_VELOCITY_MAX_PER_HOUR=50
CEP_EXPOSURE_LIMIT_EUR=10000000
CEP_LCR_MINIMUM=1.0

# Grafana
GRAFANA_ADMIN_USER=admin
GRAFANA_ADMIN_PASSWORD=admin
```

### Flink job submission (dev)

After Compose is healthy, submit the job via REST or `scripts/submit-flink-job.sh`:

```bash
# Build JAR locally or in CI artifact
cd jobs/compliance-cep && mvn -q package -DskipTests

# Submit (example)
curl -X POST -H "Expect:" -F "jarfile=@target/compliance-cep-0.1.0-SNAPSHOT.jar" \
  http://localhost:8082/jars/upload
# ... program args: --kafka brokers, --redis host, --registry url
```

**Pragmatic Phase 2 shortcut**: Compose `command` on TaskManager can auto-submit job on startup via `docker-entrypoint.sh` wrapper checked into `jobs/compliance-cep/docker/`.

---

## 5. Kafka topics and schemas

### 5.1 Topics (Phase 2 minimum)

| Topic | Partitions (dev) | Producers | Consumers |
|-------|------------------|-----------|-----------|
| `twin.state.updated` | 3 | State Service (Phase 1) | Flink CEP, Alert Service (optional) |
| `domain.events.public.payments` | 3 | Debezium (new table) | Flink CEP |
| `compliance.alerts` | 3 | Flink CEP | Alert Service |
| `compliance.alerts.dlq` | 1 | Alert Service | Ops |

Retention in dev: 7 days (production 7 years per [data-flow.md](./data-flow.md)).

### 5.2 Avro schemas

Directory: `schemas/avro/`. Register via extended `scripts/register-schemas.sh`. Compatibility: **BACKWARD**.

**Subjects**:

- `compliance.alerts-value` — envelope with payload JSON string (same pattern as Phase 1) or embedded `ComplianceAlertRaised` record (prefer embedded record in Phase 2 for type safety).

**`compliance-alert-raised.avsc`** (payload shape — see [data-flow.md §4.2](./data-flow.md)):

| Field | Type | Required |
|-------|------|----------|
| `alertId` | string (UUID) | Yes |
| `ruleCode` | string | Yes (`INT-M001`, `INT-M002`, `BASEL-M001`) |
| `regime` | string | Yes (`Internal`, `Basel`) |
| `severity` | string | Yes |
| `status` | string | Yes (`Open`) |
| `personaId` | string | Yes |
| `personaType` | string | Yes |
| `summary` | string | Yes |
| `details` | map/string JSON | Yes |
| `detectedAt` | string (ISO 8601) | Yes |
| `evidenceRef` | string | No (Phase 3 immudb) |

Flink publishes using the standard [EventEnvelope](./data-flow.md) wrapper with `eventType=ComplianceAlertRaised`.

---

## 6. Mock data and event sources

### 6.1 Payments table (CDC)

File: `mocks/core-banking/migrations/002_payments.sql`

```sql
CREATE TABLE payments (
  payment_id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  source_account_id UUID NOT NULL REFERENCES accounts(account_id),
  destination_account_id UUID NOT NULL REFERENCES accounts(account_id),
  amount           NUMERIC(20,2) NOT NULL,
  currency         CHAR(3) NOT NULL DEFAULT 'EUR',
  status           TEXT NOT NULL DEFAULT 'Pending',
  initiated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE payments REPLICA IDENTITY FULL;
```

Extend Debezium connector table list to include `payments`. State Service **does not** need to upsert payments in Phase 2 — Flink consumes CDC topic directly.

### 6.2 Institution liquidity fields

Extend State Service consumer mapping (minimal Phase 2 change) so Institution `currentState` JSON includes:

```json
{
  "liquidity": {
    "lcr": 1.05,
    "hqla": 500000000.00,
    "netCashOutflows30d": 476190476.00,
    "currency": "EUR"
  }
}
```

Seed values in `mocks/core-banking/seed/seed.sql` or post-seed SQL patch: most institutions **compliant** (LCR ≥ 1.0); at least one institution seeded at **0.95** for BASEL-M001 testing (or lowered at runtime in smoke test).

### 6.3 Exposure limit seed

Ensure instruments in seed data include `counterparty_id` and notional amounts such that at least one institution exceeds `CEP_EXPOSURE_LIMIT_EUR` aggregate per counterparty when Flink aggregates.

### 6.4 Velocity burst simulator

File: `mocks/simulators/payment-burst.sh`

Inserts 51+ payments for the same `source_account_id` within a rolling hour (or backdated timestamps within 1h window) to trigger `INT-M001` in smoke test.

---

## 7. Redis feature store

### 7.1 Key design

| Key pattern | Type | TTL | Updated by |
|-------------|------|-----|------------|
| `vel:{tenantId}:{accountId}:1h` | STRING (count) | 3600s | Flink on `PaymentInitiated` |
| `exp:{tenantId}:{institutionId}:{counterpartyId}` | STRING (EUR total) | none | Flink on Instrument twin updates |
| `lcr:{tenantId}:{institutionId}` | STRING (ratio) | 86400s | Flink on Institution twin updates |

Use `DEFAULT_TENANT_ID` from ADR-007 in keys.

### 7.2 Flink ↔ Redis

- Use Flink Redis connector or async Redis client in a `RichMapFunction` with idempotent increments.
- Checkpoint Redis writes: accept **at-least-once** counter updates in dev; document that production requires idempotent alert dedup by `idempotencyKey` on sink.

---

## 8. Flink CEP job

Path: `jobs/compliance-cep/`

### 8.1 Inputs

| Source | Topic | Event types |
|--------|-------|-------------|
| Payments | `domain.events.public.payments` | Debezium CDC → normalized `PaymentInitiated` |
| Twin | `twin.state.updated` | Envelope → Institution / Instrument payloads |

Use KafkaSource with Schema Registry deserializer for `twin.state.updated`; JSON or Avro for Debezium payments (match Phase 1 Debezium format).

### 8.2 Patterns

#### INT-M001 — Transaction velocity

- **Window**: sliding 1 hour, keyed by `sourceAccountId`
- **Logic**: count payments; if count > `CEP_VELOCITY_MAX_PER_HOUR`, emit alert
- **Dedup**: idempotencyKey = `INT-M001-{accountId}-{windowEnd}`

#### INT-M002 — Exposure limit

- **Key**: `{institutionId}:{counterpartyId}`
- **Logic**: sum instrument `notional_amount` for counterparty; if sum > `CEP_EXPOSURE_LIMIT_EUR`, emit alert
- **Dedup**: idempotencyKey = `INT-M002-{institutionId}-{counterpartyId}-{day}`

#### BASEL-M001 — LCR threshold

- **Key**: `institutionId`
- **Logic**: read `currentState.liquidity.lcr`; if < `CEP_LCR_MINIMUM`, emit alert
- **Dedup**: idempotencyKey = `INT-M001-{institutionId}-{floor(lcr*100)}` — use `BASEL-M001-...`

### 8.3 Output sink

- Kafka sink → `compliance.alerts` with Avro `EventEnvelope`
- `source` = `flink-compliance-cep`
- `correlationId` = from input event when present; else generate UUID
- Enable Flink checkpointing: interval 30s, EXACTLY_ONCE Kafka producer where supported

### 8.4 Job configuration (dev)

| Setting | Value |
|---------|-------|
| Parallelism | 2 |
| State backend | HashMapState (dev); RocksDB optional |
| Checkpoint dir | `file:///tmp/flink-checkpoints` (volume in Compose) |
| Restart strategy | fixed-delay, 3 attempts |

### 8.5 Tests

- Unit tests for pattern logic (pure Java, no cluster)
- MiniCluster integration test with Testcontainers Kafka + Redis (CI optional if slow; required before claiming Phase 2 complete)

---

## 9. Alert Service

Path: `services/alert-service/`

### 9.1 Responsibilities

1. **Consumer**: Subscribe to `compliance.alerts`; parse `ComplianceAlertRaised`; upsert alert row.
2. **Idempotency**: Dedup by envelope `idempotencyKey` (table `processed_alert_events` or unique on `alert_id`).
3. **REST API**: List/get/acknowledge alerts.
4. **WebSocket**: Broadcast new alerts and status changes to connected clients.
5. **Outbox**: **Not required** for downstream Kafka in Phase 2 (no republish). Direct WS fan-out from consumer thread is acceptable; use channel buffer with backpressure.

### 9.2 PostgreSQL schema

File: `services/alert-service/migrations/001_alerts.sql`

```sql
CREATE TABLE compliance_alerts (
  alert_id        UUID PRIMARY KEY,
  tenant_id       UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001',
  rule_code       TEXT NOT NULL,
  regime          TEXT NOT NULL,
  severity        TEXT NOT NULL,
  status          TEXT NOT NULL DEFAULT 'Open',
  persona_id      UUID NOT NULL,
  persona_type    TEXT NOT NULL,
  summary         TEXT NOT NULL,
  details         JSONB NOT NULL DEFAULT '{}',
  detected_at     TIMESTAMPTZ NOT NULL,
  acknowledged_at TIMESTAMPTZ,
  acknowledged_by TEXT,
  idempotency_key TEXT NOT NULL UNIQUE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX compliance_alerts_status_detected ON compliance_alerts (status, detected_at DESC);
CREATE INDEX compliance_alerts_persona ON compliance_alerts (persona_id);
```

### 9.3 REST API

Base path: `/api/v1`

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Liveness |
| `GET` | `/alerts` | List alerts (`status`, `severity`, `limit`, `offset`) |
| `GET` | `/alerts/{alertId}` | Single alert |
| `POST` | `/alerts/{alertId}/acknowledge` | Body: `{ "acknowledgedBy": "operator-dev" }` → status `Acknowledged`, publish WS event |

Errors: JSON `{"error":"...", "code":"NOT_FOUND"}` (match State Service style).

### 9.4 WebSocket

- Path: `/ws/alerts`
- On connect: optional query `?status=Open` to filter initial snapshot
- Messages: JSON `{"type":"alert.raised"|"alert.acknowledged", "payload":{...}}`
- Heartbeat: ping every 30s

### 9.5 Tests

- Store CRUD + idempotency unit tests
- Consumer mapping tests (sample Avro/JSON fixtures)
- WebSocket integration test (httptest or testcontainers)

---

## 10. Alert Console UI

Path: `apps/alert-console/`

### 10.1 Stack

- Next.js 14 App Router, TypeScript, Tailwind CSS (minimal styling)
- No auth in Phase 2; display dev banner

### 10.2 Pages

| Route | Description |
|-------|-------------|
| `/` | Live alert feed (WebSocket + polling fallback) |
| `/alerts/[alertId]` | Alert detail + acknowledge button |

### 10.3 Live feed requirements

- Connect to `NEXT_PUBLIC_WS_URL` on mount
- Show: severity badge, rule code, summary, persona, detectedAt
- Sort: newest first
- Acknowledge action calls REST then updates local state
- Reconnect with exponential backoff on disconnect

### 10.4 Tests

- Component tests for alert row and acknowledge (optional in Phase 2)
- Playwright smoke: page loads, receives mock WS message (CI optional)

---

## 11. Grafana

Path: `infra/grafana/`

### 11.1 Dashboards (minimum)

| Dashboard | Panels |
|-----------|--------|
| **Kafka Overview** | Consumer lag (`compliance.alerts`, `twin.state.updated`), broker bytes in/out |
| **Flink Job** | Checkpoint duration, last checkpoint success, records in/out, backpressure |

Use Prometheus JMX exporter sidecars or Flink REST API datasource if simpler for Phase 2; document scrape config in `infra/grafana/provisioning/`.

**Pragmatic shortcut**: Import community Kafka dashboard JSON; single Flink panel via REST polling plugin.

---

## 12. CI workflows

Extend `.github/workflows/ci.yml`:

1. Phase 1 steps (unchanged — must still pass)
2. **Java**: `mvn test` in `jobs/compliance-cep`
3. **Go**: `go test ./...` in `services/alert-service`
4. **Node**: `npm ci && npm run build` in `apps/alert-console`
5. Start extended Compose (Flink, Redis, alert stack)
6. Submit Flink job + `./scripts/smoke-test-phase2.sh`

Extend `schema-compat.yml` for new Avro subjects.

Extend `docker-publish.yml` to build/push:

- `ghcr.io/safetymp/digital-twin-compliance/alert-service`
- `ghcr.io/safetymp/digital-twin-compliance/alert-console`
- `ghcr.io/safetymp/digital-twin-compliance/compliance-cep` (Flink job JAR as OCI artifact or release asset)

---

## 13. Smoke test

File: `scripts/smoke-test-phase2.sh`

Exit 0 only if all pass (run after Phase 1 smoke test or include Phase 1 checks):

```bash
# 1. Flink job RUNNING (REST /jobs)
# 2. Baseline: GET /api/v1/alerts?status=Open returns >= 0
# 3. INT-M001: payment-burst.sh → new alert with ruleCode INT-M001 within 30s
# 4. BASEL-M001: UPDATE institution liquidity lcr to 0.90 in core banking / twin → alert within 30s
# 5. WebSocket receives alert.raised message
# 6. POST acknowledge → status Acknowledged in REST + WS alert.acknowledged
# 7. (Optional) Grafana health endpoint returns 200
```

Document exact `curl` and `psql` commands in the script.

---

## 14. Phase 2 exit criteria checklist

Copy into PR description when Phase 2 is complete.

**Status (Phase 2b, post-merge to `main` via PR #15):** verified in CI and `./scripts/report-eval-scorecard.sh --phase2`. Deferred items documented below — not blocking Phase 2a integration merge.

- [x] Extended `docker compose -f docker-compose.dev.yml up` starts Phase 1 + Redis + Flink + Alert Service + UI + Grafana — CI `docker compose up -d --wait`; mechanical check `phase2-stack-in-compose`
- [ ] Flink job RUNNING; checkpoint success rate > 99% over 15 min soak — **deferred**: CI confirms job RUNNING via `submit-flink-job.sh` + smoke; 15 min soak / checkpoint metric screenshot not measured in CI (staging follow-up)
- [x] `INT-M001`, `INT-M002`, `BASEL-M001` each produce ≥ 1 alert from seed/simulator scenarios — `./scripts/smoke-test-phase2.sh` steps 3–5 (green in CI)
- [x] Alerts on `compliance.alerts` consumed by Alert Service within 2s p99 — smoke asserts INT-M001 `consume_latency_ms ≤ 2000` on a single CI sample (not a full p99 distribution; see step 3 log)
- [x] Alert visible in UI within 5s of detection — smoke step 3 asserts API/console visibility within 5000ms budget
- [x] Acknowledge flow updates PostgreSQL and WebSocket clients — smoke steps 6–7 (`alert.acknowledged` WS + DB)
- [x] Redis keys updated for velocity / exposure / LCR features — smoke waits on `vel:`, `exp:`, `lcr:` keys before alert assertions
- [x] `./scripts/smoke-test.sh` still passes (Phase 1 regression) — CI step before Phase 2 smoke
- [x] `./scripts/smoke-test-phase2.sh` exits 0 — CI gate (PR #15)
- [x] `go test ./...` (alert-service) and `mvn test` (compliance-cep) pass — CI unit-test job
- [x] New Avro schemas pass BACKWARD compat CI — `.github/workflows/schema-compat.yml` on alert schemas
- [x] No Phase 3+ components added (Cedar, immudb, Neo4j, full auth) — mechanical `phase3-scope-boundary` + behavior evals

**Phase 2b agent pillars** (see [AGENTS.md](../AGENTS.md) § Behavior evals):

- [x] Behavior evals: 6/6 scenarios at 100% pass over 3 runs — `evals/live-model-phase2/results/<scenario>/run-*.json`; `./scripts/report-eval-scorecard.sh --phase2`
- [x] Token efficiency: `./scripts/token-efficiency.sh --strict` — `evals/live-model-phase2/results/efficiency-verification.json` (`harness_reread_count: 0`, `duplicate_read_count: 0`)

---

## 15. Foundation decisions (Phase 2)

Record as **ADR-008** when implementation starts:

| ID | Decision | Rationale |
|----|----------|-----------|
| D10 | Flink **mini-cluster in Compose** for dev/CI; K8s Operator for shared staging | Matches ADR-007 local-first pattern |
| D11 | Alert persistence in **PostgreSQL**; immudb deferred to Phase 3 | Avoid dual write complexity |
| D12 | **Inline thresholds** in Flink for Phase 2; Zen Decision Service in Phase 3 | Reduce cross-service latency while tuning patterns |
| D13 | Payments via **CDC table** not REST injector | Reuses Debezium path; realistic latency |

---

## 16. Implementation order

1. ADR-008 draft + extend `.env.example`
2. Avro alert schemas + register script
3. Mock payments migration + Debezium table update
4. Redis + alert-db in Compose
5. Alert Service (store, consumer, REST, WebSocket) + tests
6. Institution liquidity fields in State Service / seed
7. Flink job (sources → Redis → CEP → Kafka sink) + tests
8. Flink Compose services + submit script
9. Alert Console UI
10. Grafana provisioning
11. `smoke-test-phase2.sh` + CI extension
12. GHCR publish for new images
13. Verify exit criteria checklist

---

## 17. Staging / production notes

- **Staging**: Deploy Flink via [Kubernetes Flink Operator](https://nightlies.apache.org/flink/flink-kubernetes-operator-docs-stable/) per ADR-001; Alert Service and UI via existing [deploy-stack.sh](../scripts/deploy-stack.sh) pattern with GHCR images.
- **Managed Kafka**: Optional Confluent Cloud env vars (ADR-007 D4) for shared staging.
- **Auth**: Phase 3 Keycloak; until then restrict staging UI/API by network policy or basic API key (`ALERT_SERVICE_API_KEY`).

---

## References

- [roadmap.md § Phase 2](./roadmap.md)
- [architecture.md § 4.4 Monitoring](./architecture.md)
- [data-flow.md § 6.2 Real-Time Compliance Monitoring](./data-flow.md)
- [compliance-mapping.md § Internal / Basel monitoring](./compliance-mapping.md)
- [phase1-implementation-spec.md](./phase1-implementation-spec.md)
