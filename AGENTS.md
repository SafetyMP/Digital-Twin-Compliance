# AGENTS.md

Operational contract for coding agents working in this repository.

## Current phase

**Phase 2** — Real-time compliance monitoring and alert delivery. Full spec: [docs/phase2-implementation-spec.md](docs/phase2-implementation-spec.md). Phase 1 spec: [docs/phase1-implementation-spec.md](docs/phase1-implementation-spec.md).

Architecture and domain docs live under [docs/](docs/). Do not implement Phase 2+ components unless explicitly requested.

## Context loading

Minimize token use — load only what the task requires:

- **Always load**: this file, [docs/phase2-implementation-spec.md](docs/phase2-implementation-spec.md) (Phase 2) or [docs/phase1-implementation-spec.md](docs/phase1-implementation-spec.md) when Phase 1-only work
- **For Go State Service work**: [services/state-service/AGENTS.md](services/state-service/AGENTS.md)
- **For Go Alert Service work**: [services/alert-service/AGENTS.md](services/alert-service/AGENTS.md)
- **For envelope / idempotency / outbox tasks only**: [docs/data-flow.md](docs/data-flow.md)
- **Do not load unless the task explicitly requires**: [docs/architecture.md](docs/architecture.md), [docs/domain-model.md](docs/domain-model.md), [docs/compliance-mapping.md](docs/compliance-mapping.md), [docs/roadmap.md](docs/roadmap.md), ADRs other than [ADR-007](docs/adr/007-phase1-foundation-decisions.md)
- **Never load for implementation or verification** (unless user pastes a path for scoring):
  - `~/.cursor/projects/**/agent-transcripts/**` — prior chat archaeology
  - `~/.cursor/projects/**/terminals/**` — poll with `Await` once; do not double-Read
  - `evals/live-model/README.md`, `evals/live-model/manifest.json` — run scripts instead
- **Do not re-read** once loaded in the same session:
  - Global harness: `~/.cursor/HARNESS.md`, `~/.cursor/TOOLING.md`, `~/.cursor/memory/MEMORY.md`, `~/.cursor/rules/*.mdc`, `~/.cursor/skills/**/SKILL.md`, orchestration rules — unless the task is explicitly harness/skill/meta work
  - Repo contract: this file, `docs/phase1-implementation-spec.md`, `docs/handoff-parallel-agent.md`, `docs/handoff-verification-agent.md` — unless you or the user edited them since last read
- **Eval/reporting commands** (`/token-efficiency`, `/live-eval`, `/live-eval-phase2`): run `./scripts/token-efficiency.sh`, `./scripts/run-live-evals.sh`, or `./scripts/run-live-evals-phase2.sh`; do not Read scorer scripts or eval README/manifest

## Session hygiene

Start a **new chat** for unrelated tasks (implementation vs eval vs review vs verification). Do not carry long implementation or analysis context into verification or scope-refusal scenarios.

When the user shifts from analysis to implement / verify / assist, treat it as a **new task**: run smoke tests before reading prior chat transcripts.

| Task type | Start with | Do not load |
|-----------|------------|-------------|
| **Verification** | This file + service `AGENTS.md` + smoke scripts | Prior `agent-transcripts/`, `evals/`, superpowers skills |
| **Implementation** | Phase spec + scoped service `AGENTS.md` | Transcripts, architecture docs unless required |
| **Analysis** | User-directed reads OK | Global harness re-reads; still avoid transcripts unless asked |
| **Eval / metrics** | `./scripts/token-efficiency.sh`, `./scripts/run-live-evals.sh`, or `./scripts/run-live-evals-phase2.sh` only | Scorer source, `evals/live-model/README.md`, `evals/live-model-phase2/README.md` |

Verification handoff: [docs/handoff-verification-agent.md](docs/handoff-verification-agent.md). Use `/verify-phase2` in Cursor.

## Parallel work

- **Separate chats**: planning → implementation ([handoff-parallel-agent.md](docs/handoff-parallel-agent.md)); implementation → verification ([handoff-verification-agent.md](docs/handoff-verification-agent.md))
- **In-session subagents**: max 3 tracks with explicit file boundaries — Integration (Compose + smoke), Backend (`services/*`, `jobs/*`), Frontend (`apps/*`)
- **Parent owns**: synthesis, conflict resolution, `./scripts/smoke-test.sh` + `./scripts/smoke-test-phase2.sh`
- **Do not parallelize**: shared `docker-compose.dev.yml`, integration debugging, related failure chains

Debug discipline: grep before Read on large files; one Read per hypothesis; use `Await` once for background shells (no double-Read of terminal files).

## Context efficiency

Agents must keep sessions lean — eval scores are meaningless if real work burns context:

| Signal | Target | Investigate / fail |
|--------|--------|------------------|
| `harness_reread_count` | 0 | any re-read of `~/.cursor/**`, transcripts, or terminals |
| `duplicate_read_count` | 0–3 | >3 (`./scripts/token-efficiency.sh --strict` exits 1) |
| Same repo file read twice | avoid | grep or search first; one Read per hypothesis |
| Session type | fresh chat | implementation vs verification vs eval never share one thread |

After verification or eval scoring, run `./scripts/token-efficiency.sh --strict` and confirm `harness_reread_count: 0`.

## Repo contract

- **Primary language (Phase 1)**: Go (State Service)
- **Infrastructure**: Docker Compose for local dev ([ADR-007](docs/adr/007-phase1-foundation-decisions.md))
- **Event schemas**: Avro in `schemas/avro/`
- **Do not edit**: `.cursor/plans/` plan files attached to chat sessions
- **Default tenant ID**: `00000000-0000-0000-0000-000000000001` (single-institution mode)

## Commands

Run from repository root after Phase 1 scaffold exists:

```bash
# Start local stack
docker compose -f docker-compose.dev.yml up -d --wait

# Apply seed data
./scripts/seed.sh

# Register Avro schemas
./scripts/register-schemas.sh

# State Service — tests
cd services/state-service && go test ./...

# State Service — run locally (if not using Compose build)
cd services/state-service && go run ./cmd/server

# End-to-end smoke test
./scripts/smoke-test.sh

# Phase 2 smoke test (after Flink job RUNNING)
./scripts/smoke-test-phase2.sh

# Token efficiency (latest agent transcript)
./scripts/token-efficiency.sh

# Tear down
docker compose -f docker-compose.dev.yml down -v
```

Copy environment variables from `.env.example` before first run.

## Layout

| Path | Purpose |
|------|---------|
| `docs/` | Architecture, ADRs, Phase 1 spec (read-only for implementers) |
| `schemas/avro/` | Kafka Avro schema definitions |
| `services/state-service/` | Go REST + consumer + outbox |
| `services/alert-service/` | Go alerts REST + WebSocket + consumer |
| `jobs/compliance-cep/` | Flink CEP job (Java) |
| `apps/alert-console/` | Next.js alert UI |
| `mocks/core-banking/` | CDC source DB migrations and seed |
| `scripts/` | Seed, smoke test, schema registration |
| `.github/workflows/` | CI and schema compatibility |

## Coding rules

- Match patterns in [docs/data-flow.md](docs/data-flow.md) for event envelopes and idempotency keys.
- All entity tables include `tenant_id` (default single tenant per [ADR-007](docs/adr/007-phase1-foundation-decisions.md)).
- Legal entity hierarchy max depth: 3 levels (parent → subsidiary → sub-subsidiary).
- Validate external input at API and consumer boundaries.
- No secrets in code; use `.env` (never commit `.env`).
- Imports at top of file; no inline imports unless documented circular-dependency reason.

## Out of scope (Phase 1)

Do **not** add in Phase 1 PRs:

- Apache Flink / CEP jobs
- Cedar Policy Service / GoRules Zen
- immudb audit ledger
- Neo4j / Graph Service
- Next.js UI / WebSocket alert console
- Keycloak / auth middleware
- Regulatory reporting (XBRL)

Phase/deferral rationale: [docs/roadmap.md](docs/roadmap.md).

## Testing rules

- Unit tests for **every** `services/state-service/internal/*` package plus `cmd/server` migration wiring.
- Minimum **35%** total statement coverage in state-service (`./scripts/run-live-evals.sh --full` enforces).
- Integration tests may use testcontainers or Compose (match CI approach).
- `scripts/smoke-test.sh` must pass before claiming Phase 1 complete.
- Do not weaken tests to make CI green.
- After implementation, run `./scripts/token-efficiency.sh --strict` in a fresh verification chat.

## Review guidelines

Block on P0/P1 issues:

- Broken event envelope or idempotency contract
- Missing outbox pattern (direct Kafka publish without durability)
- Schema changes without BACKWARD compatibility check
- Secrets committed or logged
- Scope creep into Phase 2+ components

Style-only feedback is non-blocking unless it hides a bug.

## Definition of done (Phase 1)

Before claiming Phase 1 complete, all must exit 0:

```bash
docker compose -f docker-compose.dev.yml up -d --wait
./scripts/seed.sh
cd services/state-service && go test ./...
./scripts/smoke-test.sh
```

Done also means:

- [Phase 1 exit criteria checklist](docs/phase1-implementation-spec.md#10-phase-1-exit-criteria-checklist) in spec is satisfied
- PR description includes test plan and checklist copy
- [docs/review/phase1-review-checklist.md](docs/review/phase1-review-checklist.md) self-reviewed

## Definition of done (Phase 2)

Before claiming Phase 2 complete, all must exit 0:

```bash
docker compose -f docker-compose.dev.yml up -d --wait
./scripts/seed.sh
./scripts/smoke-test.sh
./scripts/smoke-test-phase2.sh
cd services/alert-service && go test ./...
```

Done also means [Phase 2 exit criteria checklist](docs/phase2-implementation-spec.md#14-phase-2-exit-criteria-checklist) in spec is satisfied.

## Handoff between agents

- **Planning agent** produces specs and ADRs under `docs/`; does not implement services.
- **Implementation agent** builds from [docs/phase1-implementation-spec.md](docs/phase1-implementation-spec.md) or [docs/phase2-implementation-spec.md](docs/phase2-implementation-spec.md).
- **Verification agent** runs [docs/handoff-verification-agent.md](docs/handoff-verification-agent.md); minimal diffs only.
- After implementation, run review using [docs/review/phase1-review-checklist.md](docs/review/phase1-review-checklist.md).

## References

- [docs/roadmap.md](docs/roadmap.md) — Phases and exit criteria
- [docs/architecture.md](docs/architecture.md) — C4 and component map
- [docs/domain-model.md](docs/domain-model.md) — Entities and personas
- [docs/adr/](docs/adr/) — Architecture decisions
