# Support

The Digital Twin Compliance Platform is an **open-source project maintained by [SafetyMP](https://github.com/SafetyMP)** under the [Apache License 2.0](LICENSE).

## What we support

| Channel | Use for | Response expectation |
|---------|---------|----------------------|
| [GitHub Issues](https://github.com/SafetyMP/Digital-Twin-Compliance/issues) | Bugs, regressions, feature ideas | Best effort; no SLA |
| [Security advisories](https://github.com/SafetyMP/Digital-Twin-Compliance/security/advisories/new) | Vulnerabilities (private) | Acknowledgement when possible; coordinated disclosure |
| Pull requests | Fixes and enhancements you can implement | Review when maintainer capacity allows |

## What we do not support

- **Production deployments** of the dev Compose stack as-is (default credentials, no TLS, mock auth)
- **Consulting or custom integration** through this repository — use Issues only for project scope
- **Guaranteed timelines** for roadmap items ([ROADMAP.md](ROADMAP.md))

## Before opening an issue

1. Run on latest `main` (or note your release tag)
2. Include reproduction commands and smoke script output
3. Check [README quick start](README.md#quick-start) and [docs/deployment.md](docs/deployment.md)

## Supported versions

| Branch / tag | Support |
|--------------|---------|
| `main` | Active development; CI must pass before merge |
| Latest `v*.*.*` release | Recommended for deploy experiments using GHCR images |
| Older tags | Best effort only |

Security details: [SECURITY.md](SECURITY.md).
