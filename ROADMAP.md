# Roadmap

Public roadmap for the **Digital Twin Compliance Platform** — an open-source reference stack for event-driven financial twins with embedded compliance monitoring.

Detailed planning, risks, and phase specs: [docs/roadmap.md](docs/roadmap.md).

## On `main` today

| Capability | What you get | Try it |
|------------|--------------|--------|
| **Ingestion & twin** | Debezium CDC → Kafka → State Service, transactional outbox, persona REST API | `./scripts/smoke-test.sh` |
| **Real-time monitoring** | Flink CEP, Redis features, Alert Service, alert console, Grafana | `./scripts/smoke-test-phase2.sh` |
| **Policy & audit** | Cedar + GoRules Zen, immudb ledger, Audit Explorer, `evidenceRef` on alerts | `./scripts/smoke-test-phase3.sh` · [demo runbook](docs/demo-phase3.md) |
| **Graph + simulation** | Neo4j graph, simulation service, graph/sim UIs (dev compose) | `./scripts/smoke-test-phase4.sh` |

CI on every PR runs all four smoke suites plus policy CI ([README § CI](README.md#ci)).

**Release [v0.1.0](https://github.com/SafetyMP/Digital-Twin-Compliance/releases/tag/v0.1.0)** — Phase 1–3 smoke-stable; Flink 1.20 runtime aligned.

## Phase 4 (beta on `main`)

Graph + simulation services ship in dev compose and CI. GHCR/deploy packaging for Phase 4 is still planned. Handoff: [docs/handoff-phase4-agent.md](docs/handoff-phase4-agent.md) · [docs/phase4-implementation-spec.md](docs/phase4-implementation-spec.md).

## Stability

| Area | Status |
|------|--------|
| Local dev stack (`docker-compose.dev.yml`) | **Active** — primary development path |
| GHCR images (8 services) | **Published** on merge to `main` and semver tags |
| GHCR deploy (`docker-compose.deploy.yml`) | **Phase 1–3 runtime** — policy bundles bind-mounted from repo clone |
| Production hardening (TLS, OIDC, secrets) | **Not yet** — mock principals and dev credentials only |

See [docs/deployment.md](docs/deployment.md#production-readiness) for the honest production gap list.

## Planned (not built yet)

| Theme | Examples | Tracking |
|-------|----------|----------|
| **Phase 4 deploy** | GHCR images + deploy compose for graph/simulation | [docs/deployment.md](docs/deployment.md) |
| **Regulatory reporting** | XBRL / SDMX outputs, report UI | [docs/roadmap.md § Phase 5](docs/roadmap.md#phase-5-regulatory-reporting) |
| **Hardening** | Keycloak/OIDC, production Compose/K8s paths | [docs/roadmap.md](docs/roadmap.md) |

Features outside this roadmap: open a [feature request](https://github.com/SafetyMP/Digital-Twin-Compliance/issues/new/choose) for discussion before large PRs.

## Releases

- **Continuous integration** on `main`
- **Semver tags** (`v*.*.*`) publish GHCR images and a GitHub Release — see [CHANGELOG.md](CHANGELOG.md)
- Maintainers aim to tag releases when a capability milestone is smoke-stable (no fixed calendar yet)

## How to influence priority

1. **+1** or comment on an existing issue
2. Open a **feature request** with component + use case
3. Submit a **PR** with tests/smoke updates ([CONTRIBUTING.md](CONTRIBUTING.md))
