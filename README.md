# Digital Twin Compliance Platform

Open-source **cutting-edge OSS supervisory** financial-compliance digital-twin reference: event-driven twin state with CDC ingestion, stream processing, policy evaluation, regulatory reporting, hardening overlays, and a tamper-evident audit ledger.

**Maintainers:** [SafetyMP](https://github.com/SafetyMP) · **License:** [Apache License 2.0](LICENSE)

[![CI](https://github.com/SafetyMP/Digital-Twin-Compliance/actions/workflows/ci.yml/badge.svg)](https://github.com/SafetyMP/Digital-Twin-Compliance/actions/workflows/ci.yml)
[![Schema Compatibility](https://github.com/SafetyMP/Digital-Twin-Compliance/actions/workflows/schema-compat.yml/badge.svg)](https://github.com/SafetyMP/Digital-Twin-Compliance/actions/workflows/schema-compat.yml)
[![Docker Publish](https://github.com/SafetyMP/Digital-Twin-Compliance/actions/workflows/docker-publish.yml/badge.svg)](https://github.com/SafetyMP/Digital-Twin-Compliance/actions/workflows/docker-publish.yml)
[![CodeQL](https://github.com/SafetyMP/Digital-Twin-Compliance/actions/workflows/codeql.yml/badge.svg)](https://github.com/SafetyMP/Digital-Twin-Compliance/actions/workflows/codeql.yml)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/SafetyMP/Digital-Twin-Compliance/badge)](https://scorecard.dev/viewer/?uri=github.com/SafetyMP/Digital-Twin-Compliance)
[![Release](https://img.shields.io/github/v/release/SafetyMP/Digital-Twin-Compliance)](https://github.com/SafetyMP/Digital-Twin-Compliance/releases)
[![License](https://img.shields.io/github/license/SafetyMP/Digital-Twin-Compliance)](LICENSE)

<p align="center">
  <img src="docs/assets/demo.gif" alt="Digital Twin demo — Alert, Audit, Graph, Simulation, and Report consoles with shared app switcher" width="900" />
</p>

## Screenshots

| Alert Console | Audit Explorer | Graph Explorer |
|---------------|----------------|----------------|
| ![Alert Console](docs/assets/alert-console.png) | ![Audit Explorer](docs/assets/audit-explorer.png) | ![Graph Explorer](docs/assets/graph-explorer.png) |

| Simulation Console | Report Console |
|--------------------|----------------|
| ![Simulation Console](docs/assets/simulation-console.png) | ![Report Console](docs/assets/report-console.png) |

## TL;DR

**Path A — full platform (cold start, ~10–40 min):**

```bash
git clone https://github.com/SafetyMP/Digital-Twin-Compliance.git && cd Digital-Twin-Compliance
cp .env.example .env
docker compose -f docker-compose.dev.yml up -d --wait && ./scripts/seed.sh
```

Then walk [phase journeys](docs/phase-journeys.md) via [Quick start](#quick-start) (Phases 1–5 smokes; Phase 6–7 overlay + harness when needed). Consoles share an app switcher on ports `3000`–`3005`.

**Path B — policy + audit demo (warm stack, ~5 min):** [docs/demo-phase3.md](docs/demo-phase3.md)

```bash
./scripts/demo-phase3.sh --trigger-alert --restart-policies
SMOKE_PHASE3_SKIP_PREREQS=1 ./scripts/smoke-test-phase3.sh
```

Open [Alert Console](http://localhost:3000) and [Audit Explorer](http://localhost:3002) — use the header switcher for Graph / Simulation / Report.

**How this differs:** streaming CDC → Kafka → twin state with Flink CEP and policy evaluation — not batch regulatory ETL. Later journeys add fixture-validated reporting, OIDC/TLS/OTel overlays, and analytics (effectiveness, contagion→audit, reg→policy proposals). Dev/CI architecture with mock principals; hardening is an explicit Compose overlay, not a multi-tenant production SaaS.

**Quick links:** [Phase journeys](docs/phase-journeys.md) · [Demo Phase 3](docs/demo-phase3.md) · [Architecture](#architecture) · [Quick start](#quick-start) · [REST API](#rest-api) · [Contributing](#contributing)

## Table of contents

- [Who this is for](#who-this-is-for)
- [Phase journeys](#phase-journeys)
- [Features & maturity](#features--maturity)
- [Architecture](#architecture)
- [Prerequisites](#prerequisites)
- [Quick start](#quick-start)
- [REST API](#rest-api)
- [Repository layout](#repository-layout)
- [Documentation](#documentation)
- [DevOps and deployment](#devops-and-deployment)
- [CI](#ci)
- [Security](#security)
- [Community](#community)
- [Contributing](#contributing)

## Who this is for

- **Platform / data engineers** building CDC → Kafka → twin pipelines
- **Compliance / regtech engineers** prototyping CEP rules, policy engines, and audit trails
- **Contributors** who want a runnable, test-gated reference architecture (not slides)

## Phase journeys

Each capability is a **journey name + Phase N** (never Phase alone). Full matrix: [docs/phase-journeys.md](docs/phase-journeys.md).

| Phase | Journey | UI | Smoke |
|-------|---------|----|-------|
| 1 | Ingestion & twin | API `:8080` | `./scripts/smoke-test.sh` |
| 2 | Monitoring & alerts | Alert Console `:3000` | `./scripts/smoke-test-phase2.sh` |
| 3 | Policy & audit | Audit Explorer `:3002` | `./scripts/smoke-test-phase3.sh` |
| 4 | Graph & simulation | Graph `:3003` · Simulation `:3004` | `./scripts/smoke-test-phase4.sh` |
| 5 | Regulatory reporting | Report Console `:3005` | `./scripts/smoke-test-phase5.sh` |
| 6 | Hardening | OIDC/TLS edge `:8180`/`:8443` | `./scripts/smoke-test-phase6-*.sh` |
| 7 | Cutting-edge analytics | APIs (no dedicated UI) | `./scripts/smoke-test-phase7-*.sh` |

Phases 1–4 smoke in CI. Phases 5–7 + `./scripts/harness/verify.sh` are local/harness gates. **CECT** is the corporate program that delivered Phases 5–7 — not a product tier.

**Roadmap & gaps:** [ROADMAP.md](ROADMAP.md) · **Support:** [SUPPORT.md](SUPPORT.md) · **Explainability:** [docs/explainability.md](docs/explainability.md)

## Features & maturity

| Journey | Status | Notes |
|---------|--------|-------|
| Ingestion & twin | Stable on `main` | CI + `./scripts/smoke-test.sh` |
| Monitoring & alerts | Stable on `main` | INT-M001, INT-M002, BASEL-M001 |
| Policy & audit | Stable on `main` | CI + [demo-phase3.md](docs/demo-phase3.md) |
| Graph & simulation | Stable on `main` (CI) | [demo-phase4.md](docs/demo-phase4.md) |
| Regulatory reporting | Smoke-gated | [demo-phase5.md](docs/demo-phase5.md) · [ADR-011](docs/adr/011-phase5-reporting-foundation.md) — not commercial XBRL parity |
| Hardening | Smoke-gated overlay | [ADR-012](docs/adr/012-phase6-hardening-foundation.md) |
| Cutting-edge analytics | Smoke-gated | [ADR-013](docs/adr/013-phase7-cutting-edge-foundation.md); contagion is on-demand, not a scheduled stream |
| GHCR deploy (12 images) | Stable | Phase 1–4 via `docker-compose.deploy.yml` |

Release history: [CHANGELOG.md](CHANGELOG.md) · [GitHub Releases](https://github.com/SafetyMP/Digital-Twin-Compliance/releases)

## Architecture

```text
mock core-banking (PostgreSQL)
        │  CDC (Debezium) — personas + payments
        ▼
   Kafka (domain.events.*, twin.state.updated, compliance.audit.*)
        │
        ├────────────────────────┬─────────────────────────┐
        ▼                        ▼                         ▼
  State Service (Go)     Flink CEP (Java)          Audit Service (Go)
   ├── twin store               ├── Redis features        ├── immudb (hash chain)
   ├── outbox                   └── compliance.alerts     └── /api/v1/audit/*
   └── /api/v1/personas                 │                         ▲
                                        ▼                         │
                                Alert Service (Go)                │
                                 ├── PostgreSQL                   │
                                 └── /api/v1/alerts ── evidenceRef ┘
                                        │
                    ┌───────────────────┼───────────────────┐
                    ▼                   ▼                   ▼
            Alert Console      Cedar Service        Decision Service
            (Next.js)          POST /evaluate       POST /evaluate (Zen)
                    │
                    ▼
            Audit Explorer (Next.js) — chain verify + search

Neo4j + Graph Service (:8093) + Simulation Service (:8094)
Graph Explorer (:3003) + Simulation Console (:3004) — Phase 4 (dev compose)

Grafana ← Flink / Kafka metrics (Compose)
```

Details: [docs/architecture.md](docs/architecture.md) · [docs/data-flow.md](docs/data-flow.md) · [docs/adr/](docs/adr/)

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and Docker Compose v2
- [Go 1.25+](https://go.dev/dl/) (`go.mod`; CI/Docker images use 1.25–1.26 per service)
- [Node 20](https://nodejs.org/) — Next.js app builds (matches CI)
- Python 3.11+ — `simulation-service` tests (`pytest`; CI uses system `python3`)
- Java 17 — Flink CEP job (`jobs/compliance-cep`; matches CI)
- CLI tools: `curl`, `jq`, `psql` (PostgreSQL client)
- **Maven not required** — run Flink job unit tests via Docker (see [CONTRIBUTING.md](CONTRIBUTING.md))

## Quick start

Run the full platform locally (~10–40 minutes cold start depending on image pulls):

### 1. Ingestion & twin

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

### 2. Monitoring & alerts

After ingestion smoke passes, ensure the Flink CEP job is **RUNNING** (Compose `flink-job-submitter` usually handles this; re-submit if needed):

```bash
./scripts/submit-flink-job.sh
./scripts/smoke-test-phase2.sh
```

**Gotchas**

- Redis host port is **6380** (Compose maps `6380:6379`); see `REDIS_URL` in `.env.example`.
- Kafka topics for Phase 2 are created by `./scripts/seed.sh` (via `create-kafka-topics.sh`).
- Monitoring smoke injects payment bursts and checks INT-M001, INT-M002, and BASEL-M001 alerts end-to-end.

### 3. Policy & audit ledger

After monitoring smoke passes:

```bash
./scripts/run-policy-ci.sh
./scripts/smoke-test-phase3.sh
```

Walkthrough: [docs/demo-phase3.md](docs/demo-phase3.md).

**Gotchas**

- UIs proxy Go APIs via Next.js `/api/*` routes — do not `fetch` Cedar/Audit/Alert backends from the browser (no CORS).
- Cedar bind mounts can appear empty until `docker compose restart cedar-service decision-service` (common when the repo path contains spaces).

### 4. Graph & simulation

```bash
./scripts/wait-graph-seeded.sh
./scripts/smoke-test-phase4.sh
```

Open [Graph Explorer](http://localhost:3003) and [Simulation Console](http://localhost:3004). Demo: [docs/demo-phase4.md](docs/demo-phase4.md).

### 5. Regulatory reporting

```bash
./scripts/smoke-test-phase5.sh
```

Open [Report Console](http://localhost:3005). Demo: [docs/demo-phase5.md](docs/demo-phase5.md).

### 6–7. Hardening & cutting-edge analytics

```bash
# Phase 6 overlay (OIDC / TLS / OTel)
docker compose -f docker-compose.dev.yml -f docker-compose.hardening.yml up -d --wait
./scripts/smoke-test-phase6-oidc.sh
./scripts/smoke-test-phase6-harden.sh

# Phase 7 API journeys + harness
./scripts/smoke-test-phase7-effectiveness.sh
./scripts/smoke-test-phase7-contagion.sh
./scripts/smoke-test-phase7-reg2policy.sh
./scripts/smoke-test-phase7-graph.sh
./scripts/harness/verify.sh
```

### Verify

```bash
# Ingestion
cd services/state-service && go test ./...
curl -s http://localhost:8080/api/v1/health
curl -s "http://localhost:8080/api/v1/personas?personaType=Institution&limit=5" | jq

# Monitoring
cd services/alert-service && go test ./...
curl -s http://localhost:8085/api/v1/health
curl -s "http://localhost:8085/api/v1/alerts?limit=5" | jq
open http://localhost:3000   # Alert Console
open http://localhost:8082   # Flink UI
open http://localhost:3030   # Grafana (Compose maps 3030:3000)

# Policy & audit
./scripts/run-policy-ci.sh
cd services/cedar-service && go test ./...
curl -s http://localhost:8091/api/v1/health | jq
curl -s http://localhost:8090/api/v1/audit/verify | jq
open http://localhost:3000   # Alert Console (switcher → other consoles)
open http://localhost:3002   # Audit Explorer
open http://localhost:3003   # Graph Explorer
open http://localhost:3004   # Simulation Console
open http://localhost:3005   # Report Console
```

### Local service URLs

| Service | URL |
|---------|-----|
| State Service | `http://localhost:8080` |
| Alert Service | `http://localhost:8085` |
| Alert Console | `http://localhost:3000` |
| Audit Explorer | `http://localhost:3002` |
| Graph Explorer | `http://localhost:3003` |
| Simulation Console | `http://localhost:3004` |
| Report Console | `http://localhost:3005` |
| Graph Service | `http://localhost:8093` |
| Simulation Service | `http://localhost:8094` |
| Reporting Service | `http://localhost:8095` |
| Cedar Service | `http://localhost:8091` |
| Decision Service | `http://localhost:8092` |
| Audit Service | `http://localhost:8090` |
| Flink JobManager UI | `http://localhost:8082` |
| Grafana | `http://localhost:3030` |
| Redis | `localhost:6380` |
| Debezium Connect | `http://localhost:8083` |

### Tear down

```bash
docker compose -f docker-compose.dev.yml down -v
```

## REST API

### State Service

Base URL: `http://localhost:8080/api/v1`

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Liveness probe |
| `GET` | `/personas` | List twin personas (`personaType`, `limit`, `offset`) |
| `GET` | `/personas/{personaId}` | Single persona by ID |

Persona types: `Institution`, `Account`, `Instrument`.

### Alert Service

Base URL: `http://localhost:8085/api/v1`

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Liveness probe |
| `GET` | `/alerts` | List alerts (`limit`, `offset`, filters) |
| `GET` | `/alerts/{alertId}` | Single alert by ID |
| `POST` | `/alerts/{alertId}/acknowledge` | Acknowledge an alert |
| `GET` | `/ws/alerts` | WebSocket stream (service-to-service / smoke scripts; alert-console uses `/api/alerts` polling) |

Full contract: [docs/phase2-implementation-spec.md](docs/phase2-implementation-spec.md) · [services/alert-service/AGENTS.md](services/alert-service/AGENTS.md)

### Cedar Service

Base URL: `http://localhost:8091/api/v1`

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Policy load status (`policiesLoaded`, `ruleCodes`) |
| `POST` | `/evaluate` | Cedar access/obligation evaluation → `RuleDecision` |

### Decision Service

Base URL: `http://localhost:8092/api/v1`

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Zen model load status |
| `GET` | `/rules` | Loaded regulatory models |
| `POST` | `/evaluate` | Zen decision evaluation → `RuleDecision` |

### Audit Service

Base URL: `http://localhost:8090/api/v1`

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | immudb connectivity |
| `GET` | `/audit/entries` | Search ledger (`ruleCode`, `from`, `to`, `subjectId`) |
| `GET` | `/audit/entries/{entryId}` | Single tamper-evident entry |
| `GET` | `/audit/verify` | Hash-chain integrity check |

Full contract: [docs/phase3-implementation-spec.md](docs/phase3-implementation-spec.md) · [docs/demo-phase3.md](docs/demo-phase3.md)

## Repository layout

| Path | Purpose |
|------|---------|
| [services/state-service/](services/state-service/) | Go REST API, Kafka consumer, transactional outbox |
| [services/alert-service/](services/alert-service/) | Alerts REST + WebSocket + `compliance.alerts` consumer |
| [services/cedar-service/](services/cedar-service/) | Cedar policy evaluation |
| [services/decision-service/](services/decision-service/) | GoRules Zen decision engine |
| [services/audit-service/](services/audit-service/) | immudb audit ledger + verify API |
| [jobs/compliance-cep/](jobs/compliance-cep/) | Flink CEP job (Java) — INT-M001, INT-M002, BASEL-M001 |
| [apps/alert-console/](apps/alert-console/) | Next.js alert UI (Monitoring · Phase 2) |
| [apps/audit-explorer/](apps/audit-explorer/) | Next.js audit explorer (Policy & audit · Phase 3) |
| [apps/graph-explorer/](apps/graph-explorer/) | Next.js exposure graph (Graph · Phase 4) |
| [apps/simulation-console/](apps/simulation-console/) | Next.js stress sim (Simulation · Phase 4) |
| [apps/report-console/](apps/report-console/) | Next.js report lifecycle (Reporting · Phase 5) |
| [packages/console-shell/](packages/console-shell/) | Shared console chrome + app switcher |
| [policies/](policies/) | Cedar (`.cedar`) and Zen (`.zen`) policy bundles |
| [infra/grafana/](infra/grafana/) | Grafana dashboards and provisioning |
| [schemas/avro/](schemas/avro/) | Avro event schemas (Schema Registry) |
| [mocks/core-banking/](mocks/core-banking/) | CDC source database migrations and seed data |
| [scripts/](scripts/) | Seed, smoke tests, schema/connector registration, Flink submit |
| [docs/](docs/) | Architecture, domain model, ADRs, phase specs |
| [docker-compose.deploy.yml](docker-compose.deploy.yml) | Staging deploy stack (GHCR images) |
| [.github/workflows/](.github/workflows/) | CI, Docker publish, release, staging deploy |

## Documentation

### Using & extending the platform

| Document | Description |
|----------|-------------|
| [ROADMAP.md](ROADMAP.md) | Public roadmap, stability, planned work |
| [CONTRIBUTING.md](CONTRIBUTING.md) | How to contribute |
| [SUPPORT.md](SUPPORT.md) | Support channels and expectations |
| [CHANGELOG.md](CHANGELOG.md) | Release history |
| [docs/architecture.md](docs/architecture.md) | C4 diagrams and component map |
| [docs/domain-model.md](docs/domain-model.md) | Entities, personas, glossary |
| [docs/data-flow.md](docs/data-flow.md) | Event envelopes, idempotency, topics |
| [docs/deployment.md](docs/deployment.md) | GHCR, releases, staging deploy |
| [docs/phase-journeys.md](docs/phase-journeys.md) | Canonical Phase 1–7 journey map |
| [docs/demo-phase3.md](docs/demo-phase3.md) | Policy + audit demo runbook |
| [docs/demo-phase4.md](docs/demo-phase4.md) | Graph + simulation demo runbook |
| [docs/demo-phase5.md](docs/demo-phase5.md) | Regulatory reporting demo runbook |
| [docs/adr/](docs/adr/) | Architecture decision records |

### Maintainer & implementation references

| Document | Description |
|----------|-------------|
| [docs/roadmap.md](docs/roadmap.md) | Detailed phased plan (internal engineering) |
| [docs/phase1-implementation-spec.md](docs/phase1-implementation-spec.md) | Ingestion specification |
| [docs/phase2-implementation-spec.md](docs/phase2-implementation-spec.md) | Monitoring specification |
| [docs/phase3-implementation-spec.md](docs/phase3-implementation-spec.md) | Policy & audit specification |
| [docs/phase4-implementation-spec.md](docs/phase4-implementation-spec.md) | Graph & simulation specification |
| [docs/phase5-implementation-spec.md](docs/phase5-implementation-spec.md) | Reporting specification |
| [docs/phase6-implementation-spec.md](docs/phase6-implementation-spec.md) | Hardening specification |
| [docs/phase7-implementation-spec.md](docs/phase7-implementation-spec.md) | Cutting-edge analytics specification |
| [AGENTS.md](AGENTS.md) | Coding-agent contract (CI scope, smoke order) |
| [docs/github-setup.md](docs/github-setup.md) | Branch protection, releases, community settings |

## DevOps and deployment

| Capability | Details |
|------------|---------|
| **CI** | Full stack + Phase 1–3 smoke, policy CI, and eval fixtures on every push and PR |
| **Container registry** | Eight application images on GHCR (`ghcr.io/safetymp/digital-twin-compliance/*`) |
| **Releases** | Tag `v*.*.*` to publish versioned images and a GitHub Release |
| **Staging deploy** | Manual workflow — SSH deploy to a host running `docker-compose.deploy.yml` |
| **Dependabot** | Weekly updates for Go (all services), npm (UIs), Maven (CEP), GitHub Actions, and Docker |
| **CodeQL** | Go security analysis on push, PR, and weekly schedule |
| **OpenSSF Scorecard** | Supply-chain security posture on push and weekly schedule ([`.github/workflows/scorecard.yml`](.github/workflows/scorecard.yml)) |
| **Issue templates** | Structured bug reports and feature requests |

Quick deploy on a host with Docker (requires repo clone for policy bind mounts):

```bash
PREFIX=ghcr.io/safetymp/digital-twin-compliance
TAG=main
export STATE_SERVICE_IMAGE=${PREFIX}/state-service:${TAG}
export ALERT_SERVICE_IMAGE=${PREFIX}/alert-service:${TAG}
export ALERT_CONSOLE_IMAGE=${PREFIX}/alert-console:${TAG}
export COMPLIANCE_CEP_IMAGE=${PREFIX}/compliance-cep:${TAG}
export AUDIT_SERVICE_IMAGE=${PREFIX}/audit-service:${TAG}
export CEDAR_SERVICE_IMAGE=${PREFIX}/cedar-service:${TAG}
export DECISION_SERVICE_IMAGE=${PREFIX}/decision-service:${TAG}
export AUDIT_EXPLORER_IMAGE=${PREFIX}/audit-explorer:${TAG}
./scripts/deploy-stack.sh bootstrap
```

Full guide: [docs/deployment.md](docs/deployment.md).

## CI

GitHub Actions on every push and pull request ([ci.yml](.github/workflows/ci.yml), job `ci`):

1. `go vet` / `go test ./...` — six Go services including `graph-service`
2. Python `pytest` — `simulation-service`; Cedar CLI + `./scripts/run-policy-ci.sh`
3. `mvn test` in `jobs/compliance-cep`; `./scripts/check-kafka-contracts.sh`
4. Agent worktree / dependency-wave script hygiene (`./scripts/check-agent-worktrees.sh`)
5. Mechanical live evals (`./scripts/run-live-evals.sh`, `./scripts/run-live-evals-phase2.sh`) and `./scripts/run-eval-fixtures.sh`
6. Docker Compose stack → seed → Phase 3/4 service bring-up → schema registration → Debezium → outbox drain → twin pipeline verify
7. `mvn package`, Next.js builds (`alert-console`, `audit-explorer`, `graph-explorer`, `simulation-console`), Flink job submit
8. `./scripts/smoke-test.sh`, `./scripts/smoke-test-phase2.sh`, `./scripts/smoke-test-phase3.sh`, `./scripts/smoke-test-phase4.sh`
9. `./scripts/check-coverage-gates.sh`

Separate workflows:

| Workflow | Trigger | Purpose |
|----------|---------|---------|
| [schema-compat.yml](.github/workflows/schema-compat.yml) | Push, PR | Avro BACKWARD compatibility |
| [policy-gates.yml](.github/workflows/policy-gates.yml) | PR (path-filtered) | Policy CI when `policies/**` or policy services change |
| [codeql.yml](.github/workflows/codeql.yml) | Push, PR, weekly | Go security analysis (`analyze` check) |
| [docker-publish.yml](.github/workflows/docker-publish.yml) | Push to `main`, version tags | GHCR image publish |
| [eval-nightly.yml](.github/workflows/eval-nightly.yml) | Daily schedule, manual | Eval fixtures, harness calibration, extended smoke |

Branch protection and environment setup: [docs/github-setup.md](docs/github-setup.md).

## Security

Local stacks use **mock principals only** — no production auth middleware. Default Compose credentials are not production-safe. Do not expose service ports to untrusted networks. See [SECURITY.md](SECURITY.md).

## Community

- **Bugs & features:** [GitHub Issues](https://github.com/SafetyMP/Digital-Twin-Compliance/issues/new/choose)
- **Security:** [Private advisories](https://github.com/SafetyMP/Digital-Twin-Compliance/security/advisories/new) — see [SECURITY.md](SECURITY.md)
- **Conduct:** [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)
- **Governance:** Maintained by SafetyMP; contributions via PR welcome on `main`

## Contributing

Contributions are welcome. Read [CONTRIBUTING.md](CONTRIBUTING.md) and [ROADMAP.md](ROADMAP.md) before opening a pull request.

GitHub setup for maintainers: [docs/github-setup.md](docs/github-setup.md).
