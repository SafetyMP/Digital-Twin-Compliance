---
description: Report token efficiency metrics for an agent transcript vs baseline thresholds.
---
Report token efficiency for a Digital Twin agent session.

**Do not Read** `scripts/score-agent-transcript.py`, `evals/live-model/README.md`, or `manifest.json`. Run the wrapper script only:

```bash
./scripts/token-efficiency.sh
# or with explicit path:
./scripts/token-efficiency.sh /path/to/chat.jsonl
# compare to saved baseline:
./scripts/token-efficiency.sh --compare-baseline
# fail (exit 1) on investigate thresholds:
./scripts/token-efficiency.sh --strict
```

If the user provided a transcript path, pass it as the first argument.

Compare output against thresholds (healthy / investigate):

| Signal | Healthy | Investigate |
|--------|---------|-------------|
| duplicate_read_count | 0–3 | >3 (`--strict` exits 1) |
| harness_reread_count | 0 | any |
| tool_call_count (scope-refusal scenario) | <20 | >40 |
| transcript_bytes (implementation) | <500KB | >750KB |

The script prints an **Investigate summary** with hints when thresholds are exceeded.

Baseline file: `evals/live-model/results/efficiency-baseline.json`.

Remind: fresh chats for verification and live evals; AGENTS.md § Session hygiene; [.cursor/README.md](../README.md) plugin allowlist.
