# Phase 2 live-model eval suite

Mechanical checks prove Phase 2 repo and harness **floor** constraints. Live scenarios measure whether an **agent session** resists common Phase 2 failure modes before claiming done.

## Two layers

| Layer | Command | What it scores |
|-------|---------|----------------|
| **Mechanical** | `./scripts/run-live-evals-phase2.sh` | Phase 2 deliverables, contracts, optional DoD commands |
| **Live session** | Fresh agent chat + `./scripts/score-agent-transcript.py --manifest evals/live-model-phase2/manifest.json` | Verification discipline, Phase 3 scope refusal, Flink integration checks, **contract retention** (`debug-int-m001-retention`) |

Mechanical evals run in seconds and are safe in CI. Live scenarios require a **new chat** per scenario (no prior context). See [AGENTS.md](../../AGENTS.md) § Session hygiene.

## Quick start

```bash
# Repo-only checks (no Docker required)
./scripts/run-live-evals-phase2.sh

# Include Phase 2 Definition of Done (stack must be up; slow)
./scripts/run-live-evals-phase2.sh --full

# Score an agent session after running a scenario prompt
./scripts/score-agent-transcript.py \
  --manifest evals/live-model-phase2/manifest.json \
  --scenario claim-phase2-complete \
  --transcript ~/.cursor/projects/Users-sagehart-Downloads-Digital-Twin/agent-transcripts/<chat-id>/<chat-id>.jsonl
```

In Cursor, use **`/live-eval-phase2`** to run mechanical checks and get scoring instructions.

Phase 1 evals remain at `./scripts/run-live-evals.sh` and `/live-eval`.

## Running a live scenario

1. Open a **new** Cursor agent chat on this repo.
2. Paste the **Prompt** section from `evals/live-model-phase2/scenarios/<file>.md`.
3. Let the agent run to completion or stop after it claims done.
4. Score the transcript (always pass `--manifest evals/live-model-phase2/manifest.json`).
5. Record results under `evals/live-model-phase2/results/<scenario-id>/run-<timestamp>.json` when tracking scores over time.

## Pass bar (recommended)

| Suite | Target |
|-------|--------|
| Mechanical | 100% pass before merge |
| Live scenarios | ≥ 5/6 scenarios at ≥80% pass over 3 runs each (see `report-eval-scorecard.sh --phase2`) |

## Efficiency

Same thresholds as Phase 1 — use `--fail-on-harness-rereads` when scoring live sessions:

```bash
./scripts/score-agent-transcript.py \
  --manifest evals/live-model-phase2/manifest.json \
  --scenario add-cedar-policy \
  --transcript <path-to.jsonl> \
  --fail-on-harness-rereads
```

Wrapper: `./scripts/token-efficiency.sh --strict` (harness_reread_count must be 0).

## Scenarios

See [manifest.json](./manifest.json) for ids, weights, and files.

## Relationship to Phase 1 evals

| Phase | Mechanical script | Scope trap scenarios |
|-------|-------------------|----------------------|
| Phase 1 | `run-live-evals.sh` | Flink in Phase 1, outbox bypass, tenant drop |
| Phase 2 | `run-live-evals-phase2.sh` | Cedar/immudb in Phase 2, Flink health without proof |

Both suites share efficiency metrics and the same transcript scorer (different manifest).
