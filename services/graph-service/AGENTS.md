# Graph Service — Agent Contract

Go REST API and Kafka consumer materializing exposure graph in Neo4j (Phase 4).

Parent contract: [AGENTS.md](../../AGENTS.md). Twin payloads: [contracts/kafka/twin.state.updated/](../../contracts/kafka/twin.state.updated/).

## Package map

| Path | Role |
|------|------|
| `cmd/server/` | HTTP server + Kafka consumer wiring |
| `internal/api/` | Graph query REST handlers |
| `internal/consumer/` | `twin.state.updated` → Neo4j upserts |
| `internal/graph/` | Neo4j driver store |
| `internal/config/` | Environment configuration |
| `internal/events/` | Envelope and twin payload parsing |

## Commands

```bash
cd services/graph-service && go test ./...
```

**Verification floor:** any edit requires `go test ./...` before claiming done.

## API

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/health` | Liveness + Neo4j connectivity |
| GET | `/api/v1/graph/summary` | Node and edge counts |
| GET | `/api/v1/graph/nodes?name=` | Institution nodes |
| GET | `/api/v1/graph/edges?layer=` | Exposure edges |

## Invariants

- Consumer group `graph-service` on `twin.state.updated`
- Instruments consumer on `domain.events.public.instruments` (Debezium CDC)
- Poison / parse failures publish to DLQ topics (`TWIN_STATE_DLQ_TOPIC`, `DOMAIN_INSTRUMENTS_DLQ_TOPIC`) — defaults `twin.state.updated.dlq`, `domain.events.public.instruments.dlq`
- Idempotent upserts keyed by `tenantId + entityId` / `edgeKey`
- Default tenant `00000000-0000-0000-0000-000000000001`
- Transient Neo4j/store errors: do not commit (retry). Poison parse errors: DLQ then commit (`twin.state.updated.dlq`, `domain.events.public.instruments.dlq`)
- CET1 comes from twin `capital.cet1` / `cet1_ratio` (never invent 0.12)
- Browser UIs proxy via Next.js — no CORS on `:8093`

## Gotchas

- Twin envelope `payload` is a **JSON string** (state-service outbox pattern)
- Restart graph-service after seed/outbox drain if smoke counts lag
- `./scripts/wait-graph-seeded.sh` polls summary until ≥10 nodes, ≥50 edges
- Neo4j auth from `NEO4J_URI`, `NEO4J_USER`, `NEO4J_PASSWORD`
