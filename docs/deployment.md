# Deployment

Guide for publishing and running the Digital Twin Compliance Platform outside local development.

**Phase 1 scope:** Docker Compose on a single host (VM or bare metal). Kubernetes and managed Kafka are future phases — see [roadmap.md](./roadmap.md) and [ADR-007](./adr/007-phase1-foundation-decisions.md).

---

## GitHub DevOps overview

| Workflow | Trigger | Purpose |
|----------|---------|---------|
| [CI](../.github/workflows/ci.yml) | Push, PR | Full stack tests + smoke test |
| [Schema Compatibility](../.github/workflows/schema-compat.yml) | Push, PR | Avro BACKWARD compatibility |
| [Docker Publish](../.github/workflows/docker-publish.yml) | Push to `main`, version tags, manual | Build and push State Service to GHCR |
| [Release](../.github/workflows/release.yml) | Tag `v*.*.*` | GitHub Release with generated notes |
| [Deploy Staging](../.github/workflows/deploy-staging.yml) | Manual | SSH deploy to staging host |

Dependabot opens weekly PRs for Go modules, GitHub Actions, and Docker base images ([dependabot.yml](../.github/dependabot.yml)).

---

## Container registry (GHCR)

State Service images are published to:

```text
ghcr.io/safetymp/digital-twin-compliance/state-service
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
| [docker-compose.dev.yml](../docker-compose.dev.yml) | Local development; builds State Service from source |
| [docker-compose.deploy.yml](../docker-compose.deploy.yml) | Staging/production-like; pulls State Service from GHCR |

Deploy Compose requires `STATE_SERVICE_IMAGE`:

```bash
export STATE_SERVICE_IMAGE=ghcr.io/safetymp/digital-twin-compliance/state-service:main
docker compose -f docker-compose.deploy.yml up -d --wait
```

Or use the helper script:

```bash
export STATE_SERVICE_IMAGE=ghcr.io/safetymp/digital-twin-compliance/state-service:main
./scripts/deploy-stack.sh bootstrap   # first-time: up + seed + schemas + debezium
./scripts/deploy-stack.sh pull        # rolling update of state-service only
./scripts/deploy-stack.sh smoke       # run smoke test against running stack
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
| Unit + integration smoke | GitHub Actions CI on every PR |
| Image build | Docker Publish on merge to `main` |
| Staging deploy | Manual Deploy Staging workflow |
| Production | Not defined in Phase 1 — extend with environments + approval gates in Phase 2+ |

---

## Security notes

- Phase 1 deploy stacks use **default dev credentials** in Compose — not production-safe.
- Do not expose ports 5433, 5434, 9092, 8080–8083 to the public internet without TLS, auth, and secret rotation.
- Store real credentials in GitHub Environment secrets or a secrets manager; never commit `.env`.
- Review [SECURITY.md](../SECURITY.md) before exposing any environment.

---

## Troubleshooting

| Symptom | Check |
|---------|--------|
| `STATE_SERVICE_IMAGE` unset | Export image URL before `docker compose -f docker-compose.deploy.yml` |
| Image pull 401/403 | `docker login ghcr.io` or make GHCR package public |
| Personas not syncing | Re-run `./scripts/register-debezium-connector.sh` and restart `state-service` |
| Smoke test timeout | Wait for initial CDC snapshot; check Debezium connector status at `:8083/connectors` |

For local development issues, see [README.md](../README.md#quick-start).
