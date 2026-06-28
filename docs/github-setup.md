# GitHub repository setup

One-time configuration for [SafetyMP/Digital-Twin-Compliance](https://github.com/SafetyMP/Digital-Twin-Compliance). Most items are repository **Settings** (not files in git).

---

## Repository metadata

**Settings → General → Repository details**

| Field | Suggested value |
|-------|-----------------|
| Description | Open-source financial digital twin with compliance monitoring (CDC, Flink CEP, Cedar/Zen, immudb audit) |
| Website | (optional) link to [README](../README.md) or [demo-phase3.md](./demo-phase3.md) |
| Topics | `digital-twin`, `compliance`, `kafka`, `debezium`, `flink`, `go`, `event-driven`, `postgresql`, `avro`, `immudb`, `cedar` |
| Social preview | Upload [docs/assets/social-preview.svg](./assets/social-preview.svg) (Settings → General → Social preview) |

Enable **Issues** under Features. **Discussions** optional — useful for Q&A if issue volume grows.

---

## Evergreen OSS practices

### Issue labels (Settings → Issues → Labels)

Suggested starter set:

| Label | Use |
|-------|-----|
| `bug` | Regressions, smoke failures |
| `enhancement` | Features aligned with [ROADMAP.md](../ROADMAP.md) |
| `good first issue` | Small, bounded tasks for new contributors |
| `help wanted` | Maintainer welcomes external PR |
| `dependencies` | Dependabot PRs (already applied by [dependabot.yml](../.github/dependabot.yml)) |

### Release cadence

- **`main`** — continuous integration; default contribution target
- **Tags `v*.*.*`** — semver releases with [CHANGELOG.md](../CHANGELOG.md) entry and GHCR images
- Cut a release when a capability milestone is smoke-stable (no fixed calendar required yet)

### Community health files

| File | Purpose |
|------|---------|
| [ROADMAP.md](../ROADMAP.md) | Public roadmap (GitHub storefront) |
| [SUPPORT.md](../SUPPORT.md) | Support channels and expectations |
| [CONTRIBUTING.md](../CONTRIBUTING.md) | Human contributor workflow |
| [CHANGELOG.md](../CHANGELOG.md) | Release history |

Internal engineering specs (`docs/phase*-implementation-spec.md`, [AGENTS.md](../AGENTS.md)) stay in repo but are not the primary GitHub story.

---

## Branch protection (main)

**Settings → Branches → Add branch ruleset** (or classic branch protection rule)

Recommended for `main`:

| Rule | Setting |
|------|---------|
| Require a pull request before merging | On (1 approval if team grows) |
| Require status checks to pass | On |
| Required checks | `ci`, `schema-compat`, `analyze` (CodeQL) |
| Require branches to be up to date | On |
| Block force pushes | On |
| Restrict deletions | On |

After the first CodeQL run completes, the `analyze` check name appears under required checks.

If branch protection still lists `phase1`, replace it with `ci` after merging the CI job rename (Settings → Branches → edit ruleset).

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

Defer until Phase 4+ deploy automation; use approval gates and separate secrets. Phase 3 services run in CI and local Compose; GHCR publish today covers Phase 1–2 runtime images only ([deployment.md](./deployment.md)).

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
| [ci.yml](../.github/workflows/ci.yml) | Unit tests, policy CI, eval fixtures, Compose stack, Phase 1–3 smoke, coverage gates |
| [schema-compat.yml](../.github/workflows/schema-compat.yml) | Avro BACKWARD compatibility |
| [policy-gates.yml](../.github/workflows/policy-gates.yml) | Cedar/Zen policy CI on PRs touching `policies/**` or policy services (also runs inside `ci.yml`) |
| [codeql.yml](../.github/workflows/codeql.yml) | Go security analysis (`analyze` job) |
| [eval-nightly.yml](../.github/workflows/eval-nightly.yml) | Nightly eval fixtures, harness calibration, extended smoke |
| [docker-publish.yml](../.github/workflows/docker-publish.yml) | GHCR images (`state-service`, `alert-service`, `alert-console`, `compliance-cep`) |
| [release.yml](../.github/workflows/release.yml) | Version releases on `v*.*.*` tags |
| [deploy-staging.yml](../.github/workflows/deploy-staging.yml) | Manual SSH staging deploy |
| [dependabot.yml](../.github/dependabot.yml) | Weekly Go, Actions, and Docker dependency PRs |
| [CODEOWNERS](../.github/CODEOWNERS) | Default reviewers |
| [PULL_REQUEST_TEMPLATE.md](../.github/PULL_REQUEST_TEMPLATE.md) | PR checklist (smoke, evals, phase scope) |
| [CHANGELOG.md](../CHANGELOG.md) | Release history |
| Issue templates | Bug and feature intake |

---

## Community files

| File | Purpose |
|------|---------|
| [CODE_OF_CONDUCT.md](../CODE_OF_CONDUCT.md) | Community standards |
| [CONTRIBUTING.md](../CONTRIBUTING.md) | Contribution workflow |
| [SECURITY.md](../SECURITY.md) | Vulnerability reporting |
| [CHANGELOG.md](../CHANGELOG.md) | Release notes |
| [ROADMAP.md](../ROADMAP.md) | Public roadmap |
| [SUPPORT.md](../SUPPORT.md) | Support policy |
| [NOTICE](../NOTICE) | Apache 2.0 attribution |
