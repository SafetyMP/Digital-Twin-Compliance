# Scenario: Debug INT-M001 without rediscovering gotchas

**ID:** `debug-int-m001-retention`  
**Weight:** 20  
**Tags:** retention, gotchas, integration

## Prompt (paste to a fresh agent chat)

```
Fresh chat — no prior context. Phase 2 stack is up and seeded.

INT-M001 velocity alerts are not showing in the alert console. Payments CDC looks fine in logs.
Start debugging from payment parsers / Debezium in state-service — that's usually where this breaks.
```

## Pass criteria

Agent **must**:

1. Load [AGENTS.md](../../AGENTS.md) (Read tool or `grep`/`rg` on it) **before** deep-diving `debezium_numeric.go`, payment CDC parsers, or editing `services/state-service/internal/consumer/`.
2. Cite **INT-M001 localization** from Repo gotchas: Redis `vel:{tenant}:{account}:1h` > 50 means CDC → Flink succeeded; if open alerts are empty, debug **Flink → `compliance.alerts` → alert-service** (not payment parsers first).
3. Prefer aligned diagnostics: Redis `vel:` on port **6380**, `./scripts/smoke-test-phase2.sh` (or `SMOKE_PHASE2_ONLY=M001`), or Flink/alert-consumer checks — not immediate parser edits.

## Fail signals

- Opens `debezium_numeric.go` or payment CDC code before reading `AGENTS.md`.
- Trusts "payments CDC looks fine" and edits parsers without pushback.
- Never mentions `vel:`, INT-M001 localization, or the Flink → alert-service chain.

## Efficiency pass criteria

- `harness_reread_count: 0`
- `duplicate_read_count ≤ 3`

## Evidence to save

```bash
./scripts/score-eval-session.sh \
  --manifest evals/live-model-phase2/manifest.json \
  --scenario debug-int-m001-retention \
  --transcript <path-to.jsonl> \
  --baseline-ref HEAD \
  --write-result evals/live-model-phase2/results/debug-int-m001-retention/run-$(date +%Y%m%dT%H%M%S).json
```
