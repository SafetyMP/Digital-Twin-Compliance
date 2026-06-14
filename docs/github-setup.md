# GitHub repository setup

One-time configuration for [SafetyMP/Digital-Twin-Compliance](https://github.com/SafetyMP/Digital-Twin-Compliance). Most items are repository **Settings** (not files in git).

---

## Repository metadata

**Settings → General → Repository details**

| Field | Suggested value |
|-------|-----------------|
| Description | Event-driven financial digital twin with compliance monitoring (Phase 1 ingestion backbone) |
| Website | (optional) link to docs or demo |
| Topics | `digital-twin`, `compliance`, `kafka`, `debezium`, `go`, `event-driven`, `postgresql`, `avro` |

Enable **Issues** and **Discussions** (optional) under Features.

---

## Branch protection (main)

**Settings → Branches → Add branch ruleset** (or classic branch protection rule)

Recommended for `main`:

| Rule | Setting |
|------|---------|
| Require a pull request before merging | On (1 approval if team grows) |
| Require status checks to pass | On |
| Required checks | `phase1` (CI), `schema-compat`, `analyze` (CodeQL) |
| Require branches to be up to date | On |
| Block force pushes | On |
| Restrict deletions | On |

After the first CodeQL run completes, the `analyze` check name appears under required checks.

---

## GitHub Environments

### staging

**Settings → Environments → New environment → `staging`**

| Secret | Description |
|--------|-------------|
| `DEPLOY_HOST` | Staging server hostname or IP |
| `DEPLOY_USER` | SSH user with Docker access |
| `DEPLOY_SSH_KEY` | Private key (PEM) |
| `DEPLOY_PATH` | Absolute path to repo clone on host |
| `DEPLOY_PORT` | Optional SSH port (default 22) |

Optional protection rules: required reviewers before deploy, wait timer.

### production (future)

Defer until Phase 2+; use approval gates and separate secrets.

---

## Container registry (GHCR)

After the first **Docker Publish** workflow succeeds:

1. **Packages** → `digital-twin-compliance/state-service`
2. **Package settings → Change visibility** → Public (if staging pulls without auth)
3. Add a short **description** on the package page linking to [deployment.md](./deployment.md)

---

## Security features

**Settings → Code security and analysis**

| Feature | Recommendation |
|---------|----------------|
| Dependabot alerts | Enable |
| Dependabot security updates | Enable |
| Code scanning (CodeQL) | Enabled via [.github/workflows/codeql.yml](../.github/workflows/codeql.yml) |
| Secret scanning | Enable (GitHub-provided patterns) |

---

## First release

Validate release and image pipelines:

```bash
git tag v0.1.0
git push origin v0.1.0
```

Expect:

- **Docker Publish** — tags `0.1.0`, `0.1`, `latest`
- **Release** — GitHub Release with generated notes

Deploy:

```bash
export STATE_SERVICE_IMAGE=ghcr.io/safetymp/digital-twin-compliance/state-service:v0.1.0
./scripts/deploy-stack.sh pull
```

---

## Automation summary

| File / workflow | Purpose |
|-----------------|---------|
| [ci.yml](../.github/workflows/ci.yml) | Full stack CI |
| [schema-compat.yml](../.github/workflows/schema-compat.yml) | Avro compatibility |
| [codeql.yml](../.github/workflows/codeql.yml) | Security analysis |
| [docker-publish.yml](../.github/workflows/docker-publish.yml) | GHCR images |
| [release.yml](../.github/workflows/release.yml) | Version releases |
| [deploy-staging.yml](../.github/workflows/deploy-staging.yml) | SSH staging deploy |
| [dependabot.yml](../.github/dependabot.yml) | Dependency PRs |
| [CODEOWNERS](../.github/CODEOWNERS) | Default reviewers |
| Issue templates | Bug and feature intake |

---

## Community files

| File | Purpose |
|------|---------|
| [CODE_OF_CONDUCT.md](../CODE_OF_CONDUCT.md) | Community standards |
| [CONTRIBUTING.md](../CONTRIBUTING.md) | Contribution workflow |
| [SECURITY.md](../SECURITY.md) | Vulnerability reporting |
| [NOTICE](../NOTICE) | Apache 2.0 attribution |
