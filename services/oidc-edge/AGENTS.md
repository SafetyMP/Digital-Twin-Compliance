# OIDC Edge — Agent Contract

Phase 6 hardening reverse proxy: validates Keycloak Bearer JWTs, strips client
`X-Principal`, injects trusted principal/roles, proxies to public APIs.

Used with `docker-compose.hardening.yml`. Direct service ports stay open for
Phase 1–4 smokes.

## Routes

| Prefix | Upstream |
|--------|----------|
| `/state/` | state-service:8080 |
| `/alert/` | alert-service:8085 |
| `/audit/` | audit-service:8090 |
| `/graph/` | graph-service:8093 |
| `/simulation/` | simulation-service:8094 |
| `/reporting/` | reporting-service:8095 |

`/api/v1/health` on the edge is open. Upstream `*/api/v1/health` is open through the edge; all other paths require Bearer.
