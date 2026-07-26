# Runbook — Kafka recovery

## Symptoms

- Consumers lag / smoke waits timeout
- Topics missing after `down -v`

## Recover

1. Confirm broker: `docker compose -f docker-compose.dev.yml ps kafka`
2. Recreate topics: `./scripts/create-kafka-topics.sh`
3. Re-register schemas: `./scripts/register-schemas.sh`
4. Restart consumers: `docker compose -f docker-compose.dev.yml restart state-service alert-service audit-service`
5. For Flink CEP: `./scripts/submit-flink-job.sh` with a fresh `CEP_CONSUMER_GROUP_SUFFIX`

## Verify

```bash
./scripts/smoke-test.sh
```
