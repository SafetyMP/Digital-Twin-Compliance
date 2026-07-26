#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
chmod +x "$ROOT/scripts/phase7/reg2policy.sh"
"$ROOT/scripts/phase7/reg2policy.sh"
echo "Phase 7 reg2policy smoke passed"
