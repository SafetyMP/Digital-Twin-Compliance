# Changelog

All notable changes to this open-source project are documented here.

Release tags (`v*.*.*`) publish all application images to GHCR and create a GitHub Release. See [docs/deployment.md](docs/deployment.md) and [ROADMAP.md](ROADMAP.md) for deploy scope.

Format loosely follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

## [0.1.0] — 2026-06-29

First semver release: Phase 1–3 local stack, GHCR deploy for eight application images, and Phase 3b Decision Service hot path for Flink CEP.

### Added

- GHCR publish and deploy stack for Phase 3 (`audit-service`, `cedar-service`, `decision-service`, `audit-explorer`)
- Public [ROADMAP.md](ROADMAP.md) and [SUPPORT.md](SUPPORT.md)
- Policy & audit stack: Cedar Service, Decision Service (Zen), Audit Service (immudb), Audit Explorer UI
- Phase 3b: Flink CEP calls Decision Service for INT-M001, INT-M002, and BASEL-M001 when `CEP_DECISION_SERVICE_URL` is set
- Agent git worktrees, dependency waves, and `./scripts/demo-agent-workflows.sh`
- `./scripts/smoke-test-phase3.sh`, `./scripts/run-policy-ci.sh`, `./scripts/verify-audit-chain.sh`
- Repo-local `scripts/agent-worktree/config.py` for CI (dependency-wave validation)

### Changed

- Flink CEP job and Compose runtime aligned to **Apache Flink 1.20.5** (`flink:1.20-java17`, kafka connector `3.4.0-1.20`)
- Next.js **14.2.35** and TypeScript **5.9.3** on alert-console and audit-explorer
- README and CONTRIBUTING lead with product capabilities
- CI runs full Phase 1–3 smoke on every PR
- Dependabot ignores semver-major npm and Maven bumps (coordinate upgrades manually)
- Restored [release.yml](.github/workflows/release.yml) workflow

## History on `main`

Capability milestones (internal phase specs remain under `docs/phase*-implementation-spec.md`):

| Milestone | Highlights |
|-----------|------------|
| Ingestion & twin | Debezium CDC, State Service, outbox, persona API |
| Monitoring | Flink CEP, Alert Service, alert console, Grafana |
| Policy & audit | Cedar + Zen, immudb ledger, Audit Explorer, `evidenceRef` |

## Cutting a release

1. Move `[Unreleased]` items to `## [x.y.z] — YYYY-MM-DD`
2. Tag `vX.Y.Z` and push — triggers [release.yml](.github/workflows/release.yml) and [docker-publish.yml](.github/workflows/docker-publish.yml)
3. Validate deploy with tagged images (see [docs/deployment.md](docs/deployment.md#release-validation))

Previous tags: [GitHub Releases](https://github.com/SafetyMP/Digital-Twin-Compliance/releases).
