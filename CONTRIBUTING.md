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
```

Optional mechanical checks:

```bash
./scripts/run-live-evals.sh
```

## Pull requests

- Keep PRs focused on a single purpose.
- Ensure CI passes (Compose, tests, smoke test, schema compatibility).
- Fill in the PR template checklist and test plan.
- Self-review against [docs/review/phase1-review-checklist.md](docs/review/phase1-review-checklist.md) for Phase 1 changes.

### Commit messages

Use [Conventional Commits](https://www.conventionalcommits.org/) when practical:

- `feat:` new capability
- `fix:` bug fix
- `docs:` documentation only
- `test:` tests only
- `ci:` CI/workflow changes
- `refactor:` behavior-preserving refactor

## Scope (Phase 1)

Unless a PR explicitly targets a later phase, stay within Phase 1 scope. Do **not** add:

- Apache Flink / CEP jobs
- Cedar Policy Service / GoRules Zen
- immudb audit ledger
- Neo4j / Graph Service
- Next.js UI / WebSocket alert console
- Keycloak / auth middleware
- Regulatory reporting (XBRL)

See [AGENTS.md § Out of scope](AGENTS.md#out-of-scope-phase-1) and [docs/roadmap.md](docs/roadmap.md).

## Coding guidelines

- Match event envelope and idempotency patterns in [docs/data-flow.md](docs/data-flow.md).
- Include `tenant_id` on all entity tables (default single-tenant UUID per [ADR-007](docs/adr/007-phase1-foundation-decisions.md)).
- Publish to Kafka **only** via the transactional outbox (`internal/outbox/`).
- Use parameterized SQL (no string concatenation for queries).
- Never commit secrets, credentials, or `.env` files.
- Avro schema changes must remain **BACKWARD** compatible; update baseline only with intent.

For Go work in the State Service, see [services/state-service/AGENTS.md](services/state-service/AGENTS.md).

## Architecture decisions

Significant design changes should be documented as ADRs under [docs/adr/](docs/adr/) before or alongside implementation.

## Deployment

- Images are published to GHCR on merge to `main` ([docker-publish.yml](.github/workflows/docker-publish.yml)).
- Staging deploy is manual via the **Deploy Staging** workflow; configure the `staging` environment secrets first.
- See [docs/deployment.md](docs/deployment.md) for host setup, releases, and troubleshooting.

## Questions

Open a [GitHub issue](https://github.com/SafetyMP/Digital-Twin-Compliance/issues) for bugs, questions, or proposed scope changes.
