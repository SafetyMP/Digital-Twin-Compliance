#!/usr/bin/env bash
# Backward-compatible alias for run-eval-fixtures.sh
exec "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/scripts/run-eval-fixtures.sh" "$@"
