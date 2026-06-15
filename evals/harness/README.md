# Eval harness (scoring secrets)

This directory holds **gate definitions and advisory rubrics** used by `scripts/score-agent-transcript.py`.

Agents running live behavioral evals must **not** read:

- `evals/harness/`
- `evals/fixtures/`
- `scripts/score-agent-transcript.py`
- `scripts/calibrate-harness.py`

Pass/fail is decided from git diff + invariant checks + command ordering, not from phrases in this folder.
