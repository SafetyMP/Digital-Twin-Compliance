# Security Policy

## Supported versions

| Version / phase | Supported | Notes |
|-----------------|-----------|-------|
| `main` (Phase 1–4 dev stack) | Yes | Local development and CI only — not production-hardened |
| Tagged releases (`v*.*.*`) | Yes | GHCR images for all application services; see [CHANGELOG.md](CHANGELOG.md) |
| Production hardening (TLS, OIDC) | Not yet released | Mock principals and dev credentials only |

Report security issues against the current `main` branch unless you are running a specific release tag.

## Reporting a vulnerability

If you discover a security issue, **do not** open a public GitHub issue.

Contact the maintainers privately through [GitHub Security Advisories](https://github.com/SafetyMP/Digital-Twin-Compliance/security/advisories/new) on this repository, or reach out to [SafetyMP](https://github.com/SafetyMP) directly.

Please include:

- Description of the vulnerability
- Steps to reproduce
- Impact assessment
- Suggested fix (if any)

We aim to acknowledge reports within a reasonable timeframe and will coordinate disclosure after a fix is available.

Support expectations for non-security issues: [SUPPORT.md](SUPPORT.md).

## Development stack security posture

This repository targets **local development and CI**, not production deployment as-is:

- **No production authentication** — Phase 3 uses mock principals only (ADR-009 D20); Cedar Service verifies `Authorization: Bearer` HS256 JWT (`sub` + `roles` + required `exp`; optional `iss`/`aud` via env). Not legacy `X-Principal` headers. Optional `INTERNAL_API_TOKEN` gates alert/decision/graph APIs when set.
- **Default credentials** in Docker Compose (PostgreSQL, Kafka, Redis, immudb, Neo4j in dev)
- **No TLS** on local service ports
- **Kafka trust boundary** — brokers listen on PLAINTEXT; any host that can reach `:9092` can produce/consume. Treat Kafka and schema-registry as a sensitive trust zone; do not expose to untrusted networks
- **GHCR deploy** publishes application images; policy bundles are bind-mounted from the repo clone — not production-hardened ([deployment.md](docs/deployment.md))

Do not expose Compose ports to untrusted networks. Do not deploy the dev stack unchanged to production.

### Phase 4 service ports (local / deploy)

| Service | Port | Notes |
|---------|------|-------|
| graph-service | 8093 | REST API — no browser CORS |
| simulation-service | 8094 | REST API — no browser CORS |
| graph-explorer | 3003 | Next.js UI (proxies to graph-service) |
| simulation-console | 3004 | Next.js UI (proxies to simulation-service) |
| Neo4j Bolt | 7687 | Graph DB — bind to localhost on deploy stacks |
| Neo4j browser | 7474 | Admin UI — bind to localhost on deploy stacks |

On `docker-compose.deploy.yml`, Kafka, PostgreSQL, Redis, immudb, Neo4j, schema-registry, Debezium, and Grafana admin ports bind to `127.0.0.1` only.

## Secure development

- Never commit `.env` or real credentials
- Use `.env.example` for documented placeholders only
- Review PRs for secrets before merge
- Avro schema changes must remain BACKWARD compatible (see [schema-compat.yml](.github/workflows/schema-compat.yml))
