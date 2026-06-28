# Contributing

Thank you for contributing to the Digital Twin Compliance Platform.

**Copyright holder:** SafetyMP · **License:** [Apache License 2.0](LICENSE)

Please read the [Code of Conduct](CODE_OF_CONDUCT.md) before participating.

## Getting started

1. Fork the repository and clone your fork.
2. Copy `.env.example` to `.env` (never commit `.env`).
3. Follow the [README quick start](README.md#quick-start) to run the stack locally.
4. Create a feature branch from `main` (e.g. `feat/...`, `fix/...`, `docs/...`).

## Development workflow

From the repository root:

```bash
docker compose -f docker-compose.dev.yml up -d --wait
./scripts/seed.sh
./scripts/register-schemas.sh
./scripts/register-debezium-connector.sh
docker compose -f docker-compose.dev.yml restart state-service
docker compose -f docker-compose.dev.yml up -d --wait state-service

cd services/state-service && go test ./...
./scripts/smoke-test.sh

# Phase 2 (monitoring + alerts)
./scripts/submit-flink-job.sh
cd services/alert-service && go test ./...
./scripts/smoke-test-phase2.sh

# Phase 3 (policies + audit ledger)
./scripts/run-policy-ci.sh
cd services/cedar-service && go test ./...
cd services/decision-service && go test ./...
cd services/audit-service && go test ./...
./scripts/smoke-test-phase3.sh
```

Optional mechanical checks:

```bash
./scripts/run-live-evals.sh
./scripts/run-live-evals-phase2.sh
```

Flink CEP unit tests (no local Maven required):

```bash
docker run --rm -v "$PWD/jobs/compliance-cep:/app" -w /app maven:3.9-eclipse-temurin-17 mvn -q test
```

## Pull requests

- Keep PRs focused on a single purpose.
- Ensure CI passes (Compose, unit tests, policy CI, Phase 1–3 smoke, schema compatibility).
- Fill in the PR template checklist and test plan.
- Self-review against:
  - [docs/review/phase1-review-checklist.md](docs/review/phase1-review-checklist.md) for State Service / Phase 1 paths
  - [docs/review/phase2-exit-checklist.md](docs/review/phase2-exit-checklist.md) for Flink, alert-service, alert-console, Grafana, or Phase 2 smoke/CI changes
  - [docs/review/phase3-exit-checklist.md](docs/review/phase3-exit-checklist.md) for Cedar, Zen, audit-service, audit-explorer, or Phase 3 smoke/CI changes

### Commit messages

Use [Conventional Commits](https://www.conventionalcommits.org/) when practical:

- `feat:` new capability
- `fix:` bug fix
- `docs:` documentation only
- `test:` tests only
- `ci:` CI/workflow changes
- `refactor:` behavior-preserving refactor

## Scope by phase

### Phase 1 (ingestion + twin)

For PRs that touch only Phase 1 components, do **not** add Phase 2+ capabilities unless the PR explicitly targets a later phase.

### Phase 2 (monitoring + alerts)

**In scope** for work that extends Phase 2:

- Flink CEP job (`jobs/compliance-cep/`)
- Redis feature store integration
- Alert Service (`services/alert-service/`)
- Alert Console (`apps/alert-console/`)
- Grafana dashboards (`infra/grafana/`)
- Phase 2 smoke test and CI extensions

### Phase 3 (policies + audit ledger)

**In scope** for current work:

- Cedar Policy Service (`services/cedar-service/`)
- Decision Service / GoRules Zen (`services/decision-service/`)
- Audit Service + immudb (`services/audit-service/`)
- Audit Explorer (`apps/audit-explorer/`)
- Policy bundles (`policies/cedar/`, `policies/zen/`)
- `./scripts/run-policy-ci.sh`, `./scripts/smoke-test-phase3.sh`, `./scripts/verify-audit-chain.sh`

### Out of scope (Phase 4+)

Do **not** add unless the PR explicitly targets a later phase:

- Neo4j / Graph Service
- Simulation Service (Python stress/contagion)
- Keycloak / full OIDC auth middleware (mock principal only in Phase 3)
- Regulatory reporting (XBRL)

See [AGENTS.md](AGENTS.md) and [docs/roadmap.md](docs/roadmap.md).

## Coding guidelines

- Match event envelope and idempotency patterns in [docs/data-flow.md](docs/data-flow.md).
- Include `tenant_id` on all entity tables (default single-tenant UUID per [ADR-007](docs/adr/007-phase1-foundation-decisions.md)).
- Publish to Kafka **only** via the transactional outbox in State Service (`services/state-service/internal/outbox/`).
- Use parameterized SQL (no string concatenation for queries).
- Never commit secrets, credentials, or `.env` files.
- Avro schema changes must remain **BACKWARD** compatible; update baseline only with intent.

Service-specific contracts:

- State Service: [services/state-service/AGENTS.md](services/state-service/AGENTS.md)
- Alert Service: [services/alert-service/AGENTS.md](services/alert-service/AGENTS.md)
- Cedar Service: [services/cedar-service/AGENTS.md](services/cedar-service/AGENTS.md)
- Decision Service: [services/decision-service/AGENTS.md](services/decision-service/AGENTS.md)
- Audit Service: [services/audit-service/AGENTS.md](services/audit-service/AGENTS.md)

## Architecture decisions

Significant design changes should be documented as ADRs under [docs/adr/](docs/adr/) before or alongside implementation.

## Deployment

- Images are published to GHCR on merge to `main` ([docker-publish.yml](.github/workflows/docker-publish.yml)): `state-service`, `alert-service`, `alert-console`, `compliance-cep`.
- Staging deploy is manual via the **Deploy Staging** workflow; configure the `staging` environment secrets first.
- See [docs/deployment.md](docs/deployment.md) for host setup, releases, and troubleshooting.
- Release history: [CHANGELOG.md](CHANGELOG.md).

## Questions

Open a [GitHub issue](https://github.com/SafetyMP/Digital-Twin-Compliance/issues) for bugs, questions, or proposed scope changes.
