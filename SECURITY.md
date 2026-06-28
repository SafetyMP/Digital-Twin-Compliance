# Security Policy

## Supported versions

| Version / phase | Supported | Notes |
|-----------------|-----------|-------|
| `main` (Phase 1–3 dev stack) | Yes | Local development and CI only — not production-hardened |
| Tagged releases (`v*.*.*`) | Yes | GHCR images for all application services; see [CHANGELOG.md](CHANGELOG.md) |
| Phase 4+ (graph, simulation, full auth) | Not yet released | — |

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

- **No production authentication** — Phase 3 uses mock principals only (ADR-009 D20); no Keycloak/OIDC middleware
- **Default credentials** in Docker Compose (PostgreSQL, Kafka, Redis, immudb)
- **No TLS** on local service ports
- **GHCR deploy** publishes application images; policy bundles are bind-mounted from the repo clone — not production-hardened ([deployment.md](docs/deployment.md))

Do not expose Compose ports to untrusted networks. Do not deploy the dev stack unchanged to production.

## Secure development

- Never commit `.env` or real credentials
- Use `.env.example` for documented placeholders only
- Review PRs for secrets before merge
- Avro schema changes must remain BACKWARD compatible (see [schema-compat.yml](.github/workflows/schema-compat.yml))
