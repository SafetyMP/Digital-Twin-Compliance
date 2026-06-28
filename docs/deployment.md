# Deployment

Guide for publishing and running the Digital Twin Compliance Platform outside local development.

**Scope:** Docker Compose on a single host (VM or bare metal). **Local dev and CI** run Phase 1–3 (`docker-compose.dev.yml`). **GHCR deploy** publishes all eight application images and runs the full Phase 1–3 stack via [docker-compose.deploy.yml](../docker-compose.deploy.yml). Kubernetes, Flink Kubernetes Operator, and managed Kafka are future phases — see [roadmap.md](./roadmap.md), [ADR-007](./adr/007-phase1-foundation-decisions.md), and [ADR-008](./adr/008-phase2-foundation-decisions.md).

---

## GitHub DevOps overview

| Workflow | Trigger | Purpose |
|----------|---------|---------|
| [CI](../.github/workflows/ci.yml) | Push, PR | Unit tests, policy CI, eval fixtures, Compose stack, Phase 1–3 smoke, coverage gates |
| [Schema Compatibility](../.github/workflows/schema-compat.yml) | Push, PR | Avro BACKWARD compatibility |
| [Policy gates](../.github/workflows/policy-gates.yml) | PR (path-filtered) | Cedar/Zen policy CI when `policies/**` or policy services change |
| [Docker Publish](../.github/workflows/docker-publish.yml) | Push to `main`, version tags, manual | Build and push eight application images to GHCR |
| [Release](../.github/workflows/release.yml) | Tag `v*.*.*` | GitHub Release with generated notes |
| [Deploy Staging](../.github/workflows/deploy-staging.yml) | Manual | SSH deploy to staging host |
| [Eval Nightly](../.github/workflows/eval-nightly.yml) | Daily schedule, manual | Eval fixture regression, harness calibration, extended smoke |
| [CodeQL](../.github/workflows/codeql.yml) | Push, PR, weekly | Go security analysis |

- Dependabot opens weekly PRs for Go (all services), npm (UIs), Maven (CEP), GitHub Actions, and Docker base images ([dependabot.yml](../.github/dependabot.yml)).

---

## Container registry (GHCR)

Images are published under `ghcr.io/safetymp/digital-twin-compliance/`:

| Image | Path |
|-------|------|
| `state-service` | Phase 1 REST + consumer |
| `alert-service` | Phase 2 alerts REST + WebSocket |
| `alert-console` | Phase 2 Next.js UI |
| `compliance-cep` | Phase 2 Flink job runtime |
| `audit-service` | Phase 3 immudb ledger + REST API |
| `cedar-service` | Phase 3 Cedar policy evaluate |
| `decision-service` | Phase 3 GoRules Zen evaluate |
| `audit-explorer` | Phase 3 Audit Explorer UI |

Example pull:

```text
ghcr.io/safetymp/digital-twin-compliance/state-service:main
```

| Event | Tags |
|-------|------|
| Push to `main` | `main`, `sha-<commit>` |
| Tag `v1.2.3` | `1.2.3`, `1.2`, `latest` |
| Manual dispatch | Optional custom tag |

### Pull an image

```bash
docker pull ghcr.io/safetymp/digital-twin-compliance/state-service:main
```

If the package is private, authenticate first:

```bash
echo "$GITHUB_TOKEN" | docker login ghcr.io -u USERNAME --password-stdin
```

Make the package **public** under GitHub → Packages → Package settings if staging hosts should pull without credentials.

---

## Compose files

| File | Use case |
|------|----------|
| [docker-compose.dev.yml](../docker-compose.dev.yml) | Local development; builds services from source |
| [docker-compose.deploy.yml](../docker-compose.deploy.yml) | Staging/production-like; pulls images from GHCR |

Deploy Compose requires image variables (same tag for all services is typical):

```bash
TAG=main
PREFIX=ghcr.io/safetymp/digital-twin-compliance
export STATE_SERVICE_IMAGE=${PREFIX}/state-service:${TAG}
export ALERT_SERVICE_IMAGE=${PREFIX}/alert-service:${TAG}
export ALERT_CONSOLE_IMAGE=${PREFIX}/alert-console:${TAG}
export COMPLIANCE_CEP_IMAGE=${PREFIX}/compliance-cep:${TAG}
export AUDIT_SERVICE_IMAGE=${PREFIX}/audit-service:${TAG}
export CEDAR_SERVICE_IMAGE=${PREFIX}/cedar-service:${TAG}
export DECISION_SERVICE_IMAGE=${PREFIX}/decision-service:${TAG}
export AUDIT_EXPLORER_IMAGE=${PREFIX}/audit-explorer:${TAG}
docker compose -f docker-compose.deploy.yml up -d --wait
```

Policy bundles (`policies/cedar`, `policies/zen`) are bind-mounted from the repo clone on the host — keep the checkout in sync with image tags when policies change.

Or use the helper script:

```bash
export STATE_SERVICE_IMAGE=ghcr.io/safetymp/digital-twin-compliance/state-service:main
export ALERT_SERVICE_IMAGE=ghcr.io/safetymp/digital-twin-compliance/alert-service:main
export ALERT_CONSOLE_IMAGE=ghcr.io/safetymp/digital-twin-compliance/alert-console:main
export COMPLIANCE_CEP_IMAGE=ghcr.io/safetymp/digital-twin-compliance/compliance-cep:main
export AUDIT_SERVICE_IMAGE=ghcr.io/safetymp/digital-twin-compliance/audit-service:main
export CEDAR_SERVICE_IMAGE=ghcr.io/safetymp/digital-twin-compliance/cedar-service:main
export DECISION_SERVICE_IMAGE=ghcr.io/safetymp/digital-twin-compliance/decision-service:main
export AUDIT_EXPLORER_IMAGE=ghcr.io/safetymp/digital-twin-compliance/audit-explorer:main
./scripts/deploy-stack.sh bootstrap   # first-time: up + seed + schemas + debezium
./scripts/deploy-stack.sh pull        # rolling update of all deployed images
./scripts/deploy-stack.sh smoke       # Phase 1 + 2 + 3 smoke tests against running stack
```

---

## Staging host setup

### Prerequisites on the host

- Docker Engine + Docker Compose v2
- Git
- `curl`, `jq`, `psql` (same as local dev)
- Outbound access to `ghcr.io` (for image pull)

### One-time bootstrap

```bash
git clone https://github.com/SafetyMP/Digital-Twin-Compliance.git
cd Digital-Twin-Compliance
cp .env.example .env

export STATE_SERVICE_IMAGE=ghcr.io/safetymp/digital-twin-compliance/state-service:main
export ALERT_SERVICE_IMAGE=ghcr.io/safetymp/digital-twin-compliance/alert-service:main
export ALERT_CONSOLE_IMAGE=ghcr.io/safetymp/digital-twin-compliance/alert-console:main
export COMPLIANCE_CEP_IMAGE=ghcr.io/safetymp/digital-twin-compliance/compliance-cep:main
export AUDIT_SERVICE_IMAGE=ghcr.io/safetymp/digital-twin-compliance/audit-service:main
export CEDAR_SERVICE_IMAGE=ghcr.io/safetymp/digital-twin-compliance/cedar-service:main
export DECISION_SERVICE_IMAGE=ghcr.io/safetymp/digital-twin-compliance/decision-service:main
export AUDIT_EXPLORER_IMAGE=ghcr.io/safetymp/digital-twin-compliance/audit-explorer:main
./scripts/deploy-stack.sh bootstrap
./scripts/deploy-stack.sh smoke
```

### Rolling update

After a new image is published to GHCR:

```bash
export STATE_SERVICE_IMAGE=ghcr.io/safetymp/digital-twin-compliance/state-service:main
export ALERT_SERVICE_IMAGE=ghcr.io/safetymp/digital-twin-compliance/alert-service:main
export ALERT_CONSOLE_IMAGE=ghcr.io/safetymp/digital-twin-compliance/alert-console:main
export COMPLIANCE_CEP_IMAGE=ghcr.io/safetymp/digital-twin-compliance/compliance-cep:main
export AUDIT_SERVICE_IMAGE=ghcr.io/safetymp/digital-twin-compliance/audit-service:main
export CEDAR_SERVICE_IMAGE=ghcr.io/safetymp/digital-twin-compliance/cedar-service:main
export DECISION_SERVICE_IMAGE=ghcr.io/safetymp/digital-twin-compliance/decision-service:main
export AUDIT_EXPLORER_IMAGE=ghcr.io/safetymp/digital-twin-compliance/audit-explorer:main
./scripts/deploy-stack.sh pull
```

---

## GitHub Actions staging deploy

Configure a GitHub **Environment** named `staging` (Settings → Environments) with these secrets:

| Secret | Description |
|--------|-------------|
| `DEPLOY_HOST` | Staging server hostname or IP |
| `DEPLOY_USER` | SSH user (must run Docker) |
| `DEPLOY_SSH_KEY` | Private key (PEM) for SSH |
| `DEPLOY_PATH` | Absolute path to repo clone on the host |
| `DEPLOY_PORT` | Optional SSH port (default 22) |

Run **Deploy Staging** from Actions → workflow dispatch:

- **pull** — fetch latest repo, pull all GHCR images, restart stack
- **bootstrap** — full stack up + seed + schema/connector registration (use on first deploy or after infra reset)

---

## Releases

Create a semver tag to publish a release image and GitHub Release:

```bash
git tag v0.1.0
git push origin v0.1.0
```

This triggers:

1. **Docker Publish** — tags `0.1.0`, `0.1`, `latest`
2. **Release** — GitHub Release with auto-generated notes

Deploy a release:

```bash
TAG=v0.1.0
PREFIX=ghcr.io/safetymp/digital-twin-compliance
export STATE_SERVICE_IMAGE=${PREFIX}/state-service:${TAG}
export ALERT_SERVICE_IMAGE=${PREFIX}/alert-service:${TAG}
export ALERT_CONSOLE_IMAGE=${PREFIX}/alert-console:${TAG}
export COMPLIANCE_CEP_IMAGE=${PREFIX}/compliance-cep:${TAG}
export AUDIT_SERVICE_IMAGE=${PREFIX}/audit-service:${TAG}
export CEDAR_SERVICE_IMAGE=${PREFIX}/cedar-service:${TAG}
export DECISION_SERVICE_IMAGE=${PREFIX}/decision-service:${TAG}
export AUDIT_EXPLORER_IMAGE=${PREFIX}/audit-explorer:${TAG}
./scripts/deploy-stack.sh pull
```

---

## Production readiness

This project is an **open-source reference implementation**. The default stacks are for **local development, demos, and CI** — not production as-is.

| Gap | Today | Path forward |
|-----|-------|--------------|
| Authentication | Mock principals only | OIDC / Keycloak ([ROADMAP.md](../ROADMAP.md)) |
| TLS | Plain HTTP on Compose ports | Reverse proxy or ingress in deploy stack |
| Secrets | `.env` / Compose defaults | GitHub Environments, vault, or cloud secret manager |
| Policy bundles on deploy | Bind-mounted from repo clone (`policies/cedar`, `policies/zen`) | Bake into images or ConfigMaps in K8s |
| HA / scaling | Single-host Compose | K8s / managed services ([docs/roadmap.md](./roadmap.md)) |

Before exposing any environment to untrusted networks, read [SECURITY.md](../SECURITY.md) and [SUPPORT.md](../SUPPORT.md).

---

| Concern | Where it runs |
|---------|----------------|
| Unit + integration smoke | GitHub Actions CI on every PR (`smoke-test.sh`, `smoke-test-phase2.sh`, `smoke-test-phase3.sh`) |
| Policy CI | Full CI always; [policy-gates.yml](../.github/workflows/policy-gates.yml) also on path-filtered PRs |
| Image build | Docker Publish on merge to `main` (eight application images) |
| Staging deploy | Manual Deploy Staging workflow |
| Production | Not defined — extend with environments + approval gates in a later phase |

---

## Security notes

- Deploy stacks use **default dev credentials** in Compose — not production-safe.
- Do not expose ports 5433–5436, 3322, 6380, 9092, 8080–8092, 3000–3002 to the public internet without TLS, auth, and secret rotation.
- Store real credentials in GitHub Environment secrets or a secrets manager; never commit `.env`.
- Review [SECURITY.md](../SECURITY.md) before exposing any environment.

---

## Next.js UIs and Go APIs

`alert-console` (`:3000`) and `audit-explorer` (`:3002`) must **not** call Go services on other ports from browser `fetch` — there is no CORS on `alert-service`, `audit-service`, etc.

| UI | Browser calls | Compose env (server-side) |
|----|---------------|---------------------------|
| Alert Console | `/api/alerts`, `/api/alerts/{id}/acknowledge` | `ALERT_SERVICE_URL=http://alert-service:8085` |
| Audit Explorer | `/api/audit/entries`, `/api/audit/verify` | `AUDIT_SERVICE_URL=http://audit-service:8090` |

Cross-links (`NEXT_PUBLIC_AUDIT_EXPLORER_URL`) are navigation `href` only. Live alert feed uses **polling** (not browser WebSocket to `:8085`).

Flink CEP on deploy calls Decision Service when `CEP_DECISION_SERVICE_URL` is set (default in `docker-compose.deploy.yml`).

---

## Troubleshooting

| Symptom | Check |
|---------|--------|
| `STATE_SERVICE_IMAGE` unset | Export image URL before `docker compose -f docker-compose.deploy.yml` |
| Image pull 401/403 | `docker login ghcr.io` or make GHCR package public |
| Personas not syncing | Re-run `./scripts/register-debezium-connector.sh` and restart `state-service` |
| Smoke test timeout | Wait for initial CDC snapshot; check Debezium connector status at `:8083/connectors` |
| Phase 2 smoke fails on Flink | Confirm job RUNNING at `:8082`; check `flink-job-submitter` logs (deploy) or re-run `./scripts/submit-flink-job.sh` (dev) |
| No alerts on `compliance.alerts` | Check Flink logs; verify Redis at `localhost:6380`; confirm payment seed / burst simulator |
| Alert Console empty but API has data | Rebuild `alert-console` image; ensure `ALERT_SERVICE_URL=http://alert-service:8085` in Compose (not `NEXT_PUBLIC_*` to `:8085`) |
| WebSocket ack not received (smoke script) | Set `ALERT_SERVICE_WS_URL=ws://localhost:8085/ws/alerts`; verify `alert-service` health at `:8085` |
| Phase 3 smoke fails on audit chain | Verify `audit-service` at `:8090`; check `compliance.audit.pending` consumer; run `./scripts/verify-audit-chain.sh` |
| Cedar/Zen policies empty in container | `git pull` policies on host; restart `cedar-service` and `decision-service` (bind mount) |
| `*_IMAGE` unset | Export all eight image variables before `docker compose -f docker-compose.deploy.yml` |

For local development issues, see [README.md](../README.md#quick-start).
