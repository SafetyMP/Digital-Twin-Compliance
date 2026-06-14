---
description: Run Phase 2 live-model evals (mechanical checks and optional transcript scoring).
---
Run Digital Twin Phase 2 effectiveness evals.

**Do not Read** eval harness source unless debugging the harness itself. Run scripts directly:

1. **Mechanical (always):** `./scripts/run-live-evals-phase2.sh`
2. **Full DoD (stack up):** `./scripts/run-live-evals-phase2.sh --full`
3. **Live scenario scoring:** `./scripts/score-agent-transcript.py --manifest evals/live-model-phase2/manifest.json --scenario <id> --transcript <path>`
   - List ids: `./scripts/score-agent-transcript.py --manifest evals/live-model-phase2/manifest.json --list-scenarios`
   - Fail on harness rereads (eval sessions): add `--fail-on-harness-rereads`

Scenario prompts: `evals/live-model-phase2/scenarios/` — **fresh chat** per scenario ([AGENTS.md](../../AGENTS.md) § Session hygiene).

Phase 1 evals: `/live-eval` or `./scripts/run-live-evals.sh`.

Efficiency: `./scripts/token-efficiency.sh` (not `/token-efficiency` via Read loop).

**Three-pillar report:** `./scripts/report-eval-scorecard.sh --phase2 [--full] [--transcript path.jsonl]`

Baseline refresh: `./scripts/refresh-efficiency-baseline.sh` (after clean verification sessions).

Report: mechanical result, optional DoD, pass bar (100% mechanical, ≥4/5 live scenarios).
