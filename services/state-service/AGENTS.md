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
go run ./cmd/server
```

From repo root: see [AGENTS.md](../../AGENTS.md) for Compose, seed, schema registration, and smoke test.

## Invariants

- **Outbox-only Kafka writer** — only `internal/outbox/` may use `kafka.Writer`; never publish directly from consumer or store
- **`tenant_id` on all entity tables** — default `00000000-0000-0000-0000-000000000001`
- **Hierarchy depth ≤ 3** — parent → subsidiary → sub-subsidiary (ADR-007 D9)

## Key files

- `internal/outbox/publisher.go` — durable publish path
- `internal/consumer/mapper.go` — CDC row → twin entity mapping
- `migrations/001_init.sql` — state store schema

## Gotchas

- Run `./scripts/register-schemas.sh` before starting the consumer (Schema Registry must have Avro subjects)
- Run `./scripts/register-debezium-connector.sh` after Compose stack is healthy
- `./scripts/smoke-test.sh` requires the full Compose stack up (`docker compose -f docker-compose.dev.yml up -d --wait`)
- Default tenant UUID: `00000000-0000-0000-0000-000000000001` — use consistently in seed, migrations, and tests
