# Deployment notes (CECT Phase 6)

## Profiles

| Compose files | Purpose |
|---------------|---------|
| `docker-compose.dev.yml` | Default local/CI — direct service ports, mock/open public edges |
| `+ docker-compose.hardening.yml` | Keycloak (`:8088`) + OIDC edge (`:8180`) for Phase 6 proofs |

```bash
docker compose -f docker-compose.dev.yml -f docker-compose.hardening.yml up -d --build --wait oidc-edge
./scripts/smoke-test-phase6-oidc.sh
```

OIDC edge prefixes: `/state/`, `/alert/`, `/audit/`, `/graph/`, `/simulation/`, `/reporting/`.
Unauthenticated non-health calls return **401**. Tokens via Keycloak direct grant (`analyst` / `analyst`, client `digital-twin-api`).

Direct ports (e.g. `:8080`, `:8095`) remain available so Phase 1–4 / Phase 5 smokes do not require tokens.

## Secrets

- `.env.example` contains **non-production** placeholders only.
- Hardening profile uses documented Keycloak admin/dev client secrets — **not** production credentials.
- Residual: Hashicorp Vault (or equivalent) injection for staging/prod is deferred; list remains open.

## TLS edge

Self-signed cert under `infra/tls/` terminated by nginx (`tls-edge` on `:8443`). Private keys (`infra/tls/*.key`) are gitignored — generate locally if missing:

```bash
mkdir -p infra/tls
openssl req -x509 -nodes -newkey rsa:2048 -days 365 \
  -keyout infra/tls/edge.key -out infra/tls/edge.crt \
  -subj "/CN=localhost"
```

```bash
curl -k https://localhost:8443/healthz
curl -k -H "Authorization: Bearer $TOKEN" https://localhost:8443/reporting/api/v1/taxonomies
```

## Residuals (honest)

- Full multi-tenant supervisory SaaS
- HA immudb / multi-AZ Kafka
- Full Kubernetes/Terraform reference pack
- Commercial EBA taxonomy parity
- 10K evt/s on default CI hosts (see load profile docs when authored)
