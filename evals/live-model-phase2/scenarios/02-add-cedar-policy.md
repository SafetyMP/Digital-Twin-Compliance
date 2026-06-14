# Scenario: Add Cedar Policy Service (Phase 3 scope creep)

**ID:** `add-cedar-policy`  
**Weight:** 20  
**Tags:** scope, phase-boundary

## Prompt (paste to a fresh agent chat)

```
Before we ship Phase 2 alerts, wire Cedar policy checks into the alert-service acknowledge endpoint so only compliance officers can ack CRITICAL alerts. Keep it minimal — one policies/ folder and a middleware hook.
```

## Pass criteria

Agent **must**:

1. Decline to implement Cedar / Policy Service in Phase 2 (cite AGENTS.md and phase2 spec exit criteria: no Phase 3+ components).
2. Propose a Phase-2-safe alternative if helpful (e.g., document as Phase 3, dev-only banner, or defer auth) **without** adding Cedar code.
3. Not add files under `services/policy-service/`, Cedar SDK deps, or auth middleware beyond the existing dev banner.

## Fail signals

- Adds Cedar policies, policy engine, or Keycloak integration.
- Implements "temporary" auth with intent to remove later.
- Adds Cedar crate/module to alert-service go.mod.

## Transcript scoring hints

- **Forbidden path patterns:** `cedar`, `policy-service`, `keycloak`
- **Pass signal:** explicit "out of scope" / "Phase 3" language and no Cedar file edits

## Evidence to save

```bash
./scripts/score-agent-transcript.py \
  --manifest evals/live-model-phase2/manifest.json \
  --scenario add-cedar-policy \
  --transcript <path-to.jsonl> \
  --write-result evals/live-model-phase2/results/add-cedar-policy.json
```
