# Changelog

All notable changes to this project are documented here. Release tags (`v*.*.*`) publish Phase 1–2 runtime images to GHCR; see [docs/deployment.md](docs/deployment.md).

Format loosely follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Added

- Phase 3 stack: Cedar Service, Decision Service (Zen), Audit Service (immudb), Audit Explorer UI
- `./scripts/smoke-test-phase3.sh`, `./scripts/run-policy-ci.sh`, `./scripts/verify-audit-chain.sh`
- CI job `ci` — full Phase 1–3 smoke, policy gates, eval fixtures, coverage gates
- GitHub issue/PR templates updated for Phase 3 scope and smoke triage

### Changed

- README architecture diagram and quick start cover Phase 3
- Dependabot watches all Go services, both Next.js apps, Flink CEP (Maven), Actions, and Docker bases

### Notes

- Phase 3 services are exercised in CI and `docker-compose.dev.yml` but are **not** yet published to GHCR or `docker-compose.deploy.yml`

## Phase milestones on `main`

| Phase | Highlights | Evidence |
|-------|------------|----------|
| **Phase 1** | Debezium CDC, State Service, outbox, persona API | [phase1-exit checklist](docs/phase1-implementation-spec.md#10-phase-1-exit-criteria-checklist) |
| **Phase 2** | Flink CEP, Alert Service, alert console, Grafana | [phase2-exit-checklist.md](docs/review/phase2-exit-checklist.md) |
| **Phase 3** | Cedar + Zen, immudb audit ledger, Audit Explorer | [phase3-exit-checklist.md](docs/review/phase3-exit-checklist.md) |

## Tagged releases

When cutting a release:

1. Update this file under a new `## [x.y.z] — YYYY-MM-DD` heading
2. Tag `vX.Y.Z` and push — triggers [release.yml](.github/workflows/release.yml) and [docker-publish.yml](.github/workflows/docker-publish.yml)

Previous tags (if any) are listed on the [GitHub Releases](https://github.com/SafetyMP/Digital-Twin-Compliance/releases) page.
