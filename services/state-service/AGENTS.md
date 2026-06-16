# State Service — Agent Contract

Go REST API, Kafka consumer, and transactional outbox for Phase 1 twin state.

Parent contract: [AGENTS.md](../../AGENTS.md). Envelope/idempotency details: [docs/data-flow.md](../../docs/data-flow.md).

## Package map

| Path | Role |
|------|------|
| `cmd/server/` | HTTP server, consumer, outbox publisher wiring |
| `internal/api/` | REST handlers |
| `internal/consumer/` | Debezium CDC → entity upsert mapping |
| `internal/outbox/` | Transactional outbox + Kafka publish |
| `internal/store/` | PostgreSQL entity store + hierarchy validation |
| `internal/config/` | Environment configuration |
| `internal/events/` | Event envelope helpers |

## Commands

From `services/state-service/`:

```bash
go test ./...
go test ./internal/store/...
go test ./internal/outbox/...
go test ./... -cover
go run ./cmd/server
```

## Test expectations

| Package | Minimum |
|---------|---------|
| `internal/api` | Handler routes, query parsing, error mapping |
| `internal/config` | Env defaults and overrides |
| `internal/events` | Envelope builder and idempotency fields |
| `internal/consumer` | CDC mapping (existing) |
| `internal/outbox` | Envelope/idempotency from outbox rows |
| `internal/store` | Field helpers, hierarchy, store contracts |
| `cmd/server` | Migration SQL path resolution |

Repo gate: `./scripts/run-live-evals.sh --full` requires **≥35%** total coverage and package `_test.go` files in every package above.

## Context efficiency (state-service work)

- Load this file + the one package you are changing; do not preload `docs/architecture.md` or `docs/roadmap.md`.
- Grep for symbols before opening `store.go` or `handlers.go` (large files).
- Never read `agent-transcripts/` to debug — reproduce with `go test` and smoke scripts.

From repo root: see [AGENTS.md](../../AGENTS.md) for Compose, seed, schema registration, and smoke test.

## Invariants

- **Outbox-only Kafka writer** — only `internal/outbox/` may use `kafka.Writer` for domain publish; never publish twin events directly from consumer or store (`internal/consumer/dlq.go` is allowed for poison routing only)
- **`tenant_id` on all entity tables** — default `00000000-0000-0000-0000-000000000001`
- **Hierarchy depth ≤ 3** — parent → subsidiary → sub-subsidiary (ADR-007 D9)
- **Poison CDC messages** — handler failures route to `domain.events.dlq` (env `STATE_CDC_DLQ_TOPIC`); offset committed only after DLQ write succeeds

## Key files

- `internal/outbox/publisher.go` — durable publish path
- `internal/consumer/mapper.go` — CDC row → twin entity mapping
- `migrations/001_init.sql` — state store schema

## Gotchas

- **CDC `updated_at`** — unparseable or missing `updated_at` fails mapping (DLQ path); no silent `time.Now()` fallback
- **Instrument numeric enrichment** — `enrichInstrumentState` in `internal/consumer/enrichment.go` converts CDC string decimals (e.g. `notional_amount`) to JSON numbers in `currentState` before outbox publish; golden publisher contract: `contracts/kafka/twin.state.updated/instrument.payload.json` (`TestKafkaContract_*` in `kafka_contract_test.go`)
- Run `./scripts/register-schemas.sh` before starting the consumer (Schema Registry must have Avro subjects)
- Run `./scripts/register-debezium-connector.sh` after Compose stack is healthy
- **Outbox publisher** — **at-least-once** to Kafka: batch `WriteMessages` then per-row `published_at` (crash between them redelivers). Batches up to `OUTBOX_BATCH_SIZE` (default 100) with `OUTBOX_BATCH_TIMEOUT` (default `10ms`); per-row writes with `kafka-go` defaults stall at ~1 msg/s. Downstream idempotency makes redelivery safe. After restart: `./scripts/wait-outbox-drained.sh` then `./scripts/verify-state-twin-pipeline.sh`.
- **`upsertLegalEntityMirror`** — intentionally no-op; institution CDC is stored in `twin_personas.current_state` only (no `legal_entities` mirror table in state DB). Accounts/instruments have normalized mirror tables.
- `./scripts/smoke-test.sh` requires the full Compose stack up (`docker compose -f docker-compose.dev.yml up -d --wait`)
- Default tenant UUID: `00000000-0000-0000-0000-000000000001` — use consistently in seed, migrations, and tests
