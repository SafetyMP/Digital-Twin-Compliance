# Digital Twin — Cursor workspace config

Lean harness settings for Phase 2 (Go + Kafka + Flink + alert-console + Docker Compose).

## Session types

See [AGENTS.md](../AGENTS.md) § Session hygiene:

| Task | Fresh chat? | Start with |
|------|-------------|------------|
| Verification | Yes (after analysis/impl) | Smoke scripts via `/verify-phase2` |
| Implementation | Often | Phase spec + service `AGENTS.md` |
| Eval / metrics | Yes | `./scripts/token-efficiency.sh` only |
| Analysis | Optional | User-directed; avoid transcript reads |

Verification chats should pass `./scripts/token-efficiency.sh --strict` (`harness_reread_count: 0`, `duplicate_read_count ≤ 3`).

## Recommended plugins (enable)

| Plugin | Why |
|--------|-----|
| cursor-team-kit | Verification, CI, review skills |
| superpowers | TDD, debugging, planning (optional) |
| context7 | Library/API docs lookup |

## Disable for day-to-day Phase 2 work

These inject large skill catalogs (~30–40K tokens/turn) unrelated to this repo:

| Plugin | Reason |
|--------|--------|
| Prisma | No Prisma in Phase 2 |
| Vercel | No production Next.js deployment |
| Neon | State/alert DBs are local Postgres in Compose |
| Encore | Services are plain Go, not Encore |

Re-enable when working on those stacks in other projects.

## How to apply

1. Cursor → **Settings → Plugins** — disable plugins listed above *(done for this workspace)*
2. Cursor → **Settings → Rules → User Rules** — paste [USER-RULES.expected.txt](file:///Users/sagehart/.cursor/rules/USER-RULES.expected.txt) *(done)*
3. Open a **fresh chat** and run `./scripts/token-efficiency.sh --strict` on a verification task to confirm `harness_reread_count: 0` and `duplicate_read_count ≤ 3`

## Project rules

- [.cursor/rules/phase2-context.mdc](rules/phase2-context.mdc) — doc boundaries and skill filter
- [.cursor/rules/worktrees-repo.mdc](rules/worktrees-repo.mdc) — worktree / parallel constraints (load when using worktrees)
- [AGENTS.md](../AGENTS.md) § Context loading — canonical read list

## Commands

| Command | Purpose |
|---------|---------|
| `/dependency-waves` | Plan and gate ordered waves before spawning child worktrees |
| `/parallel-parent` | Compare, merge, smoke, cleanup after parallel child agents |
| `/worktree-agent` | Create/list/remove git worktrees for isolated parallel tracks |

Docs: [docs/quickstart-agent-worktrees.md](../docs/quickstart-agent-worktrees.md)
| `/best-of-n-worktrees` | Create N worktrees for parallel solution attempts + compare |
| `/verify-phase2` | Phase 2 smoke verification (no transcript archaeology) |
| `/live-eval` | Phase 1 mechanical + live scenario evals |
| `/live-eval-phase2` | Phase 2 mechanical + live scenario evals |
| `/token-efficiency` | Transcript efficiency metrics vs baseline |
| `/harness-health` | Global harness validation |

**Scorecard:** `./scripts/report-eval-scorecard.sh --all [--full]` — Product + Behavior + Efficiency pillars.

**Baseline:** `./scripts/refresh-efficiency-baseline.sh` after several clean verification sessions.

Prefer `./scripts/token-efficiency.sh` over reading scorer/eval source in chat. Use `--strict` to fail on investigate thresholds.
