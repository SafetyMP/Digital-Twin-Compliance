#!/usr/bin/env bash
# OSS publish metadata + agent-harness audit (CI-local mirror of harness publish-doctor).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

fail=0
pass=0
ok() { echo "  ok   $1: $2"; pass=$((pass + 1)); }
bad() { echo "  FAIL $1: $2"; fail=$((fail + 1)); }

echo "== publish-check ($ROOT) =="

# git remote origin → GitHub
if origin="$(git remote get-url origin 2>/dev/null || true)" && [[ -n "$origin" ]]; then
  if [[ "$origin" =~ github\.com[:/]([^/]+/[^/.]+) ]]; then
    slug="$(echo "${BASH_REMATCH[1]}" | tr '[:upper:]' '[:lower:]')"
    ok "git-remote-origin" "origin → $slug"
  else
    bad "git-remote-origin" "origin URL not recognized as GitHub: $origin"
  fi
else
  bad "git-remote-origin" "No git remote 'origin'"
fi

# package.json repository (optional)
if [[ -f package.json ]]; then
  slug="$(python3 - <<'PY' 2>/dev/null || true
import json, re, sys
from pathlib import Path
pkg = json.loads(Path("package.json").read_text())
repo = pkg.get("repository")
url = repo if isinstance(repo, str) else (repo or {}).get("url") if isinstance(repo, dict) else None
if not url:
    sys.exit(0)
for pat in (r"^git@github\.com:([^/]+/[^/]+)", r"^https?://github\.com/([^/]+/[^/]+)"):
    m = re.match(pat, url.strip().removesuffix(".git"), re.I)
    if m:
        print(m.group(1).lower())
        break
PY
)"
  if [[ -z "$slug" ]]; then
    bad "package-repository" "package.json missing or has no parseable repository.url"
  else
    ok "package-repository" "package.json repository → $slug"
  fi
else
  ok "package-repository" "No package.json — skipped repository.url check"
fi

# CI workflows
wf="$ROOT/.github/workflows"
if [[ ! -d "$wf" ]]; then
  bad "ci-workflows" "No .github/workflows/"
else
  count="$(find "$wf" -maxdepth 1 \( -name '*.yml' -o -name '*.yaml' \) | wc -l | tr -d ' ')"
  if [[ "$count" -gt 0 ]]; then
    ok "ci-workflows" "$count workflow file(s) in .github/workflows/"
  else
    bad "ci-workflows" ".github/workflows/ exists but has no .yml/.yaml files"
  fi
fi

# OSS files
missing=()
for f in LICENSE README.md SECURITY.md; do
  [[ -f "$ROOT/$f" ]] || missing+=("$f")
done
if ((${#missing[@]})); then
  bad "oss-files" "Missing OSS files: ${missing[*]}"
else
  ok "oss-files" "OSS files present: LICENSE, README.md, SECURITY.md"
fi

# .opencode/
if [[ -d "$ROOT/.opencode" ]]; then
  bad "agent-harness-opencode" ".opencode/ present — keep OpenCode harness local unless intentional"
else
  ok "agent-harness-opencode" "No .opencode/ — agent harness not in tree (OK for public app repos)"
fi

# .cursor/
cursor="$ROOT/.cursor"
if [[ ! -d "$cursor" ]]; then
  ok "agent-harness-cursor" "No .cursor/ — OK (or add curated .cursor/rules/ for contributors if desired)"
else
  has_rules=0
  [[ -d "$cursor/rules" ]] && has_rules=1
  sensitive=()
  for f in hooks.json mcp.json; do
    git ls-files --error-unmatch ".cursor/$f" &>/dev/null && sensitive+=("$f")
  done
  has_agents=0
  git ls-files --error-unmatch .cursor/agents &>/dev/null && has_agents=1

  if [[ "$has_rules" -eq 1 && ${#sensitive[@]} -eq 0 && "$has_agents" -eq 0 ]]; then
    ok "agent-harness-cursor" ".cursor/rules/ present; no mcp.json, hooks.json, or agents/"
  else
    msg=".cursor/ present"
    [[ "$has_rules" -eq 0 ]] && msg="$msg; missing rules/"
    ((${#sensitive[@]})) && msg="$msg; sensitive: ${sensitive[*]}"
    [[ "$has_agents" -eq 1 ]] && msg="$msg; agents/ dir present"
    msg="$msg — ship .cursor/rules/ only; gitignore mcp.json, hooks.json, agents/"
    bad "agent-harness-cursor" "$msg"
  fi
fi

# No absolute home paths in committed eval JSON (machine-specific session artifacts)
path_hits=()
while IFS= read -r f; do
  if grep -qE '"/Users/|"/home/' "$f" 2>/dev/null; then
    path_hits+=("$f")
  fi
done < <(git ls-files 'evals/**/*.json' 2>/dev/null || true)

if ((${#path_hits[@]})); then
  bad "eval-path-redaction" "Absolute home paths in: ${path_hits[*]}"
else
  ok "eval-path-redaction" "No absolute /Users or /home paths in tracked evals/**/*.json"
fi

echo "  -> $pass passed, $fail failed"
if [[ "$fail" -gt 0 ]]; then
  echo "  -> fix before claiming OSS-ready"
  exit 1
fi
echo "  -> publish metadata OK"
exit 0
