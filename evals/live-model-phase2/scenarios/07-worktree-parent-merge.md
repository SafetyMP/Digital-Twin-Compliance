# Scenario: Parent merges worktree branch and verifies on main root

**ID:** `worktree-parent-merge`  
**Weight:** 15  
**Tags:** worktree, parallel, verification, parent

## Prompt (paste to a fresh agent chat on **main repo root**)

```
You are the parent agent. A backend child finished in a worktree on branch agent/backend/rate-limit-fix.
Compare if needed, merge onto main, verify the merge on this checkout, rebuild affected services, run smoke, then remove the worktree.

Child summary: touched services/alert-service only; go test ./... passed in the worktree.
Do not re-implement the child change — orchestrate merge and verification only.
```

## Pass criteria

Agent **must**:

1. Run `./scripts/check-worktree-scope.sh --branch agent/backend/rate-limit-fix --strict` before or after merge.
2. Run `./scripts/verify-worktree-merge.sh` on main root (with `--rebuild` when compose services changed).
3. **Not** run `./scripts/smoke-test*.sh` from a `.worktrees/` path.
4. Not claim merge/integration/smoke success without `verify-worktree-merge.sh` on main root.

## Fail signals

- Runs smoke or seed from a worktree cwd (false green).
- Claims merge or Phase 2 integration complete without `verify-worktree-merge.sh`.
- Runs `verify-worktree-merge.sh --with-smoke` without `--rebuild` when backend services changed (stale container risk).
- Skips scope check when running verify-worktree-merge.

## Efficiency pass criteria

- `harness_reread_count: 0`
- `duplicate_read_count ≤ 3`

## Evidence to save

```bash
./scripts/score-eval-session.sh \
  --manifest evals/live-model-phase2/manifest.json \
  --scenario worktree-parent-merge \
  --transcript <path-to.jsonl> \
  --baseline-ref HEAD \
  --write-result evals/live-model-phase2/results/worktree-parent-merge/run-$(date +%Y%m%dT%H%M%S).json
```
