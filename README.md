# Digital Twin Compliance Platform

Event-driven financial digital twin with embedded compliance monitoring. **Phase 1** delivers the ingestion backbone and minimal twin (Debezium CDC, Kafka, Go State Service, transactional outbox, persona REST API). **Phase 2** adds real-time compliance monitoring (Flink CEP, Redis features, Alert Service, WebSocket, alert console, Grafana).

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
| **Phase 2** | Flink CEP, Redis, Alert Service, WebSocket, alert console, Grafana | Implemented on `main` — [mechanical smoke green](docs/review/phase2-exit-checklist.md); [behavior eval pillar](AGENTS.md#behavior-evals-phase-2) tracked separately |
| **Phase 3** | Cedar + Zen, immudb audit ledger, Audit Explorer, `evidenceRef` on alerts | Implemented (3a) — [`smoke-test-phase3.sh`](scripts/smoke-test-phase3.sh) + [spec §13](docs/phase3-implementation-spec.md#13-phase-3-exit-criteria-checklist) |

See [docs/roadmap.md](docs/roadmap.md) for the full plan.

## Architecture

```text
mock core-banking (PostgreSQL)
        │  CDC (Debezium) — personas + payments
        ▼
   Kafka (domain.events.*, twin.state.updated)
        │
        ├──────────────────────────────────┐
        ▼                                  ▼
  State Service (Go)              Flink CEP (Java)
   ├── consumer → twin store            ├── Redis (vel/exp/lcr features)
   ├── outbox → twin.state.updated     └── compliance.alerts
   └── REST /api/v1/personas                    │
                                                ▼
                                        Alert Service (Go)
                                         ├── PostgreSQL (alerts)
                                         ├── REST /api/v1/alerts
                                         └── WebSocket /ws/alerts
                                                │
                                                ▼
                                        Alert Console (Next.js)

Grafana ← Flink / Kafka metrics (Compose)
```

Details: [docs/architecture.md](docs/architecture.md) · [docs/data-flow.md](docs/data-flow.md) · [docs/adr/](docs/adr/)

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and Docker Compose v2
- [Go 1.22+](https://go.dev/dl/) (for local service tests)
- CLI tools: `curl`, `jq`, `psql` (PostgreSQL client)
- **Maven not required** — run Flink job unit tests via Docker (see [CONTRIBUTING.md](CONTRIBUTING.md))

## Quick start

### Phase 1 (ingestion + twin)

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

### Phase 2 (monitoring + alerts)

After Phase 1 smoke passes, ensure the Flink CEP job is **RUNNING** (Compose `flink-job-submitter` usually handles this; re-submit if needed):

```bash
./scripts/submit-flink-job.sh
./scripts/smoke-test-phase2.sh
```

**Gotchas**

- Redis host port is **6380** (Compose maps `6380:6379`); see `REDIS_URL` in `.env.example`.
- Kafka topics for Phase 2 are created by `./scripts/seed.sh` (via `create-kafka-topics.sh`).
- Phase 2 smoke injects payment bursts and checks INT-M001, INT-M002, and BASEL-M001 alerts end-to-end.

### Verify

```bash
# Phase 1
cd services/state-service && go test ./...
curl -s http://localhost:8080/api/v1/health
curl -s "http://localhost:8080/api/v1/personas?personaType=Institution&limit=5" | jq

# Phase 2
cd services/alert-service && go test ./...
curl -s http://localhost:8085/api/v1/health
curl -s "http://localhost:8085/api/v1/alerts?limit=5" | jq
open http://localhost:3000   # Alert Console
open http://localhost:8082   # Flink UI
open http://localhost:3001   # Grafana
```

### Local service URLs

| Service | URL |
|---------|-----|
| State Service | `http://localhost:8080` |
| Alert Service | `http://localhost:8085` |
| Alert Console | `http://localhost:3000` |
| Flink JobManager UI | `http://localhost:8082` |
| Grafana | `http://localhost:3001` |
| Redis | `localhost:6380` |
| Debezium Connect | `http://localhost:8083` |

### Tear down

```bash
docker compose -f docker-compose.dev.yml down -v
```

## REST API

### State Service (Phase 1)

Base URL: `http://localhost:8080/api/v1`

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Liveness probe |
| `GET` | `/personas` | List twin personas (`personaType`, `limit`, `offset`) |
| `GET` | `/personas/{personaId}` | Single persona by ID |

Persona types: `Institution`, `Account`, `Instrument`.

### Alert Service (Phase 2)

Base URL: `http://localhost:8085/api/v1`

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Liveness probe |
| `GET` | `/alerts` | List alerts (`limit`, `offset`, filters) |
| `GET` | `/alerts/{alertId}` | Single alert by ID |
| `POST` | `/alerts/{alertId}/acknowledge` | Acknowledge an alert |
| `GET` | `/ws/alerts` | WebSocket stream (see `NEXT_PUBLIC_WS_URL` in `.env.example`) |

Full contract: [docs/phase2-implementation-spec.md](docs/phase2-implementation-spec.md) · [services/alert-service/AGENTS.md](services/alert-service/AGENTS.md)

## Repository layout

| Path | Purpose |
|------|---------|
| [services/state-service/](services/state-service/) | Go REST API, Kafka consumer, transactional outbox |
| [services/alert-service/](services/alert-service/) | Alerts REST + WebSocket + `compliance.alerts` consumer |
| [jobs/compliance-cep/](jobs/compliance-cep/) | Flink CEP job (Java) — INT-M001, INT-M002, BASEL-M001 |
| [apps/alert-console/](apps/alert-console/) | Next.js live alert UI |
| [infra/grafana/](infra/grafana/) | Grafana dashboards and provisioning |
| [schemas/avro/](schemas/avro/) | Avro event schemas (Schema Registry) |
| [mocks/core-banking/](mocks/core-banking/) | CDC source database migrations and seed data |
| [scripts/](scripts/) | Seed, smoke tests, schema/connector registration, Flink submit |
| [docs/](docs/) | Architecture, domain model, ADRs, phase specs |
| [docker-compose.deploy.yml](docker-compose.deploy.yml) | Staging deploy stack (GHCR images) |
| [.github/workflows/](.github/workflows/) | CI, Docker publish, release, staging deploy |

## Documentation

| Document | Description |
|----------|-------------|
| [docs/phase1-implementation-spec.md](docs/phase1-implementation-spec.md) | Executable Phase 1 specification |
| [docs/phase2-implementation-spec.md](docs/phase2-implementation-spec.md) | Executable Phase 2 specification |
| [docs/review/phase2-exit-checklist.md](docs/review/phase2-exit-checklist.md) | Phase 2 mechanical exit evidence |
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
| **CI** | Full stack + Phase 2 smoke on every push and PR |
| **Container registry** | `state-service`, `alert-service`, `alert-console`, `compliance-cep` on GHCR (`ghcr.io/safetymp/digital-twin-compliance/*`) |
| **Releases** | Tag `v*.*.*` to publish versioned images and a GitHub Release |
| **Staging deploy** | Manual workflow — SSH deploy to a host running `docker-compose.deploy.yml` |
| **Dependabot** | Weekly updates for Go, GitHub Actions, and Docker |
| **CodeQL** | Go security analysis on push, PR, and weekly schedule |
| **Issue templates** | Structured bug reports and feature requests |

Quick deploy on a host with Docker:

```bash
export STATE_SERVICE_IMAGE=ghcr.io/safetymp/digital-twin-compliance/state-service:main
export ALERT_SERVICE_IMAGE=ghcr.io/safetymp/digital-twin-compliance/alert-service:main
export ALERT_CONSOLE_IMAGE=ghcr.io/safetymp/digital-twin-compliance/alert-console:main
export COMPLIANCE_CEP_IMAGE=ghcr.io/safetymp/digital-twin-compliance/compliance-cep:main
./scripts/deploy-stack.sh bootstrap
```

Full guide: [docs/deployment.md](docs/deployment.md).

## CI

GitHub Actions on every push and pull request:

1. Mechanical live evals (`./scripts/run-live-evals.sh`, `./scripts/run-live-evals-phase2.sh`)
2. Docker Compose stack + seed + schema registration + Debezium connector
3. `go vet` and `go test ./...` in `services/state-service` and `services/alert-service`
4. `mvn test` in `jobs/compliance-cep`
5. `./scripts/smoke-test.sh` and `./scripts/smoke-test-phase2.sh`
6. Avro BACKWARD compatibility check (separate workflow)
7. Docker image build and push to GHCR on merge to `main` (separate workflow)

Nightly: extended Phase 2 evals and behavioral eval dry-runs ([eval-nightly.yml](.github/workflows/eval-nightly.yml)).

## Security

Local development stacks (Phase 1 and Phase 2) have **no authentication middleware**. Do not expose default ports to untrusted networks. See [SECURITY.md](SECURITY.md).

## Contributing

Contributions are welcome. Read [CONTRIBUTING.md](CONTRIBUTING.md) before opening a pull request.

This project follows the [Code of Conduct](CODE_OF_CONDUCT.md). GitHub setup (branch protection, environments, first release): [docs/github-setup.md](docs/github-setup.md).
