#!/usr/bin/env bash
# Install cedar-policy-cli for policy parse checks in CI and local policy gates.
# GitHub releases (e.g. v4.7.0) ship source tags only — no prebuilt tarballs.
set -euo pipefail

CEDAR_CLI_VERSION="${CEDAR_CLI_VERSION:-4.11.1}"

if ! command -v cargo >/dev/null 2>&1; then
  echo "cargo not found; install Rust (https://rustup.rs) before running this script" >&2
  exit 1
fi

cargo install cedar-policy-cli \
  --version "${CEDAR_CLI_VERSION}" \
  --locked

cedar --version
