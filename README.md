# Digital Twin Compliance Platform

Event-driven financial digital twin with embedded compliance monitoring. Phase 1 delivers the **ingestion backbone and minimal twin**: Debezium CDC from a mock core-banking database, Kafka, a Go State Service with transactional outbox, and a read-only persona REST API.

**Author:** [SafetyMP](https://github.com/SafetyMP)  
**License:** [Apache License 2.0](LICENSE)

## Status

| Phase | Scope | Status |
|-------|--------|--------|
| **Phase 1** | Kafka, Debezium CDC, State Service, outbox, persona API, schema CI | Complete |
| Phase 2+ | Flink CEP, alerts UI, Cedar, immudb, graph, auth | Not started |

See [docs/roadmap.md](docs/roadmap.md) for the full implementation plan.

## Architecture (Phase 1)

```text
mock core-banking (PostgreSQL)
        │  CDC (Debezium)
        ▼
   Kafka (domain.events.*)
        │
        ▼
  State Service (Go)
   ├── consumer → twin store (PostgreSQL)
   ├── outbox → twin.state.updated
   └── REST API (/api/v1/personas)
```

Details: [docs/architecture.md](docs/architecture.md) · [docs/data-flow.md](docs/data-flow.md) · [docs/adr/](docs/adr/)

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and Docker Compose v2
- [Go 1.22+](https://go.dev/dl/)
- CLI tools: `curl`, `jq`, `psql` (PostgreSQL client)

## Quick start

```bash
git clone https://github.com/SafetyMP/Digital-Twin-Compliance.git
cd Digital-Twin-Compliance

cp .env.example .env

docker compose -f docker-compose.dev.yml up -d --wait
./scripts/seed.sh
./scripts/register-schemas.sh
./scripts/register-debezium-connector.sh

# Restart state-service after connector registration (CI does this automatically)
docker compose -f docker-compose.dev.yml restart state-service
docker compose -f docker-compose.dev.yml up -d --wait state-service

./scripts/smoke-test.sh
```

### Verify

```bash
cd services/state-service && go test ./...
curl -s http://localhost:8080/api/v1/health
curl -s "http://localhost:8080/api/v1/personas?personaType=Institution&limit=5" | jq
```

### Tear down

```bash
docker compose -f docker-compose.dev.yml down -v
```

## REST API (Phase 1)

Base URL: `http://localhost:8080/api/v1`

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Liveness probe |
| `GET` | `/personas` | List twin personas (`personaType`, `limit`, `offset`) |
| `GET` | `/personas/{personaId}` | Single persona by ID |

Persona types: `Institution`, `Account`, `Instrument`.

## Repository layout

| Path | Purpose |
|------|---------|
| [services/state-service/](services/state-service/) | Go REST API, Kafka consumer, transactional outbox |
| [schemas/avro/](schemas/avro/) | Avro event schemas (Schema Registry) |
| [mocks/core-banking/](mocks/core-banking/) | CDC source database migrations and seed data |
| [scripts/](scripts/) | Seed, smoke test, schema and connector registration |
| [docs/](docs/) | Architecture, domain model, ADRs, Phase 1 spec |
| [.github/workflows/](.github/workflows/) | CI and Avro schema compatibility checks |

## Documentation

| Document | Description |
|----------|-------------|
| [docs/phase1-implementation-spec.md](docs/phase1-implementation-spec.md) | Executable Phase 1 specification |
| [docs/architecture.md](docs/architecture.md) | C4 diagrams and component map |
| [docs/domain-model.md](docs/domain-model.md) | Entities, personas, glossary |
| [docs/data-flow.md](docs/data-flow.md) | Event envelopes, idempotency, topics |
| [docs/roadmap.md](docs/roadmap.md) | Phased delivery plan |
| [docs/adr/](docs/adr/) | Architecture decision records |
| [AGENTS.md](AGENTS.md) | Coding-agent contract (commands, scope, definition of done) |
| [CONTRIBUTING.md](CONTRIBUTING.md) | How to contribute |

## CI

GitHub Actions on every push and pull request:

1. Mechanical live evals (`./scripts/run-live-evals.sh`)
2. Docker Compose stack + seed + schema registration + Debezium connector
3. `go test ./...` in `services/state-service`
4. `./scripts/smoke-test.sh`
5. Avro BACKWARD compatibility check (separate workflow)

## Security

Phase 1 is **local development only**. There is no authentication middleware. Do not expose the default stack to untrusted networks. See [SECURITY.md](SECURITY.md).

## Contributing

Contributions are welcome. Read [CONTRIBUTING.md](CONTRIBUTING.md) before opening a pull request.
