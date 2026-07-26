# Runbook — immudb / audit recovery

## Symptoms

- Audit verify fails / sequence gaps
- `audit-service` unhealthy after volume wipe

## Recover

1. If PostgreSQL `audit_entry_index` was truncated but immudb was not, restart
   `audit-service` (it resets immudb head when the PG index is empty).
2. Full reset (dev only):  
   `docker compose -f docker-compose.dev.yml stop audit-service && docker compose -f docker-compose.dev.yml rm -f -v immudb audit-db && docker compose -f docker-compose.dev.yml up -d --wait audit-service`
3. Re-run a producer (alert ack, simulation, or report submit) to append entries.

## Verify

```bash
./scripts/verify-audit-chain.sh
```
