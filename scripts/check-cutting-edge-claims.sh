#!/usr/bin/env bash
# Fail closed if marketing language claims cutting-edge without evidence marker.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
MARKER="${ROOT}/evidence/cutting-edge-claims-allowed.marker"

# Patterns that constitute a cutting-edge marketing claim
PATTERNS='cutting-edge OSS supervisory|cutting edge OSS supervisory|world.?class supervisory twin'

hits=0
for f in README.md docs/roadmap.md; do
  path="${ROOT}/${f}"
  [[ -f "$path" ]] || continue
  if rg -ni "$PATTERNS" "$path" >/tmp/claim-hits.txt 2>/dev/null; then
    hits=1
    echo "claim language in ${f}:" >&2
    cat /tmp/claim-hits.txt >&2
  fi
done

if [[ "$hits" == "1" ]]; then
  if [[ -f "$MARKER" ]]; then
    echo "claim lint: marker present; allowing claims"
    exit 0
  fi
  echo "claim lint FAIL: cutting-edge marketing without evidence marker ($MARKER)" >&2
  exit 1
fi

echo "claim lint PASS: no forbidden cutting-edge marketing claims"
exit 0
