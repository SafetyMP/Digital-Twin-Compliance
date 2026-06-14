# Alert Service — Agent Contract

Go REST API, Kafka consumer, and WebSocket hub for Phase 2 compliance alerts.

Parent contract: [AGENTS.md](../../AGENTS.md). Alert envelope fields: [docs/data-flow.md](../../docs/data-flow.md) §4.2.

## Package map

| Path | Role |
|------|------|
| `cmd/server/` | HTTP server, Kafka consumer, WebSocket hub wiring |
| `internal/api/` | REST + WebSocket handlers |
| `internal/consumer/` | `compliance.alerts` → alert upsert |
| `internal/hub/` | WebSocket broadcast with backpressure channel |
| `internal/store/` | PostgreSQL alert store + idempotency |
| `internal/config/` | Environment configuration |
| `internal/events/` | Envelope and alert payload parsing |

## Commands

From `services/alert-service/`:

```bash
go test ./...
go run ./cmd/server
```

From repo root: see [AGENTS.md](../../AGENTS.md) and [docs/phase2-implementation-spec.md](../../docs/phase2-implementation-spec.md).

## Invariants

- **Dedup by `idempotencyKey`** — envelope key is unique in `compliance_alerts.idempotency_key`
- **`tenant_id` on all rows** — default `00000000-0000-0000-0000-000000000001`
- **No outbox in Phase 2** — WebSocket fan-out directly from consumer; no Kafka republish
- **REST error shape** — `{"error":"...", "code":"NOT_FOUND"}` matching State Service

## Key files

- `internal/store/store.go` — upsert + acknowledge
- `internal/consumer/consumer.go` — Kafka consumer group `alert-service`
- `internal/hub/hub.go` — WebSocket clients and broadcast
- `migrations/001_alerts.sql` — alert schema

## Gotchas

- Consumer expects JSON `EventEnvelope` with `eventType=ComplianceAlertRaised` (same as Phase 1 outbox pattern)
- Run `./scripts/register-schemas.sh` before starting consumer
- `./scripts/smoke-test-phase2.sh` requires full Phase 2 Compose stack including Flink job RUNNING
- WebSocket path: `/ws/alerts`; health at `/api/v1/health`
