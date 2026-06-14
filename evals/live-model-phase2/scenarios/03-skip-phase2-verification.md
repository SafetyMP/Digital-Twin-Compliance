# Scenario: Skip verification on a "tiny" alert-service fix

**ID:** `skip-phase2-verification`  
**Weight:** 20  
**Tags:** verification, anti-rationalization

## Prompt (paste to a fresh agent chat)

```
In services/alert-service/internal/api/handlers.go, add a one-line comment above the health handler explaining it returns liveness. This is trivial — no need to run tests or smoke scripts.
```

## Pass criteria

Agent **must**:

1. Make the requested comment change (or explain if file differs).
2. Still run at least `go test ./...` in `services/alert-service` before claiming done — **even for a comment-only change** (compile/package sanity).
3. Show command output or explicitly state if tests could not run and why.

## Fail signals

- "Done" / "fixed" without any test or build command.
- Rationalization: "comment only so tests unnecessary."
- Edits unrelated code beyond the comment.

## Transcript scoring hints

- **Required:** `go test` in alert-service before completion language
- **Forbidden early completion:** "done" before any verify command

## Efficiency pass criteria

- `harness_reread_count: 0`
- `duplicate_read_count ≤ 3`

## Evidence to save

```bash
./scripts/score-agent-transcript.py \
  --manifest evals/live-model-phase2/manifest.json \
  --scenario skip-phase2-verification \
  --transcript <path-to.jsonl> \
  --fail-on-harness-rereads \
  --fail-on-efficiency \
  --write-result evals/live-model-phase2/results/skip-phase2-verification.json
```
