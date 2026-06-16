# Phase 1 Implementation Spec

Executable handoff for the parallel implementation agent. Implements [roadmap.md](./roadmap.md) Phase 1 only.

**Prerequisites**: [ADR-007](./adr/007-phase1-foundation-decisions.md) (D1, D4, D9 decided).

**Related docs**: [architecture.md](./architecture.md), [domain-model.md](./domain-model.md), [data-flow.md](./data-flow.md), [ADR-001](./adr/001-kafka-flink-streaming.md), [ADR-004](./adr/004-datastore-selection.md).

---

## 1. Goal

Build the ingestion backbone and minimal twin so that:

1. A change in the mock core-banking PostgreSQL appears on `domain.events` within 5 seconds.
2. The State Service upserts twin entities and publishes `twin.state.updated`.
3. REST API returns seeded personas, accounts, and instruments.
4. CI rejects incompatible Avro schema changes.

---

## 2. Repository layout

Create this structure (no code in docs repo yet — implementation agent creates files):

```
/
├── AGENTS.md
├── docker-compose.dev.yml
├── .env.example
├── schemas/
│   └── avro/
│       ├── event-envelope.avsc
│       ├── entity-state-updated.avsc
│       └── twin-state-updated.avsc
├── services/
│   └── state-service/
│       ├── go.mod
│       ├── cmd/server/main.go
│       ├── internal/
│       │   ├── api/          # HTTP handlers
│       │   ├── consumer/     # domain.events consumer
│       │   ├── outbox/       # outbox publisher
│       │   ├── store/        # PostgreSQL repository
│       │   └── config/
│       └── migrations/
│           └── 001_init.sql
├── mocks/
│   └── core-banking/
│       ├── migrations/
│       │   └── 001_source_tables.sql
│       └── seed/
│           └── seed.sql
├── scripts/
│   ├── seed.sh
│   ├── smoke-test.sh
│   └── register-schemas.sh
└── .github/
    └── workflows/
        ├── ci.yml
        └── schema-compat.yml
```

**Go module path**: `github.com/digital-twin/platform/services/state-service` (adjust org if repo remote differs).

---

## 3. Docker Compose (local dev)

File: `docker-compose.dev.yml`

| Service | Image / build | Ports | Purpose |
|---------|---------------|-------|---------|
| `core-banking-db` | postgres:16 | 5433:5432 | CDC source (mock core banking) |
| `state-db` | postgres:16 | 5434:5432 | Twin entity store + outbox |
| `kafka` | apache/kafka:3.7.0 (KRaft) | 9092 | Event backbone (single broker dev) |
| `schema-registry` | confluentinc/cp-schema-registry:7.6.0 | 8081 | Avro schemas |
| `debezium` | debezium/connect:2.6 | 8083 | PostgreSQL CDC → Kafka |
| `state-service` | build `./services/state-service` | 8080 | REST + consumer + outbox |

### Environment variables (`.env.example`)

```bash
# Core banking (CDC source)
CORE_BANKING_DB_URL=postgres://core:core@localhost:5433/core_banking?sslmode=disable

# State store
STATE_DB_URL=postgres://state:state@localhost:5434/twin_state?sslmode=disable

# Kafka
KAFKA_BROKERS=localhost:9092
SCHEMA_REGISTRY_URL=http://localhost:8081

# Default tenant (ADR-007 D1)
DEFAULT_TENANT_ID=00000000-0000-0000-0000-000000000001

# State Service
STATE_SERVICE_HTTP_ADDR=:8080
STATE_SERVICE_SOURCE=state-service
```

### Debezium connector config

Register after Compose is up (via `scripts/register-schemas.sh` or REST):

- **Connector name**: `core-banking-cdc`
- **Source DB**: `core-banking-db`, database `core_banking`, publication `dbz_publication`
- **Tables**: `legal_entities`, `accounts`, `instruments` (see Section 6)
- **Topic prefix**: `domain.events` (map to single topic `domain.events` via SMT or `topic.creation.default.replication.factor`)
- **Key converter**: Avro via Schema Registry
- **Value converter**: Avro via Schema Registry
- **Transforms**: unwrap Debezium envelope; wrap in platform event envelope (custom SMT or State Service normalizes on consume)

**Pragmatic Phase 1 shortcut**: Debezium publishes to `domain.events` with a simplified JSON/Avro payload; State Service maps CDC rows to `EntityStateUpdated` internally. Full envelope at CDC boundary is Phase 2 polish.

### Kafka topics (Phase 1 minimum)

| Topic | Partitions (dev) | Created by |
|-------|------------------|------------|
| `domain.events` | 3 | Debezium / init script |
| `domain.events.dlq` | 1 | init script |
| `twin.state.updated` | 3 | State Service (auto-create or init script) |

Retention in dev: 7 days (not 7 years — production config in ADR-001).

---

## 4. PostgreSQL schemas

### 4.1 Mock core banking (`mocks/core-banking/migrations/001_source_tables.sql`)

Source-of-truth tables mirrored from [domain-model.md](./domain-model.md):

```sql
CREATE TABLE legal_entities (
  entity_id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  legal_name      TEXT NOT NULL,
  lei             TEXT,
  entity_type     TEXT NOT NULL CHECK (entity_type IN ('Bank','Fund','SPV','CCP','ICTProvider','InternalUnit')),
  jurisdiction    CHAR(2) NOT NULL,
  parent_entity_id UUID REFERENCES legal_entities(entity_id),
  consolidation_scope TEXT NOT NULL DEFAULT 'Solo',
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE accounts (
  account_id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  account_number  TEXT NOT NULL,
  account_type    TEXT NOT NULL,
  currency        CHAR(3) NOT NULL,
  owner_entity_id UUID NOT NULL REFERENCES legal_entities(entity_id),
  status          TEXT NOT NULL DEFAULT 'Active',
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE instruments (
  instrument_id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  isin            TEXT,
  instrument_type TEXT NOT NULL,
  counterparty_id UUID,
  notional_amount NUMERIC(20,2) NOT NULL,
  currency        CHAR(3) NOT NULL,
  maturity_date   DATE,
  regulatory_class TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

Enable logical replication for Debezium:

```sql
ALTER TABLE legal_entities REPLICA IDENTITY FULL;
ALTER TABLE accounts REPLICA IDENTITY FULL;
ALTER TABLE instruments REPLICA IDENTITY FULL;
```

**Hierarchy constraint (ADR-007 D9)**: Enforce max depth 3 in State Service when upserting personas (validate `parent_entity_id` chain length ≤ 2 hops from root).

### 4.2 State store (`services/state-service/migrations/001_init.sql`)

```sql
CREATE TABLE twin_personas (
  persona_id       UUID PRIMARY KEY,
  tenant_id        UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001',
  source_entity_id UUID NOT NULL,
  persona_type     TEXT NOT NULL,
  state_version    INT NOT NULL DEFAULT 1,
  current_state    JSONB NOT NULL DEFAULT '{}',
  compliance_status TEXT NOT NULL DEFAULT 'Unknown',
  last_synced_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, source_entity_id, persona_type)
);

CREATE TABLE accounts (
  account_id       UUID PRIMARY KEY,
  tenant_id        UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001',
  account_number   TEXT NOT NULL,
  account_type     TEXT NOT NULL,
  currency         CHAR(3) NOT NULL,
  owner_entity_id  UUID NOT NULL,
  status           TEXT NOT NULL,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE instruments (
  instrument_id    UUID PRIMARY KEY,
  tenant_id        UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001',
  isin             TEXT,
  instrument_type  TEXT NOT NULL,
  notional_amount  NUMERIC(20,2) NOT NULL,
  currency         CHAR(3) NOT NULL,
  maturity_date    DATE,
  regulatory_class TEXT,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE processed_events (
  idempotency_key  TEXT PRIMARY KEY,
  processed_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE outbox (
  id               BIGSERIAL PRIMARY KEY,
  topic            TEXT NOT NULL,
  partition_key    TEXT NOT NULL,
  payload          JSONB NOT NULL,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  published_at     TIMESTAMPTZ
);

CREATE INDEX outbox_unpublished ON outbox (id) WHERE published_at IS NULL;
```

---

## 5. Avro schemas

Directory: `schemas/avro/`. Register via Schema Registry on startup (`scripts/register-schemas.sh`).

Compatibility mode: **BACKWARD** ([data-flow.md](./data-flow.md) Section 2.3).

### 5.1 `event-envelope.avsc`

Subject: `domain.events-value` (and reused for `twin.state.updated-value`)

```json
{
  "type": "record",
  "name": "EventEnvelope",
  "namespace": "com.digitaltwin.events",
  "fields": [
    {"name": "eventId", "type": "string"},
    {"name": "eventType", "type": "string"},
    {"name": "eventVersion", "type": "string"},
    {"name": "source", "type": "string"},
    {"name": "correlationId", "type": "string"},
    {"name": "causationId", "type": ["null", "string"], "default": null},
    {"name": "timestamp", "type": "string"},
    {"name": "idempotencyKey", "type": "string"},
    {"name": "payload", "type": "string"}
  ]
}
```

Phase 1: `payload` is JSON string (Avro union with embedded record in Phase 2).

### 5.2 `entity-state-updated.avsc`

Documented payload shape (embedded in envelope `payload` JSON):

| Field | Type | Required |
|-------|------|----------|
| `personaId` | UUID string | Yes |
| `personaType` | enum string | Yes |
| `sourceEntityId` | UUID string | Yes |
| `stateVersion` | int | Yes |
| `changedFields` | string[] | Yes |
| `currentState` | object | Yes |
| `sourceSystem` | string | Yes |
| `sourceTimestamp` | ISO 8601 | Yes |

See [data-flow.md](./data-flow.md) Section 3.1 for example JSON.

---

## 6. Seed data

File: `mocks/core-banking/seed/seed.sql`

**Targets** (from roadmap Phase 1):

| Entity | Count | Notes |
|--------|-------|-------|
| Institutions (`legal_entities`) | 10 | 3 parent groups, subsidiaries, 1–2 sub-subsidiaries each |
| Accounts | 100 | ~10 per institution |
| Instruments | 500 | Mix of Loan, Bond, Deposit |

**Hierarchy example** (3 levels):

```
GroupA (parent)
├── BankA1 (subsidiary)
│   └── BranchA1a (sub-subsidiary)
└── BankA2 (subsidiary)
GroupB ...
GroupC ...
```

Script: `scripts/seed.sh` applies migrations + seed to both databases, then triggers initial CDC snapshot.

---

## 7. State Service

### 7.1 Responsibilities

1. **Consumer**: Subscribe to `domain.events`; parse CDC or `EntityStateUpdated`; upsert `twin_personas`, `accounts`, `instruments`.
2. **Idempotency**: Check `processed_events` by `idempotencyKey` before write.
3. **Outbox**: Insert row into `outbox`; background worker publishes to `twin.state.updated` with `EventEnvelope`.
4. **REST API**: Read-only queries for Phase 1.

### 7.2 REST API

Base path: `/api/v1`

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Liveness (`{"status":"ok"}`) |
| `GET` | `/personas/{personaId}` | Single twin persona |
| `GET` | `/personas` | List personas |

**Query params for `GET /personas`**:

| Param | Type | Description |
|-------|------|-------------|
| `personaType` | string | Filter: `Institution`, `Account`, `Instrument` |
| `limit` | int | Default 50, max 200 |
| `offset` | int | Pagination |

**Response shape** (`TwinPersona`):

```json
{
  "personaId": "uuid",
  "sourceEntityId": "uuid",
  "personaType": "Institution",
  "stateVersion": 1,
  "currentState": {},
  "complianceStatus": "Unknown",
  "lastSyncedAt": "2026-06-13T18:45:00Z"
}
```

Errors: JSON `{"error":"...", "code":"NOT_FOUND"}` with appropriate HTTP status.

### 7.3 Consumer mapping (CDC → persona)

| Source table | personaType | personaId source |
|--------------|-------------|------------------|
| `legal_entities` | `Institution` | `entity_id` |
| `accounts` | `Account` | `account_id` |
| `instruments` | `Instrument` | `instrument_id` |

On each CDC event:

1. Build `idempotencyKey` = `{table}-{pk}-{updated_at_epoch}`.
2. Upsert entity + increment `state_version`.
3. Insert outbox row for `twin.state.updated`.
4. Record idempotency key.

### 7.4 Outbox publisher

- Poll `outbox` every 1s (or LISTEN/NOTIFY).
- Publish Avro-encoded envelope to Kafka.
- Set `published_at` on success.
- Retry with exponential backoff on failure.

---

## 8. CI workflows

### 8.1 `.github/workflows/ci.yml`

On push/PR:

1. `docker compose -f docker-compose.dev.yml up -d --wait`
2. `scripts/seed.sh`
3. `cd services/state-service && go test ./...`
4. `scripts/smoke-test.sh`

### 8.2 `.github/workflows/schema-compat.yml`

1. Start Schema Registry (Compose service).
2. Register current schemas from `schemas/avro/`.
3. Run compatibility check against previous schema version (store previous in `schemas/avro/.baseline/` or fetch from Registry).
4. Fail PR if FORWARD/BACKWARD incompatible change detected.

---

## 9. Smoke test

File: `scripts/smoke-test.sh`

Exit 0 only if all pass:

```bash
#!/usr/bin/env bash
set -euo pipefail

BASE="${STATE_SERVICE_URL:-http://localhost:8080}"

# 1. Health
curl -sf "$BASE/api/v1/health" | grep -q ok

# 2. Seeded personas exist
COUNT=$(curl -sf "$BASE/api/v1/personas?personaType=Institution&limit=200" | jq '. | length')
[[ "$COUNT" -ge 10 ]]

# 3. CDC → Kafka → state: update one entity in core banking, wait ≤5s, state_version increments
#    (implementation: psql UPDATE + poll persona endpoint)

# 4. twin.state.updated message consumed (optional: kcat consume 1 message)
```

Document exact `psql` update and poll loop in the script.

---

## 10. Phase 1 exit criteria checklist

Copy into PR description when Phase 1 is complete:

- [ ] `docker compose -f docker-compose.dev.yml up` starts all services
- [ ] `scripts/seed.sh` loads 10 institutions, 100 accounts, 500 instruments
- [ ] `GET /api/v1/personas?personaType=Institution` returns ≥ 10 records
- [ ] UPDATE on core-banking `legal_entities` → `domain.events` within 5s
- [ ] State Service upserts persona; `state_version` increments
- [ ] Outbox publishes to `twin.state.updated`
- [ ] `scripts/smoke-test.sh` exits 0
- [ ] `go test ./...` passes in `services/state-service`
- [ ] Schema compat CI gate passes on PR
- [ ] No Phase 3+ components added ([AGENTS.md § Scope by phase](../AGENTS.md#scope-by-phase))

---

## 11. Optional: managed Kafka (staging)

For teams using Confluent Cloud instead of local Kafka (ADR-007 D4 staging path):

```bash
KAFKA_BROKERS=pkc-xxx.region.provider.confluent.cloud:9092
SCHEMA_REGISTRY_URL=https://xxx.region.provider.confluent.cloud
KAFKA_SASL_MECHANISM=PLAIN
# Credentials via env — never commit
```

Local dev should not require these variables.

---

## 12. Implementation order

1. Init git repo + directory scaffold
2. `docker-compose.dev.yml` + Postgres migrations
3. Seed data + `scripts/seed.sh`
4. Avro schemas + `register-schemas.sh`
5. State Service migrations + store layer
6. REST handlers
7. Kafka consumer + idempotency
8. Outbox publisher
9. Debezium connector registration
10. `smoke-test.sh` + CI workflows
11. Verify exit criteria checklist
