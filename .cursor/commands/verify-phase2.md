---
description: Run Phase 2 verification — smoke tests first, no transcript archaeology.
---
Verify Phase 2 integration per [docs/handoff-verification-agent.md](../../docs/handoff-verification-agent.md).

**Session hygiene**: Use a **fresh chat** if this thread already did analysis or long implementation. Do not Read `agent-transcripts/`, `evals/live-model/README.md`, or superpowers skills.

**Context budget**: [AGENTS.md](../../AGENTS.md) + [services/alert-service/AGENTS.md](../../services/alert-service/AGENTS.md) only if alert pipeline fails.

Run sequentially (evidence before claiming done):

```bash
docker compose -f docker-compose.dev.yml up -d --wait
./scripts/seed.sh
# Flink RUNNING? curl -sf http://localhost:8082/jobs | jq '.jobs[] | select(.status=="RUNNING")'
# If no RUNNING job: FLINK_JOBMANAGER_URL=http://localhost:8082 ./scripts/submit-flink-job.sh
./scripts/smoke-test.sh
./scripts/smoke-test-phase2.sh
./scripts/token-efficiency.sh --strict
```

If either smoke test or efficiency check fails: minimal focused diffs, re-run until all exit 0.

Return: failures found, fixes applied, exit codes, efficiency metrics (`harness_reread_count: 0`, `duplicate_read_count ≤ 3`), and relevant command output.
