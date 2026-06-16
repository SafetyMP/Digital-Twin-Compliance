#!/usr/bin/env bash
#
# Regression tests for efficiency and behavioral scenario scoring fixtures.
#
# Usage: ./scripts/run-eval-fixtures.sh

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

FIXTURES="$ROOT/evals/fixtures/transcripts"
DIFFS="$ROOT/evals/fixtures/diffs"
PASS=0
FAIL=0

pass() { echo "  ok  - $1"; PASS=$((PASS + 1)); }
fail() { echo "  FAIL - $1"; FAIL=$((FAIL + 1)); }

score_expect() {
  local label="$1"
  local expect_exit="$2"
  shift 2
  if "$@" >/dev/null 2>&1; then
    rc=0
  else
    rc=$?
  fi
  if [[ "$rc" -eq "$expect_exit" ]]; then
    pass "$label"
  else
    fail "$label (expected exit $expect_exit, got $rc)"
  fi
}

echo "== Efficiency fixture regression =="

score_expect "efficiency-pass.jsonl passes --strict" 0 \
  ./scripts/token-efficiency.sh --strict "$FIXTURES/efficiency-pass.jsonl"

score_expect "efficiency-dup-read-fail.jsonl fails --strict" 1 \
  ./scripts/token-efficiency.sh --strict "$FIXTURES/efficiency-dup-read-fail.jsonl"

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

score_expect "harness read fixture fails --strict" 1 \
  ./scripts/token-efficiency.sh --strict "$HARNESS_FIXTURE"

run_one_fixture() {
  local manifest="$1"
  local scenario="$2"
  local pass_file="$3"
  local fail_file="$4"
  local diff_file="${5:-}"

  score_expect "$scenario pass ($pass_file)" 0 \
    ./scripts/score-agent-transcript.py \
      --manifest "$manifest" \
      --scenario "$scenario" \
      --transcript "$FIXTURES/$pass_file" \
      --fail-on-harness-rereads \
      --fail-on-efficiency

  local fail_cmd=(
    ./scripts/score-agent-transcript.py
    --manifest "$manifest"
    --scenario "$scenario"
    --transcript "$FIXTURES/$fail_file"
    --fail-on-harness-rereads
    --fail-on-efficiency
  )
  if [[ -n "$diff_file" && -f "$DIFFS/$diff_file" ]]; then
    fail_cmd+=(--diff "$DIFFS/$diff_file")
  fi

  score_expect "$scenario fail ($fail_file)" 1 "${fail_cmd[@]}"
}

MANIFEST_P1="$ROOT/evals/live-model/manifest.json"
MANIFEST_P2="$ROOT/evals/live-model-phase2/manifest.json"

echo
echo "== Scenario fixtures (phase1) =="
run_one_fixture "$MANIFEST_P1" claim-phase1-complete \
  scenario-claim-phase1-complete-pass.jsonl scenario-claim-phase1-complete-fail.jsonl
run_one_fixture "$MANIFEST_P1" add-flink-job \
  scenario-add-flink-job-pass.jsonl scenario-add-flink-job-fail.jsonl add-flink-job-fail.patch
run_one_fixture "$MANIFEST_P1" skip-verification \
  scenario-skip-verification-pass.jsonl scenario-skip-verification-fail.jsonl
run_one_fixture "$MANIFEST_P1" bypass-outbox \
  scenario-bypass-outbox-pass.jsonl scenario-bypass-outbox-fail.jsonl bypass-outbox-fail.patch
run_one_fixture "$MANIFEST_P1" break-tenant-contract \
  scenario-break-tenant-contract-pass.jsonl scenario-break-tenant-contract-fail.jsonl break-tenant-contract-fail.patch

echo
echo "== Scenario fixtures (phase2) =="
run_one_fixture "$MANIFEST_P2" claim-phase2-complete \
  scenario-claim-phase2-complete-pass.jsonl scenario-claim-phase2-complete-fail.jsonl
run_one_fixture "$MANIFEST_P2" add-cedar-policy \
  scenario-add-cedar-policy-pass.jsonl scenario-add-cedar-policy-fail.jsonl add-cedar-policy-fail.patch
run_one_fixture "$MANIFEST_P2" skip-phase2-verification \
  scenario-skip-phase2-verification-pass.jsonl scenario-skip-phase2-verification-fail.jsonl
run_one_fixture "$MANIFEST_P2" skip-flink-running-check \
  scenario-skip-flink-running-check-pass.jsonl scenario-skip-flink-running-check-fail.jsonl
run_one_fixture "$MANIFEST_P2" add-immudb-audit \
  scenario-add-immudb-audit-pass.jsonl scenario-add-immudb-audit-fail.jsonl add-immudb-audit-fail.patch
run_one_fixture "$MANIFEST_P2" debug-int-m001-retention \
  scenario-debug-int-m001-retention-pass.jsonl scenario-debug-int-m001-retention-fail.jsonl

echo
echo "RESULT: $PASS passed, $FAIL failed"
[[ "$FAIL" -eq 0 ]]
