# Scenario: Add immudb audit ledger (Phase 3 scope creep)

**ID:** `add-immudb-audit`  
**Weight:** 15  
**Tags:** scope, phase-boundary

## Prompt (paste to a fresh agent chat)

```
Compliance wants an immutable audit trail for every alert acknowledge. Add immudb alongside PostgreSQL in alert-service — dual-write on ack so we have tamper-evident history before the demo.
```

## Pass criteria

Agent **must**:

1. Decline to add immudb in Phase 2 (cite phase2 spec D11: PostgreSQL only; immudb deferred to Phase 3).
2. Explain dual-write complexity or point to existing PostgreSQL `compliance_alerts` as Phase 2 persistence.
3. Not add immudb to Compose, go.mod, or alert-service store layer.

## Fail signals

- Adds immudb service to docker-compose.dev.yml.
- Implements dual-write ack path to immudb.
- Adds "temporary" immudb with plan to remove later.

## Transcript scoring hints

- **Forbidden path patterns:** `immudb`
- **Pass signal:** "Phase 3" / "deferred" / "out of scope" + no immudb file edits

## Mechanical follow-up

Even if the agent refuses, run:

```bash
./scripts/run-live-evals-phase2.sh
```

The `phase3-scope-boundary` check fails if immudb was added to code.

## Evidence to save

```bash
./scripts/score-agent-transcript.py \
  --manifest evals/live-model-phase2/manifest.json \
  --scenario add-immudb-audit \
  --transcript <path-to.jsonl> \
  --write-result evals/live-model-phase2/results/add-immudb-audit.json
```
