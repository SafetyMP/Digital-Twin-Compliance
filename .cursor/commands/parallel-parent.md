---
description: Orchestrate parallel child agents — compare, merge, verify-worktree-merge, smoke from main root, cleanup worktrees.
---
You are the **parent agent** on the main repo root per [docs/handoff-parallel-parent.md](../../docs/handoff-parallel-parent.md).

Load **`22-parallel-agents`** (`~/.cursor/rules/22-parallel-agents.mdc`) and repo **`worktrees-repo`** rule when worktrees are involved.

**Do not** run integration scripts from a worktree path. **Do not** trust child smoke/Compose claims.

## Gather state

```bash
./scripts/agent-worktree-best-of-n.sh status --prefix <id>   # if best-of-N
./scripts/agent-worktree.sh list
./scripts/check-agent-worktrees.sh
```

## Compare and merge

```bash
./scripts/agent-worktree-best-of-n.sh compare --prefix <id>
git checkout <integration-branch> && git merge agent/<track>/<name>
```

## Post-merge verify (main root — required)

```bash
./scripts/check-worktree-scope.sh --branch agent/<track>/<name> --strict
./scripts/verify-worktree-merge.sh agent/<track>/<name>
./scripts/verify-worktree-merge.sh agent/<track>/<name> --rebuild
./scripts/verify-worktree-merge.sh agent/<track>/<name> --rebuild --with-smoke
```

`--with-smoke` requires `--rebuild` when backend/compose paths changed. Do not run smoke from a `.worktrees/` cwd.

## Cleanup

```bash
./scripts/agent-worktree-best-of-n.sh remove --prefix <id>
./scripts/check-agent-worktrees.sh
```

Return: winner chosen, scope check, verify-worktree-merge output, rebuild commands run, smoke exit codes, cleanup status.
