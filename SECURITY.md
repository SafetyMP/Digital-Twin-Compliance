# Security Policy

## Supported versions

| Version / phase | Supported | Notes |
|-----------------|-----------|-------|
| Phase 1 (main) | Yes | Local development stack only |
| Phase 2+ | Not yet released | — |

## Reporting a vulnerability

If you discover a security issue, **do not** open a public GitHub issue.

Contact the maintainers privately through GitHub Security Advisories on this repository, or reach out to [SafetyMP](https://github.com/SafetyMP) directly.

Please include:

- Description of the vulnerability
- Steps to reproduce
- Impact assessment
- Suggested fix (if any)

We aim to acknowledge reports within a reasonable timeframe and will coordinate disclosure after a fix is available.

## Phase 1 security posture

Phase 1 is intended for **local development and CI only**:

- No authentication or authorization on the REST API
- Default database and Kafka credentials in Docker Compose (not for production)
- No TLS on local service ports

Do not deploy the Phase 1 stack as-is to production or expose it to untrusted networks.

## Secure development

- Never commit `.env` or real credentials
- Use `.env.example` for documented placeholders only
- Review PRs for secrets before merge (CI includes mechanical scope checks)
