# GitHub repository setup

One-time configuration for [SafetyMP/Digital-Twin-Compliance](https://github.com/SafetyMP/Digital-Twin-Compliance). Most items are repository **Settings** (not files in git).

---

## Repository metadata

**Settings → General → Repository details**

### Topic rules (GitHub standard)

Per [GitHub Docs — Classifying your repository with topics](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/classifying-your-repository-with-topics):

- **Format**: lowercase letters, numbers, and hyphens only (no underscores or spaces)
- **Length**: ≤ 50 characters per topic
- **Count**: ≤ 20 topics per repository
- **Mix**: purpose + tech stack + industry/domain (6–15 tags is a practical target)
- **Skip** the primary language topic (`go`) — GitHub already surfaces language from code
- **Avoid** meta or version tags (`beta`, `2026`, `wip`, `release-1`)

Align topics with the README, About description, and actual dependencies — not roadmap-only tech.

### Canonical values

| Field | Value |
|-------|-------|
| Description | Open-source reference stack for an event-driven financial digital twin with embedded compliance monitoring: CDC ingestion, stream processing, policy evaluation, and a tamper-evident audit ledger. |
| Website | `https://github.com/SafetyMP/Digital-Twin-Compliance#readme` (**applied**) |
| Topics | See table below (**applied** — 14 topics) |
| Social preview | Upload [docs/assets/social-preview.png](./assets/social-preview.png) (1280×640 PNG; export from SVG if needed) — **Settings → General → Social preview** |

**Topics** (14 — purpose, stack, domain):

| Category | Topics |
|----------|--------|
| Purpose / architecture | `digital-twin`, `compliance`, `event-driven`, `transactional-outbox` |
| Industry / domain | `regtech`, `fintech`, `core-banking` |
| Tech stack | `kafka`, `debezium`, `flink`, `postgresql`, `avro`, `immudb`, `cedar` |

Apply from a maintainer shell (replaces the full topic list):

```bash
printf '%s' '{"names":["digital-twin","compliance","event-driven","transactional-outbox","regtech","fintech","core-banking","kafka","debezium","flink","postgresql","avro","immudb","cedar"]}' \
  | gh api -X PUT repos/SafetyMP/Digital-Twin-Compliance/topics --input -
```

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
| Required checks | `ci`, `schema-compat`, `analyze` (CodeQL), `Scorecard analysis` (after first run) |
| Require branches to be up to date | On |
| Block force pushes | On |
| Restrict deletions | On |

After the first CodeQL run completes, the `analyze` check name appears under required checks. After the first OpenSSF Scorecard run, add `Scorecard analysis` from [.github/workflows/scorecard.yml](../.github/workflows/scorecard.yml).

If branch protection still lists `phase1`, replace it with `ci` after merging the CI job rename (Settings → Branches → edit ruleset).

**Agents and direct pushes:** enable **Require a pull request before merging** so coding agents use feature branches (`feat/...`, `chore/...`) instead of pushing to `main`. Bypass only for maintainers when intentionally hotfixing.

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

Defer until Phase 4+ deploy automation; use approval gates and separate secrets. Full Phase 1–3 stack deploy uses GHCR + `docker-compose.deploy.yml` ([deployment.md](./deployment.md)).

---

## Container registry (GHCR)

After the first **Docker Publish** workflow succeeds:

1. **Packages** → `digital-twin-compliance/*` (eight application images)
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
| OpenSSF Scorecard | Enabled via [.github/workflows/scorecard.yml](../.github/workflows/scorecard.yml); badge on [README](../README.md) after first published run |
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

## Automation summary

| File / workflow | Purpose |
|-----------------|---------|
| [ci.yml](../.github/workflows/ci.yml) | Unit tests, policy CI, eval fixtures, Compose stack, Phase 1–3 smoke, coverage gates |
| [schema-compat.yml](../.github/workflows/schema-compat.yml) | Avro BACKWARD compatibility |
| [policy-gates.yml](../.github/workflows/policy-gates.yml) | Cedar/Zen policy CI on PRs touching `policies/**` or policy services (also runs inside `ci.yml`) |
| [codeql.yml](../.github/workflows/codeql.yml) | Go security analysis (`analyze` job) |
| [eval-nightly.yml](../.github/workflows/eval-nightly.yml) | Nightly eval fixtures, harness calibration, extended smoke |
| [docker-publish.yml](../.github/workflows/docker-publish.yml) | GHCR images (state, alert, alert-console, compliance-cep, audit, cedar, decision, audit-explorer) |
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
