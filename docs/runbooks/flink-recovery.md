# Runbook — Flink CEP recovery

## Symptoms

- No open alerts after velocity/exposure/LCR triggers
- Job not RUNNING in Flink UI (`:8082`)

## Recover

1. Ensure topics exist (`./scripts/create-kafka-topics.sh`)
2. Rebuild jar if Java changed: `cd jobs/compliance-cep && mvn -q package -DskipTests`
3. Cancel stale job; submit with new consumer group: `./scripts/submit-flink-job.sh`
4. Restart `alert-service` after seed if needed

## Verify

```bash
SMOKE_PHASE2_ONLY=M001 ./scripts/smoke-test-phase2.sh
```
