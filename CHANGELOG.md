# Changelog

All notable changes to this open-source project are documented here.

Release tags (`v*.*.*`) publish all application images to GHCR and create a GitHub Release. See [docs/deployment.md](docs/deployment.md) and [ROADMAP.md](ROADMAP.md) for deploy scope.

Format loosely follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Added

- GHCR publish and deploy stack for Phase 3 (`audit-service`, `cedar-service`, `decision-service`, `audit-explorer`)
- Public [ROADMAP.md](ROADMAP.md) and [SUPPORT.md](SUPPORT.md) for evergreen OSS presentation
- Policy & audit stack: Cedar Service, Decision Service (Zen), Audit Service (immudb), Audit Explorer UI
- `./scripts/smoke-test-phase3.sh`, `./scripts/run-policy-ci.sh`, `./scripts/verify-audit-chain.sh`

### Changed

- README and CONTRIBUTING lead with product capabilities instead of internal phase delivery language
- Issue/PR templates use component-based triage
- CI job `ci` runs full ingestion → monitoring → policy/audit smoke on every PR
- Restored [release.yml](.github/workflows/release.yml) workflow (accidentally truncated in docs refresh)

### Removed

- CHANGELOG note that policy/audit services were dev/CI-only for deploy

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

Previous tags: [GitHub Releases](https://github.com/SafetyMP/Digital-Twin-Compliance/releases).
