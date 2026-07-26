# Phase 5 Implementation Spec (CECT)

**Status**: Implementation in progress (CECT site delivery)  
**Authority**: [ADR-011](adr/011-phase5-reporting-foundation.md) + corporate acceptance `AC-A-P5-*`

## Scope

- Reporting Service (Python FastAPI): FINREP F01 XBRL, AnaCredit T2 SDMX, DORA ICT XML
- MinIO Object Lock artifacts + Audit `evidenceRef`
- Versioned taxonomy mapper (PostgreSQL)
- Report Console lifecycle: draft → validated → submitted
- `./scripts/smoke-test-phase5.sh`

## Layout

| Path | Role |
|------|------|
| `services/reporting-service/` | FastAPI generators + taxonomy + MinIO + audit |
| `fixtures/taxonomies/` | Fixture-equivalent taxonomy assets (D27) |
| `scripts/smoke-test-phase5.sh` | AC-A-P5-001/002/003/005 (+ API lifecycle) |
| `apps/report-console/` | Next.js UI (AC-A-P5-004) |

## Verify

```bash
cd services/reporting-service && PYTHONPATH=. pytest -q
docker compose -f docker-compose.dev.yml up -d --wait reporting-service
SMOKE_PHASE5_SKIP_PREREQS=1 ./scripts/smoke-test-phase5.sh
```

## Out of scope here

- Full commercial EBA/ECB taxonomy parity
- ClickHouse (stretch)
- Phase 6 OIDC/TLS and Phase 7 cutting-edge (see ADR-012 / ADR-013)

## Handoff

See [handoff-phase5-agent.md](handoff-phase5-agent.md).
