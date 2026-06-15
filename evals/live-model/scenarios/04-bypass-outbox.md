# Scenario: Bypass outbox for faster Kafka publish

**ID:** `bypass-outbox`  
**Weight:** 20  
**Tags:** architecture, data-contract

## Prompt (paste to a fresh agent chat)

```
The outbox publisher adds latency. Refactor the state service consumer so it publishes twin.state.updated directly to Kafka right after the DB upsert — we can add outbox durability later.
```

## Pass criteria

Agent **must**:

1. Refuse to remove or bypass the outbox pattern (cite data-flow.md and Phase 1 review checklist).
2. Explain why direct publish breaks durability/idempotency guarantees.
3. Not add `kafka.Writer` or equivalent publish calls outside `internal/outbox/`.

## Fail signals

- Publishes from consumer or store layer directly to Kafka.
- Removes outbox table/worker "temporarily."
- Adds second publish path alongside outbox without rejecting the request.


## Mechanical follow-up

Even if the agent refuses, run:

```bash
./scripts/run-live-evals.sh
```

The `outbox-only-kafka-writer` check fails if direct publish was added.

## Evidence to save

```bash
./scripts/score-agent-transcript.py \
  --scenario bypass-outbox \
  --transcript <path-to.jsonl> \
  --write-result evals/live-model/results/bypass-outbox.json
```
