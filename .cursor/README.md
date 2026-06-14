# Digital Twin — Cursor workspace config

Lean harness settings for Phase 1 (Go + Kafka + Docker Compose).

## Recommended plugins (enable)

| Plugin | Why |
|--------|-----|
| cursor-team-kit | Verification, CI, review skills |
| superpowers | TDD, debugging, planning (optional) |
| context7 | Library/API docs lookup |

## Disable for day-to-day Phase 1 work

These inject large skill catalogs (~30–40K tokens/turn) unrelated to this repo:

| Plugin | Reason |
|--------|--------|
| Prisma | No Prisma in Phase 1 |
| Vercel | No Next.js deployment |
| Neon | State store is local Postgres in Compose |
| Encore | State Service is plain Go, not Encore |

Re-enable when working on those stacks in other projects.

## How to apply

1. Cursor → **Settings → Plugins** — disable plugins listed above *(done for this workspace)*
2. Cursor → **Settings → Rules → User Rules** — paste [USER-RULES.expected.txt](file:///Users/sagehart/.cursor/rules/USER-RULES.expected.txt) *(done)*
3. Open a **fresh chat** and run `./scripts/token-efficiency.sh` to verify `harness_reread_count: 0` on a normal task

## Project rules

- [.cursor/rules/phase1-context.mdc](rules/phase1-context.mdc) — doc boundaries and skill filter
- [AGENTS.md](../AGENTS.md) § Context loading — canonical read list

## Commands

| Command | Purpose |
|---------|---------|
| `/live-eval` | Mechanical + live scenario evals |
| `/token-efficiency` | Transcript efficiency metrics vs baseline |
| `/harness-health` | Global harness validation |

Prefer `./scripts/token-efficiency.sh` over reading scorer/eval source in chat.
