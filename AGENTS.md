# AGENTS.md

Operational contract for coding agents working in this repository.

## Current phase

**Phase 1** — Ingestion backbone and minimal twin. Full spec: [docs/phase1-implementation-spec.md](docs/phase1-implementation-spec.md).

Architecture and domain docs live under [docs/](docs/). Do not implement Phase 2+ components unless explicitly requested.

## Context loading

Minimize token use — load only what the task requires:

- **Always load**: this file, [docs/phase1-implementation-spec.md](docs/phase1-implementation-spec.md)
- **For Go State Service work**: [services/state-service/AGENTS.md](services/state-service/AGENTS.md)
- **For envelope / idempotency / outbox tasks only**: [docs/data-flow.md](docs/data-flow.md)
- **Do not load unless the task explicitly requires**: [docs/architecture.md](docs/architecture.md), [docs/domain-model.md](docs/domain-model.md), [docs/compliance-mapping.md](docs/compliance-mapping.md), [docs/roadmap.md](docs/roadmap.md), ADRs other than [ADR-007](docs/adr/007-phase1-foundation-decisions.md)
- **Do not re-read** once loaded in the same session:
  - Global harness: `~/.cursor/HARNESS.md`, `~/.cursor/TOOLING.md`, `~/.cursor/memory/MEMORY.md`, `~/.cursor/rules/*.mdc`, `~/.cursor/skills/**/SKILL.md`, orchestration rules — unless the task is explicitly harness/skill/meta work
  - Repo contract: this file, `docs/phase1-implementation-spec.md`, `docs/handoff-parallel-agent.md` — unless you or the user edited them since last read
- **Eval/reporting commands** (`/token-efficiency`, `/live-eval`): run `./scripts/token-efficiency.sh` or `./scripts/run-live-evals.sh`; do not Read scorer scripts or eval README/manifest

## Session hygiene

Start a **new chat** for unrelated tasks (implementation vs eval vs review). Do not carry long implementation context into verification or scope-refusal scenarios.

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

- Unit tests for store, consumer mapping, and outbox logic in `services/state-service`.
- Integration tests may use testcontainers or Compose (match CI approach).
- `scripts/smoke-test.sh` must pass before claiming Phase 1 complete.
- Do not weaken tests to make CI green.

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

## Handoff between agents

- **Planning agent** produces specs and ADRs under `docs/`; does not implement services.
- **Implementation agent** builds from [docs/phase1-implementation-spec.md](docs/phase1-implementation-spec.md).
- After implementation, run review using [docs/review/phase1-review-checklist.md](docs/review/phase1-review-checklist.md).

## References

- [docs/roadmap.md](docs/roadmap.md) — Phases and exit criteria
- [docs/architecture.md](docs/architecture.md) — C4 and component map
- [docs/domain-model.md](docs/domain-model.md) — Entities and personas
- [docs/adr/](docs/adr/) — Architecture decisions
