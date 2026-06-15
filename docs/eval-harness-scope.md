# Eval harness scope

The Digital Twin eval stack measures **agent discipline**, not **code correctness**. Those are separate instruments.

## What the harness detects

| Pillar | Instrument | Detects |
|--------|------------|---------|
| Product | `run-live-evals*.sh` | Repo invariants: tenant columns, outbox-only writers, Phase 3 boundary, deliverables |
| Behavior | `score-agent-transcript.py` + git diff | Process compliance: verify before claiming done, refuse scope creep, preserve contracts |
| Efficiency | `token-efficiency.sh` | Context hygiene: duplicate reads, harness re-reads |

Behavioral **pass/fail** is gated by:

1. **Invariant checks** on `git diff` (or transcript edits when no diff is supplied)
2. **Command ordering** (tests/smoke/Flink checks before completion claims)
3. **Integration checks** (no false pipeline-health claims without verification)

Prose pushback and optional LLM advisory grades are **non-blocking**.

## What the harness cannot detect

Examples of defects that pass a green behavioral harness:

- Subtle logic bugs (e.g. consumer commits on error while appearing well-tested)
- Race conditions in streaming pipelines
- Incorrect idempotency under partial failure
- Schema mapping errors that still compile

Those require the **correctness track**: unit/integration/property tests and domain review.

## Correctness track (required alongside harness)

- Go package tests under `services/*/internal/**`
- Compose smoke tests (`smoke-test.sh`, `smoke-test-phase2.sh`)
- Static boundary tests (see `services/state-service/internal/outbox/publish_boundary_test.go`)

Do not treat a green behavior pillar as proof the system is correct.

## Contamination

Gate definitions live in `evals/harness/` and are listed in `.cursorignore`. Agents under eval must not read scoring logic or fixture answer keys during live sessions.

## Calibration

Run `./scripts/calibrate-harness.py --fail-on-regression` after scorer changes. Labeled samples: `evals/fixtures/labeled/manifest.json`.

## N-sampling

Store multiple runs per scenario under `evals/live-model-phase2/results/<scenario>/run-*.json`. Report pass rates with `./scripts/report-eval-scorecard.sh --phase2`.
