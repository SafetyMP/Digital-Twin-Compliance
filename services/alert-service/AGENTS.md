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
go test ./... -cover
go run ./cmd/server
```

**Verification floor (not waivable):** any edit under `services/alert-service/` — including comment- or doc-only edits — requires at minimum `go build ./...` or `go test ./internal/<touched-pkg>/...` before claiming done. "Trivial" scopes how much to run, never whether to run. Logic changes also need the relevant `./scripts/smoke-test-phase2.sh` path.

## Test expectations

| Package | Minimum |
|---------|---------|
| `internal/api` | Health, list/get/ack routes, error mapping |
| `internal/config` | Env defaults and overrides |
| `internal/consumer` | Envelope handling and broadcast on create |
| `internal/events` | Envelope and alert payload parsing |
| `internal/hub` | Broadcast with zero clients |
| `internal/store` | Upsert idempotency and acknowledge (integration) |
| `cmd/server` | Migration SQL path resolution |

Repo gate: `./scripts/run-live-evals-phase2.sh --full` requires **≥25%** total coverage and package `_test.go` files in every package above.

From repo root: see [AGENTS.md](../../AGENTS.md) and [docs/phase2-implementation-spec.md](../../docs/phase2-implementation-spec.md).

## Invariants

- **Dedup by `idempotencyKey`** — envelope key is unique in `compliance_alerts.idempotency_key`
- **`tenant_id` on all rows** — default `00000000-0000-0000-0000-000000000001`
- **No outbox in Phase 2** — WebSocket fan-out directly from consumer; no Kafka republish
- **Poison messages** — unparseable payloads route to `compliance.alerts.dlq` (env `COMPLIANCE_ALERTS_DLQ_TOPIC`); offset committed only after DLQ write succeeds
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
