# Digital Twin Compliance Platform

Event-driven financial digital twin with embedded compliance monitoring. Phase 1 delivers the **ingestion backbone and minimal twin**: Debezium CDC from a mock core-banking database, Kafka, a Go State Service with transactional outbox, and a read-only persona REST API.

**Author:** [SafetyMP](https://github.com/SafetyMP)  
**License:** [Apache License 2.0](LICENSE)

[![CI](https://github.com/SafetyMP/Digital-Twin-Compliance/actions/workflows/ci.yml/badge.svg)](https://github.com/SafetyMP/Digital-Twin-Compliance/actions/workflows/ci.yml)
[![Schema Compatibility](https://github.com/SafetyMP/Digital-Twin-Compliance/actions/workflows/schema-compat.yml/badge.svg)](https://github.com/SafetyMP/Digital-Twin-Compliance/actions/workflows/schema-compat.yml)
[![Docker Publish](https://github.com/SafetyMP/Digital-Twin-Compliance/actions/workflows/docker-publish.yml/badge.svg)](https://github.com/SafetyMP/Digital-Twin-Compliance/actions/workflows/docker-publish.yml)
[![CodeQL](https://github.com/SafetyMP/Digital-Twin-Compliance/actions/workflows/codeql.yml/badge.svg)](https://github.com/SafetyMP/Digital-Twin-Compliance/actions/workflows/codeql.yml)

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
| [docker-compose.deploy.yml](docker-compose.deploy.yml) | Staging deploy stack (GHCR image) |
| [.github/workflows/](.github/workflows/) | CI, Docker publish, release, staging deploy |

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
| [docs/deployment.md](docs/deployment.md) | GHCR, releases, staging deploy |
| [docs/github-setup.md](docs/github-setup.md) | Branch protection, environments, first release |

## DevOps and deployment

| Capability | Details |
|------------|---------|
| **CI** | Full stack tests on every push and PR |
| **Container registry** | State Service published to `ghcr.io/safetymp/digital-twin-compliance/state-service` on merge to `main` |
| **Releases** | Tag `v*.*.*` to publish versioned images and a GitHub Release |
| **Staging deploy** | Manual workflow — SSH deploy to a host running `docker-compose.deploy.yml` |
| **Dependabot** | Weekly updates for Go, GitHub Actions, and Docker |
| **CodeQL** | Go security analysis on push, PR, and weekly schedule |
| **Issue templates** | Structured bug reports and feature requests |

Quick deploy on a host with Docker:

```bash
export STATE_SERVICE_IMAGE=ghcr.io/safetymp/digital-twin-compliance/state-service:main
./scripts/deploy-stack.sh bootstrap
```

Full guide: [docs/deployment.md](docs/deployment.md).

## CI

GitHub Actions on every push and pull request:

1. Mechanical live evals (`./scripts/run-live-evals.sh`)
2. Docker Compose stack + seed + schema registration + Debezium connector
3. `go vet` and `go test ./...` in `services/state-service`
4. `./scripts/smoke-test.sh`
5. Avro BACKWARD compatibility check (separate workflow)
6. Docker image build and push to GHCR on merge to `main` (separate workflow)

## Security

Phase 1 is **local development only**. There is no authentication middleware. Do not expose the default stack to untrusted networks. See [SECURITY.md](SECURITY.md).

## Contributing

Contributions are welcome. Read [CONTRIBUTING.md](CONTRIBUTING.md) before opening a pull request.

This project follows the [Code of Conduct](CODE_OF_CONDUCT.md). GitHub setup (branch protection, environments, first release): [docs/github-setup.md](docs/github-setup.md).
