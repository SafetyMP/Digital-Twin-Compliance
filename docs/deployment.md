# Deployment

Guide for publishing and running the Digital Twin Compliance Platform outside local development.

**Scope:** Docker Compose on a single host (VM or bare metal). **Local dev and CI** run Phase 1–3 (`docker-compose.dev.yml`). **GHCR deploy** today publishes Phase 1–2 runtime images only; Phase 3 UIs/services are dev/CI until added to [docker-compose.deploy.yml](../docker-compose.deploy.yml). Kubernetes, Flink Kubernetes Operator, and managed Kafka are future phases — see [roadmap.md](./roadmap.md), [ADR-007](./adr/007-phase1-foundation-decisions.md), and [ADR-008](./adr/008-phase2-foundation-decisions.md).

---

## GitHub DevOps overview

| Workflow | Trigger | Purpose |
|----------|---------|---------|
| [CI](../.github/workflows/ci.yml) | Push, PR | Unit tests, policy CI, eval fixtures, Compose stack, Phase 1–3 smoke, coverage gates |
| [Schema Compatibility](../.github/workflows/schema-compat.yml) | Push, PR | Avro BACKWARD compatibility |
| [Policy gates](../.github/workflows/policy-gates.yml) | PR (path-filtered) | Cedar/Zen policy CI when `policies/**` or policy services change |
| [Docker Publish](../.github/workflows/docker-publish.yml) | Push to `main`, version tags, manual | Build and push `state-service`, `alert-service`, `alert-console`, `compliance-cep` to GHCR |
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

Deploy Compose requires image variables:

```bash
export STATE_SERVICE_IMAGE=ghcr.io/safetymp/digital-twin-compliance/state-service:main
export ALERT_SERVICE_IMAGE=ghcr.io/safetymp/digital-twin-compliance/alert-service:main
export ALERT_CONSOLE_IMAGE=ghcr.io/safetymp/digital-twin-compliance/alert-console:main
export COMPLIANCE_CEP_IMAGE=ghcr.io/safetymp/digital-twin-compliance/compliance-cep:main
docker compose -f docker-compose.deploy.yml up -d --wait
```

Or use the helper script:

```bash
export STATE_SERVICE_IMAGE=ghcr.io/safetymp/digital-twin-compliance/state-service:main
./scripts/deploy-stack.sh bootstrap   # first-time: up + seed + schemas + debezium
./scripts/deploy-stack.sh pull        # rolling update of deployed images
./scripts/deploy-stack.sh smoke       # run Phase 1 + Phase 2 smoke tests against running stack
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
./scripts/deploy-stack.sh bootstrap
./scripts/deploy-stack.sh smoke
```

### Rolling update

After a new image is published to GHCR:

```bash
export STATE_SERVICE_IMAGE=ghcr.io/safetymp/digital-twin-compliance/state-service:main
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

- **pull** — fetch latest repo, pull new image, restart `state-service`
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
export STATE_SERVICE_IMAGE=ghcr.io/safetymp/digital-twin-compliance/state-service:v0.1.0
./scripts/deploy-stack.sh pull
```

---

## CI vs deploy

| Concern | Where it runs |
|---------|----------------|
| Unit + integration smoke | GitHub Actions CI on every PR (`smoke-test.sh`, `smoke-test-phase2.sh`, `smoke-test-phase3.sh`) |
| Policy CI | Full CI always; [policy-gates.yml](../.github/workflows/policy-gates.yml) also on path-filtered PRs |
| Image build | Docker Publish on merge to `main` (four Phase 1–2 service images) |
| Staging deploy | Manual Deploy Staging workflow |
| Production | Not defined — extend with environments + approval gates in a later phase |

---

## Security notes

- Deploy stacks use **default dev credentials** in Compose — not production-safe.
- Do not expose ports 5433–5435, 6380, 9092, 8080–8085, 3000–3001 to the public internet without TLS, auth, and secret rotation.
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

Phase 3 deploy images (`audit-explorer`, `audit-service`, …) are not yet in `docker-compose.deploy.yml` — use `docker-compose.dev.yml` for full Phase 3 demos.

---

## Troubleshooting

| Symptom | Check |
|---------|--------|
| `STATE_SERVICE_IMAGE` unset | Export image URL before `docker compose -f docker-compose.deploy.yml` |
| Image pull 401/403 | `docker login ghcr.io` or make GHCR package public |
| Personas not syncing | Re-run `./scripts/register-debezium-connector.sh` and restart `state-service` |
| Smoke test timeout | Wait for initial CDC snapshot; check Debezium connector status at `:8083/connectors` |
| Phase 2 smoke fails on Flink | Confirm job RUNNING at `:8082`; re-run `./scripts/submit-flink-job.sh` |
| No alerts on `compliance.alerts` | Check Flink logs; verify Redis at `localhost:6380`; confirm payment seed / burst simulator |
| Alert Console empty but API has data | Rebuild `alert-console` image; ensure `ALERT_SERVICE_URL=http://alert-service:8085` in Compose (not `NEXT_PUBLIC_*` to `:8085`). Browser uses same-origin `/api/alerts` — direct `fetch` to `:8085` fails without CORS. |
| WebSocket ack not received (smoke script) | Set `ALERT_SERVICE_WS_URL=ws://localhost:8085/ws/alerts`; verify `alert-service` health at `:8085`. UI uses 5s polling, not browser WebSocket. |
| `ALERT_*` or `COMPLIANCE_CEP_*` image unset | Export all four image variables before `docker compose -f docker-compose.deploy.yml` |

For local development issues, see [README.md](../README.md#quick-start).
