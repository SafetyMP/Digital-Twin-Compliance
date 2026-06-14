#!/usr/bin/env bash
#
# Token efficiency report for the latest substantive agent transcript.
# Avoids agents reading scorer/eval source — run this script directly.
#
# Usage:
#   ./scripts/token-efficiency.sh                    # newest parent chat by mtime
#   ./scripts/token-efficiency.sh --substantive        # newest >= 1KB (legacy)
#   ./scripts/token-efficiency.sh /path/to/chat.jsonl
#   ./scripts/token-efficiency.sh --compare-baseline

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

TRANSCRIPT_DIR="${TRANSCRIPT_DIR:-$HOME/.cursor/projects/Users-sagehart-Downloads-Digital-Twin/agent-transcripts}"
BASELINE="$ROOT/evals/live-model/results/efficiency-baseline.json"
MIN_BYTES=100
SUBSTANTIVE=0
WARN_BYTES=102400
COMPARE=0
TRANSCRIPT=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --compare-baseline)
      COMPARE=1
      shift
      ;;
    --substantive)
      SUBSTANTIVE=1
      MIN_BYTES=1024
      shift
      ;;
    --transcript-dir)
      TRANSCRIPT_DIR="$2"
      shift 2
      ;;
    -*)
      echo "Unknown option: $1" >&2
      exit 2
      ;;
    *)
      TRANSCRIPT="$1"
      shift
      ;;
  esac
done

if [[ -z "$TRANSCRIPT" ]]; then
  TRANSCRIPT="$(python3 - "$TRANSCRIPT_DIR" "$MIN_BYTES" <<'PY'
import sys
from pathlib import Path

tdir = Path(sys.argv[1])
min_bytes = int(sys.argv[2])
if not tdir.is_dir():
    sys.exit(1)
candidates = [
    p for p in tdir.glob("*/*.jsonl")
    if "subagents" not in p.parts and p.stat().st_size >= min_bytes
]
if not candidates:
    sys.exit(2)
print(max(candidates, key=lambda p: p.stat().st_mtime))
PY
)" || {
    echo "No transcript (>= ${MIN_BYTES} bytes) under: $TRANSCRIPT_DIR" >&2
    exit 2
  }
  MODE="newest"
  [[ "$SUBSTANTIVE" -eq 1 ]] && MODE="newest substantive (>=1KB)"
  echo "Selected ($MODE): $TRANSCRIPT"
  SIZE="$(wc -c < "$TRANSCRIPT" | tr -d ' ')"
  if [[ "$SIZE" -gt "$WARN_BYTES" && "$SUBSTANTIVE" -eq 0 ]]; then
    echo "WARNING: Large session (${SIZE} bytes) — likely a long implementation thread."
    echo "         Post-trim verification: score a fresh chat explicitly, e.g.:"
    ALT="$(python3 - "$TRANSCRIPT_DIR" <<'PY'
import sys
from pathlib import Path
tdir = Path(sys.argv[1])
cands = sorted(
    [p for p in tdir.glob("*/*.jsonl") if "subagents" not in p.parts and p.stat().st_size >= 100],
    key=lambda p: p.stat().st_mtime,
    reverse=True,
)
for p in cands[1:4]:
    if p.stat().st_size < 102400:
        print(p)
        break
PY
)"
    if [[ -n "$ALT" ]]; then
      echo "         ./scripts/token-efficiency.sh $ALT"
    fi
  fi
  echo "---"
fi

if [[ ! -f "$TRANSCRIPT" ]]; then
  echo "Transcript not found: $TRANSCRIPT" >&2
  exit 2
fi

./scripts/score-agent-transcript.py --metrics-only --transcript "$TRANSCRIPT"

if [[ "$COMPARE" -eq 1 && -f "$BASELINE" ]]; then
  echo "---"
  echo "Baseline comparison:"
  python3 - "$BASELINE" "$TRANSCRIPT" "$ROOT" <<'PY'
import json
import subprocess
import sys
from pathlib import Path

baseline_path = Path(sys.argv[1])
transcript_path = Path(sys.argv[2])
root = Path(sys.argv[3])
chat_id = transcript_path.parent.name

baseline = json.loads(baseline_path.read_text())
entry = next((s for s in baseline.get("sessions", []) if s.get("chat_id") == chat_id), None)
if not entry:
    print(f"  No baseline entry for chat_id={chat_id}")
    sys.exit(0)

out = subprocess.check_output(
    [str(root / "scripts" / "score-agent-transcript.py"), "--metrics-json", "--transcript", str(transcript_path)],
    text=True,
)
cur = json.loads(out)

print(f"  chat_id: {chat_id}")
for key in (
    "transcript_bytes",
    "tool_call_count",
    "duplicate_read_count",
    "harness_reread_count",
    "estimated_tokens",
):
    b, c = int(entry.get(key, 0)), int(cur.get(key, 0))
    delta = c - b
    suffix = f" ({delta / b * 100:+.0f}%)" if b else ""
    print(f"  {key}: baseline={b:,}  current={c:,}  delta={delta:+,}{suffix}")
PY
fi
