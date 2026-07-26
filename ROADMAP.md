# Roadmap

Public roadmap for the **Digital Twin Compliance Platform** — an open-source supervisory financial-compliance digital-twin reference.

**Phase journeys (canonical map):** [docs/phase-journeys.md](docs/phase-journeys.md). Detailed engineering plan: [docs/roadmap.md](docs/roadmap.md).

**CECT** is the corporate delivery program that landed Phases 5–7 on this site; it is not a product tier.

## On `main` today

| Journey | Phase | What you get | Try it |
|---------|-------|--------------|--------|
| **Ingestion & twin** | 1 | Debezium CDC → Kafka → State Service, transactional outbox, persona REST API | `./scripts/smoke-test.sh` |
| **Monitoring & alerts** | 2 | Flink CEP, Redis features, Alert Service, Alert Console, Grafana | `./scripts/smoke-test-phase2.sh` |
| **Policy & audit** | 3 | Cedar + GoRules Zen, immudb ledger, Audit Explorer, `evidenceRef` on alerts | `./scripts/smoke-test-phase3.sh` · [demo](docs/demo-phase3.md) |
| **Graph & simulation** | 4 | Neo4j exposure graph, deterministic stress simulation, Graph Explorer + Simulation Console | `./scripts/smoke-test-phase4.sh` · [demo](docs/demo-phase4.md) |
| **Regulatory reporting** | 5 | FINREP F01 / AnaCredit / DORA; MinIO Object Lock; Report Console lifecycle | `./scripts/smoke-test-phase5.sh` · [demo](docs/demo-phase5.md) · [ADR-011](docs/adr/011-phase5-reporting-foundation.md) |
| **Hardening** | 6 | Keycloak/OIDC edge, TLS nginx, OpenTelemetry, DR runbooks, explainability pack | `docker-compose.hardening.yml` · `./scripts/smoke-test-phase6-*.sh` · [ADR-012](docs/adr/012-phase6-hardening-foundation.md) |
| **Cutting-edge analytics** | 7 | Control-effectiveness twin, contagion→audit, reg→policy proposals (no auto-deploy), graph path/centrality | `./scripts/smoke-test-phase7-*.sh` · [ADR-013](docs/adr/013-phase7-cutting-edge-foundation.md) |

CI on every PR runs Phase 1–4 smoke suites plus policy CI ([README § CI](README.md#ci)). Phase 5–7 smokes and `./scripts/harness/verify.sh` are local/harness gates (CI wiring is a follow-up).

**Release [v0.1.0](https://github.com/SafetyMP/Digital-Twin-Compliance/releases/tag/v0.1.0)** — Phase 1–3 smoke-stable; Flink 1.20 runtime aligned. Phases 5–7 (CECT delivery) land under `[Unreleased]` until the next semver tag.

## Stability

| Area | Status |
|------|--------|
| Local dev stack (`docker-compose.dev.yml`) | **Active** — primary development path |
| Hardening overlay (`docker-compose.hardening.yml`) | **Active** — OIDC/TLS/OTel proofs; direct ports remain for Phase 1–4 smokes |
| GHCR images (12 Phase 1–4 services) | **Published** on merge to `main` and semver tags |
| GHCR deploy (`docker-compose.deploy.yml`) | **Phase 1–4 runtime** — Phase 5–7 images not yet in docker-publish |
| Production multi-tenant SaaS / Vault injection | **Not yet** — see [docs/deployment.md](docs/deployment.md#residuals-honest) |

## Honest residuals (shipped with Phases 5–7)

| Residual | Notes |
|----------|-------|
| Direct service ports without OIDC | Intentional so Phase 1–4 smokes stay token-free; OIDC is via the edge overlay |
| Fixture-equivalent FINREP taxonomy | Allowed for reference stack; not commercial EBA taxonomy parity |
| CI-sized load profile | Full 10K evt/s not run on default CI hosts |
| Contagion | On-demand API BFS → audit, not a scheduled continuous stream |
| Reg→policy Zen stub | Cedar stub reject/accept demonstrated; Zen path deferred |

## Planned (not built yet)

| Theme | Examples | Tracking |
|-------|----------|----------|
| **Phase 5–7 CI + GHCR** | Wire phase5–7 smokes into `ci.yml`; publish reporting/oidc/report-console images | Follow-up after Phases 5–7 merge |
| **Staging secrets** | Vault (or equivalent) injection; drop documented Keycloak demo secrets | [docs/deployment.md](docs/deployment.md) |
| **HA / multi-AZ** | HA immudb, multi-AZ Kafka, full K8s/Terraform reference | Out of current delivery scope |

Features outside this roadmap: open a [feature request](https://github.com/SafetyMP/Digital-Twin-Compliance/issues/new/choose) for discussion before large PRs.

## Releases

- **Continuous integration** on `main`
- **Semver tags** (`v*.*.*`) publish GHCR images and a GitHub Release — see [CHANGELOG.md](CHANGELOG.md)
- Maintainers aim to tag releases when a capability milestone is smoke-stable (no fixed calendar yet)

## How to influence priority

1. **+1** or comment on an existing issue
2. Open a **feature request** with component + use case
3. Submit a **PR** with tests/smoke updates ([CONTRIBUTING.md](CONTRIBUTING.md))
