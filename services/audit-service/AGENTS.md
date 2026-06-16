# Audit Service — Agent Contract

Go Kafka consumer, immudb ledger writer, and REST API for Phase 3 tamper-evident audit trail.

Parent contract: [AGENTS.md](../../AGENTS.md). Audit envelope and hash chain: [docs/data-flow.md](../../docs/data-flow.md) §5.

## Package map

| Path | Role |
|------|------|
| `cmd/server/` | HTTP server, Kafka consumer, immudb wiring |
| `internal/api/` | REST handlers (health, search, verify) |
| `internal/consumer/` | `compliance.audit.pending` → immudb + `compliance.audit.recorded` |
| `internal/store/` | PostgreSQL idempotency, entry index, DLQ |
| `internal/chain/` | Canonical JSON hashing and chain verification |
| `internal/immudb/` | immudb gRPC client wrapper |
| `internal/config/` | Environment configuration |
| `internal/events/` | EventEnvelope, AuditPending, AuditEntry types |

## Commands

From `services/audit-service/`:

```bash
go test ./...
go run ./cmd/server
```

**Verification floor:** any edit requires `go test ./...` or `go build ./...` before claiming done.

## Invariants

- **Dedup by `idempotencyKey`** — envelope key stored in `audit_idempotency_keys`
- **Sole immudb writer** — only audit-service appends ledger entries (ADR-009 D16)
- **Hash chain** — `payloadHash = SHA-256(canonical JSON payload + metadata)`; `previousHash` links to prior entry
- **Poison messages** — unparseable payloads route to Kafka DLQ + `audit_outbox_dlq`; offset committed after DLQ
- **REST error shape** — `{"error":"...", "code":"NOT_FOUND"}` matching alert-service

## Key files

- `internal/chain/chain.go` — hash and verify logic
- `internal/consumer/handler.go` — pending → recorded pipeline
- `migrations/001_audit.sql` — idempotency, index, DLQ tables

## Gotchas

- Consumer expects JSON `EventEnvelope` with `eventType=AuditPending`
- Default HTTP port `8090`; health at `/api/v1/health` includes immudb ping
- immudb database `digitaltwin_audit` is auto-created on first connect if missing
- Verify API loads full entries from immudb ordered by PostgreSQL `sequence_number`
