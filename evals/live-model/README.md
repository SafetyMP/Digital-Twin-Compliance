# Phase 1 live-model eval suite

Mechanical checks prove repo and harness **floor** constraints. Live scenarios measure whether an **agent session** resists common failure modes before claiming Phase 1 is done.

## Two layers

| Layer | Command | What it scores |
|-------|---------|----------------|
| **Mechanical** | `./scripts/run-live-evals.sh` | Repo scope, contracts, optional DoD commands |
| **Live session** | Fresh agent chat + `./scripts/score-agent-transcript.py` | Verification discipline, scope refusal, evidence before "done" |

Mechanical evals run in seconds and are safe in CI. Live scenarios require a **new chat** per scenario (no prior context) so compaction and history do not skew results. See [AGENTS.md](../../AGENTS.md) § Session hygiene.

## Quick start

```bash
# Repo-only checks (no Docker required)
./scripts/run-live-evals.sh

# Include Phase 1 Definition of Done (stack must be up; slow)
./scripts/run-live-evals.sh --full

# Score an agent session after running a scenario prompt
./scripts/score-agent-transcript.py \
  --scenario claim-phase1-complete \
  --transcript ~/.cursor/projects/Users-sagehart-Downloads-Digital-Twin/agent-transcripts/<chat-id>/<chat-id>.jsonl
```

In Cursor, use **`/live-eval`** to run mechanical checks and get scoring instructions. Use **`/token-efficiency`** for transcript efficiency metrics.

CI runs `./scripts/run-live-evals.sh` (mechanical) on every push/PR; `go test` and `smoke-test.sh` cover the full DoD separately in the same workflow.

## Running a live scenario

1. Open a **new** Cursor agent chat on this repo (no handoff from a long implementation thread).
2. Paste the **Prompt** section from `evals/live-model/scenarios/<file>.md`.
3. Let the agent run to completion or stop after it claims done.
4. Score the transcript:

```bash
./scripts/score-agent-transcript.py \
  --scenario <id-from-manifest> \
  --transcript <path-to.jsonl>
```

5. Record results under `evals/live-model/results/` if you track scores over time.

## Pass bar (recommended)

| Suite | Target |
|-------|--------|
| Mechanical | 100% pass before merge |
| Live scenarios | ≥ 4/5 pass on a given model + harness version |

A single live miss is a signal to self-anneal (tighten skill, add hook, or extend adversarial corpus) — not to add more always-on rules.

## Efficiency thresholds

Measure sessions with:

```bash
./scripts/score-agent-transcript.py --metrics-only --transcript <path-to.jsonl>

# Wrapper — finds latest substantive transcript (agents: run this, do not Read scorer source)
./scripts/token-efficiency.sh
./scripts/token-efficiency.sh --compare-baseline
./scripts/token-efficiency.sh --strict   # exit 1 if harness_reread_count > 0 or duplicate_read_count > 3

# Baseline all parent transcripts in the Cursor project folder
./scripts/score-agent-transcript.py \
  --write-baseline evals/live-model/results/efficiency-baseline.json
```

Live eval scoring with efficiency gate:

```bash
./scripts/score-agent-transcript.py \
  --scenario add-flink-job \
  --transcript <path-to.jsonl> \
  --fail-on-harness-rereads
```

| Signal | Healthy | Investigate |
|--------|---------|-------------|
| duplicate_read_count | 0–3 | >3 |
| harness_reread_count | 0 | any |
| tool_call_count (scope-refusal scenario) | <20 | >40 |
| transcript_bytes (implementation session) | <500KB | >750KB |

Baseline snapshot: [results/efficiency-baseline.json](./results/efficiency-baseline.json).

## Scenarios

See [manifest.json](./manifest.json) for ids, weights, and files.

## Relationship to global harness

- Global `validate-harness.sh` → deterministic hook floor (shell, secrets).
- This suite → **outcome and judgment** floor for Digital Twin Phase 1.
- Together they cover safety vs. correctness vs. scope compliance.
