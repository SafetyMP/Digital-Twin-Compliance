# Cedar Policy Service — Agent Contract

Go REST API for Phase 3 Cedar policy evaluation and deny-path audit publishing.

Parent contract: [AGENTS.md](../../AGENTS.md). RuleDecision and audit envelope: [docs/data-flow.md](../../docs/data-flow.md) §5.

## Package map

| Path | Role |
|------|------|
| `cmd/server/` | HTTP server, policy loader, Kafka publisher wiring |
| `internal/api/` | REST handlers (`/evaluate`, `/health`) |
| `internal/engine/` | Policy loader, rule mapping, cedar-go evaluation |
| `internal/decision/` | RuleDecision type + input hashing |
| `internal/audit/` | `compliance.audit.pending` Kafka publisher |
| `internal/config/` | Environment configuration |

## Commands

From `services/cedar-service/`:

```bash
go test ./...
go run ./cmd/server
```

**Verification floor:** any edit requires `go test ./...` or `go build ./...` before claiming done.

## Invariants

- **Five Cedar rules** — `INT-R003`, `INT-R004`, `COREP-R005`, `EMIR-R004`, `DORA-R001` mapped to `policies/cedar/*.cedar`
- **Role membership** — policies use `principal in DigitalTwin::Role::"..."` (cedar-go compatible)
- **Deny audit** — `outcome=Deny` publishes `AuditPending` (`entryType=RuleDecision`) to `compliance.audit.pending`
- **Dev auth** — when `principal` omitted in body, use `X-Principal` + `X-Roles` headers
- **REST error shape** — `{"error":"...", "code":"BAD_REQUEST"}` matching alert-service

## Gotchas

- Policies load from `CEDAR_POLICY_DIR` (default `policies/cedar`); Compose mounts at `/policies/cedar`
- Health at `GET /api/v1/health` reports `schemaLoaded`, `policiesLoaded`, `ruleCodes`
- Default HTTP port `8091`
- Resource attrs accept JSON key `attrs` or `attributes`
