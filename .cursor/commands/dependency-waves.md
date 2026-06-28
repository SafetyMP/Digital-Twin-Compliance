---
description: Plan and gate parallel agent dependency waves before spawning child worktrees.
---
Orchestrate **ordered** parallel work per [docs/handoff-dependency-waves.md](../../docs/handoff-dependency-waves.md).

Load **`agent-worktrees`** skill → `references/dependency-waves.md`.

## Plan

```bash
./scripts/check-dependency-waves.sh validate
./scripts/check-dependency-waves.sh plan
./scripts/check-dependency-waves.sh init --task <task-id>
./scripts/check-dependency-waves.sh status --task <task-id>
```

## Spawn children (only when ready)

```bash
./scripts/check-dependency-waves.sh ready --task <task-id> --wave <wave-id>
./scripts/check-dependency-waves.sh handoff --task <task-id> --wave <wave-id>
```

## After merge

```bash
./scripts/check-dependency-waves.sh complete --task <task-id> --wave <wave-id> --branch agent/<track>/<name>
```

Integration wave (`runner: parent`): use `/parallel-parent` — never delegate smoke to children.

Return: current wave, ready/blocked status, next handoff or integration steps.
