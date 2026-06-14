#!/usr/bin/env bash
#
# Regression tests for efficiency and scenario scoring fixtures.
#
# Usage: ./scripts/run-efficiency-fixtures.sh

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

FIXTURES="$ROOT/evals/fixtures/transcripts"
PASS=0
FAIL=0

pass() { echo "  ok  - $1"; PASS=$((PASS + 1)); }
fail() { echo "  FAIL - $1"; FAIL=$((FAIL + 1)); }

echo "== Efficiency fixture regression =="

if ./scripts/token-efficiency.sh --strict "$FIXTURES/efficiency-pass.jsonl" >/dev/null; then
  pass "efficiency-pass.jsonl passes --strict"
else
  fail "efficiency-pass.jsonl should pass --strict"
fi

if ./scripts/token-efficiency.sh --strict "$FIXTURES/efficiency-dup-read-fail.jsonl" >/dev/null 2>&1; then
  fail "efficiency-dup-read-fail.jsonl should fail --strict"
else
  pass "efficiency-dup-read-fail.jsonl fails --strict (duplicate reads)"
fi

HARNESS_FIXTURE="$(mktemp)"
trap 'rm -f "$HARNESS_FIXTURE"' EXIT
python3 - "$HARNESS_FIXTURE" <<'PY'
import json
import os
import sys

path = os.path.join(os.environ["HOME"], ".cursor", "HARNESS.md")
row = {
    "role": "assistant",
    "message": {
        "content": [
            {"type": "tool_use", "name": "Read", "input": {"path": path}},
            {"type": "text", "text": "Read harness file."},
        ]
    },
}
with open(sys.argv[1], "w", encoding="utf-8") as f:
    f.write(json.dumps(row) + "\n")
PY

if ./scripts/token-efficiency.sh --strict "$HARNESS_FIXTURE" >/dev/null 2>&1; then
  fail "dynamic harness fixture should fail --strict"
else
  pass "harness read fixture fails --strict"
fi

echo
echo "== Scenario fixture regression =="

if ./scripts/score-agent-transcript.py \
  --scenario skip-verification \
  --transcript "$FIXTURES/scenario-skip-verification-pass.jsonl" \
  --fail-on-harness-rereads \
  --fail-on-efficiency >/dev/null; then
  pass "scenario-skip-verification-pass.jsonl"
else
  fail "scenario-skip-verification-pass.jsonl"
fi

if ./scripts/score-agent-transcript.py \
  --scenario bypass-outbox \
  --transcript "$FIXTURES/scenario-bypass-outbox-pass.jsonl" \
  --fail-on-harness-rereads \
  --fail-on-efficiency >/dev/null; then
  pass "scenario-bypass-outbox-pass.jsonl"
else
  fail "scenario-bypass-outbox-pass.jsonl"
fi

echo
echo "RESULT: $PASS passed, $FAIL failed"
[[ "$FAIL" -eq 0 ]]
