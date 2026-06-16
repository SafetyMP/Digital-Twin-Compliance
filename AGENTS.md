# AGENTS.md

Operational contract for coding agents working in this repository.

## Current phase

**Phase 3** — Rules engine and audit ledger. Full spec: [docs/phase3-implementation-spec.md](docs/phase3-implementation-spec.md). Phase 2 spec: [docs/phase2-implementation-spec.md](docs/phase2-implementation-spec.md). Phase 1 spec: [docs/phase1-implementation-spec.md](docs/phase1-implementation-spec.md).

Architecture and domain docs live under [docs/](docs/). Do not implement Phase 4+ components unless explicitly requested.

## Context loading

Minimize token use — load only what the task requires:

- **Always load**: this file, [docs/phase3-implementation-spec.md](docs/phase3-implementation-spec.md) (Phase 3) or [docs/phase2-implementation-spec.md](docs/phase2-implementation-spec.md) / [docs/phase1-implementation-spec.md](docs/phase1-implementation-spec.md) when earlier-phase-only work
- **For Go State Service work**: [services/state-service/AGENTS.md](services/state-service/AGENTS.md)
- **For Go Alert Service work**: [services/alert-service/AGENTS.md](services/alert-service/AGENTS.md)
- **For Go Audit / Cedar / Decision Service work**: `services/<svc>/AGENTS.md` (created during Phase 3 implementation)
- **For envelope / idempotency / outbox / audit Kafka tasks**: [docs/data-flow.md](docs/data-flow.md)
- **Do not load unless the task explicitly requires**: [docs/architecture.md](docs/architecture.md), [docs/domain-model.md](docs/domain-model.md), [docs/compliance-mapping.md](docs/compliance-mapping.md), [docs/roadmap.md](docs/roadmap.md), ADRs other than [ADR-007](docs/adr/007-phase1-foundation-decisions.md)
- **Never load for implementation or verification** (unless user pastes a path for scoring):
  - `~/.cursor/projects/**/agent-transcripts/**` — prior chat archaeology
  - `~/.cursor/projects/**/terminals/**` — poll with `Await` once; do not double-Read
  - `evals/harness/`, `evals/fixtures/` — scoring answer keys during live behavioral evals
  - `evals/live-model/README.md`, `evals/live-model/manifest.json` — run scripts instead
- **Do not re-read** once loaded in the same session:
  - Global harness: `~/.cursor/HARNESS.md`, `~/.cursor/TOOLING.md`, `~/.cursor/memory/MEMORY.md`, `~/.cursor/rules/*.mdc`, `~/.cursor/skills/**/SKILL.md`, orchestration rules — unless the task is explicitly harness/skill/meta work
  - Repo contract: this file, `docs/phase1-implementation-spec.md`, `docs/handoff-parallel-agent.md`, `docs/handoff-verification-agent.md`, `docs/handoff-continuation.md` — unless you or the user edited them since last read
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

Continuation handoff (multi-session work): [docs/handoff-continuation.md](docs/handoff-continuation.md) — paste template with Done/Blocked/Next; do not read prior transcripts.

## Agent learning

Learning is **externalized memory**, not chat recall. Agents do not retain prior sessions unless knowledge is written to files.

### First-read contract

Every new session, before debugging or editing:

1. This file — especially **Repo gotchas** below
2. Phase spec ([phase2](docs/phase2-implementation-spec.md) or [phase1](docs/phase1-implementation-spec.md))
3. Scoped `services/*/AGENTS.md` for the service you touch

Then `grep` before `Read` on large files.

### Capture checklist (session close-out)

Capture when the user **corrects** you, states a **durable preference**, or you fix a **non-obvious gotcha**:

| Learning | Write to |
|----------|----------|
| Repo-specific gotcha | **Repo gotchas** below (one line, lead with the rule) |
| Service-specific pattern | `services/<svc>/AGENTS.md` |
| Cross-project preference | `~/.cursor/memory/MEMORY.md` or user rule |
| Repeatable workflow | Cursor skill |

Do **not** capture one-off task state (use [handoff-continuation.md](docs/handoff-continuation.md)). Do **not** duplicate phase-spec content.

### Retention metric

Track **repeat discovery**: a new session re-fixes a gotcha already in **Repo gotchas**. If that happens, shorten or elevate the gotcha line, or add a behavioral scenario. Run `./scripts/check-agent-learning.sh` for hygiene checks.

## Parallel work

**Decision rule**: default to **serial**. Parallelize only when tracks have **truly independent, non-overlapping file boundaries** and no shared integration state. Serial is usually faster for this stack because most work converges on the shared Compose file and smoke tests.

- **Separate chats**: planning → implementation ([handoff-parallel-agent.md](docs/handoff-parallel-agent.md)); implementation → verification ([handoff-verification-agent.md](docs/handoff-verification-agent.md))
- **In-session subagents**: max 3 tracks with explicit file boundaries — Integration (Compose + smoke), Backend (`services/*`, `jobs/*`), Frontend (`apps/*`)
- **Parent owns**: synthesis, conflict resolution, `./scripts/smoke-test.sh` + `./scripts/smoke-test-phase2.sh`
- **Do not parallelize**: shared `docker-compose.dev.yml`, integration debugging, related failure chains

Subagent reliability:

- Give each subagent a **bounded, single-track scope** with explicit non-overlapping file boundaries; do not let two tracks write the same path.
- Parent **checkpoints after each track returns** — confirm the track's claimed output exists before merging.
- If a subagent **times out or errors** (e.g. PING timeout), do not silently absorb partial work. Either **re-delegate** the remaining scope to a fresh subagent or finish it **serially**, then **re-verify** (smoke tests) before treating it as done.

Debug discipline: grep before Read on large files; one Read per hypothesis. To inspect a command's output, use `Await` on the live shell or re-run the command — never `Read` files under `~/.cursor/**/terminals/`. Reading terminal files (even once) is scored as `harness_reread_count` and inflates the efficiency pillar.

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

Run from repository root:

```bash
# First run: copy env (scripts also fall back to .env.example defaults)
cp .env.example .env

# Start local stack
docker compose -f docker-compose.dev.yml up -d --wait

# Apply seed data (also creates Kafka topics when stack is healthy)
./scripts/seed.sh

# If Debezium/state lag after seed (CI order):
./scripts/register-debezium-connector.sh
docker compose -f docker-compose.dev.yml restart state-service && docker compose -f docker-compose.dev.yml up -d --wait state-service

# Phase 2: restart alert-service after seed so consumers see new topics
docker compose -f docker-compose.dev.yml restart alert-service && docker compose -f docker-compose.dev.yml up -d --wait alert-service

# Phase 2: ensure Flink CEP job RUNNING
./scripts/submit-flink-job.sh

# State Service — tests
cd services/state-service && go test ./...

# Alert Service — tests (store package needs Docker for testcontainers)
cd services/alert-service && go test ./...

# End-to-end smoke tests
./scripts/smoke-test.sh
./scripts/smoke-test-phase2.sh

# Twin-path canary (after state-service restart / Debezium register)
./scripts/wait-outbox-drained.sh
./scripts/verify-state-twin-pipeline.sh

# Token efficiency (latest agent transcript)
./scripts/token-efficiency.sh

# Agent learning hygiene (gotchas, retention scenario, fixtures)
./scripts/check-agent-learning.sh

# Tear down
docker compose -f docker-compose.dev.yml down -v
```

`seed.sh` conditionally runs schema registration and Debezium connector setup when the stack is healthy; use the explicit restart/submit steps above when smoke tests fail after a cold start.

Long-running commands (stack bring-up, image pulls, smoke tests can exceed 7 min):

- **Pre-pull images** before timed steps so `up -d --wait` is not blocked on registry downloads.
- **Background** long commands and **poll with `Await` once** on a sentinel line (e.g. `Phase 2 smoke test passed`) instead of blocking the foreground.
- Size `block_until_ms` to the command's expected runtime so it is not killed mid-pull.

## Repo gotchas

Hard-won fixes from Phase 2 — check here before rediscovering them ([capturing-learnings](docs/) → this section, not global memory):

- **Debezium numerics/dates**: base64-encoded numerics and epoch-day dates break instrument CDC; decode in `services/state-service/internal/consumer/debezium_numeric.go` (see `debezium_numeric_test.go`).
- **Redis host port is `6380`**: Compose maps `6380:6379` to avoid clashing with a local Redis; `REDIS_URL=redis://localhost:6380/0` in `.env.example`.
- **Pre-create Kafka topics**: `domain.events.public.payments`, `domain.events.dlq`, `compliance.alerts`, `compliance.alerts.dlq`, and `twin.state.updated` must exist before the Flink job/consumers start — run `./scripts/create-kafka-topics.sh` (invoked by `seed.sh`).
- **Flink `JobConfig` must implement `Serializable`**: otherwise job submission fails. Submit with `./scripts/submit-flink-job.sh` (uses `basename(jar)` and a Docker Maven fallback).
- **`mvn` is not assumed on host**: `./scripts/run-live-evals-phase2.sh --full` fails locally without Maven. Run CEP tests via Docker: `docker run --rm -v "$PWD/jobs/compliance-cep:/app" -w /app maven:3.9-eclipse-temurin-17 mvn -q test` (CI uses Maven directly).
- **CEP jar staleness in CI**: `flink-job-submitter` in Compose builds `jobs/compliance-cep/target/*.jar` at stack bring-up; host/CI changes are not in that jar until `mvn package` runs again before `./scripts/submit-flink-job.sh` (`mvn test` alone is insufficient).
- **Flink `latest` offsets need a fresh consumer group**: `OffsetsInitializer.latest()` does not override committed offsets for an existing group. CI passes `CEP_CONSUMER_GROUP_SUFFIX` to `submit-flink-job.sh`; local resubmits after seed should cancel the old job and use a new suffix or `earliest` when debugging replay.
- **Restart `alert-service` after `./scripts/seed.sh`**: same pattern as state-service after Debezium — long-running consumers started before topics/seed are wired can miss the smoke-test window (CI restarts alert-service post-seed).
- **Rebuild `state-service` image after Go consumer/enrichment changes**: `docker compose -f docker-compose.dev.yml build state-service` then restart — stale container code leaves `twin_personas` with old `liquidity.lcr` / base64 CDC columns even when core banking updated.
- **INT-M001 smoke-test localization**: Redis `vel:{tenant}:{account}:1h` > 50 means Debezium → Flink velocity logic succeeded; if open alerts stay empty, debug Flink → `compliance.alerts` → alert-service (not payment CDC or parsers). `./scripts/smoke-test-phase2.sh` dumps offsets/DB rows on failure.
- **Alert consumer must not stall on DLQ failure**: `services/alert-service/internal/consumer` commits after handler errors (DLQ publish failure included); `continue` without commit blocks the partition and `/api/v1/health` stays green.
- **BASEL-M001 smoke-test localization**: requires `legal_entities.lcr` (and related columns) in core banking — enrichment no longer hard-codes liquidity. Smoke lowers Delta Independent Bank to `lcr = 0.90`; if open alerts stay empty, check core row → `twin_personas.current_state` liquidity block → `twin.state.updated` → Flink `lcr:*` Redis key. Smoke waits on twin `liquidity.lcr` mirror before Redis, then alert API.
- **INT-M002 smoke-test localization**: needs two instruments with matching `owner_entity_id` + `counterparty_id` (see `002_phase2_exposure.sql`). After updates, Redis `exp:{tenant}:{owner}:{counterparty}` should exceed `CEP_EXPOSURE_LIMIT_EUR` (10M); if empty, check `twin_personas.current_state` mirrored notional/owner/counterparty (state-service) before Flink; smoke waits on twin mirror then Redis — if twin never updates, restart state-service after Debezium/seed.
- **Cross-service numeric contract**: Debezium CDC decimals arrive as strings in Go; state-service `enrichInstrumentState` / `enrichInstitutionState` normalize before `twin.state.updated`. Flink must still use `JsonParsers.parseDouble` defensively. Golden fixtures live under `contracts/kafka/`; run `./scripts/check-kafka-contracts.sh`.
- **Outbox publisher throughput**: `kafka-go` `Writer` defaults `BatchTimeout=1s` — per-row `WriteMessages` ≈ 1 msg/s. Publisher batches rows and sets explicit `BatchTimeout` (`OUTBOX_BATCH_TIMEOUT`, default `10ms`). After Debezium register + `state-service` restart, run `./scripts/wait-outbox-drained.sh` before verify — restart leaves a large `outbox` backlog unrelated to the canary.
- **Verification after edits** (fail fast before Docker smoke):

| Touch | Minimum |
|-------|---------|
| `services/state-service/` | `cd services/state-service && go test ./...` |
| `services/alert-service/` | `cd services/alert-service && go test ./...` |
| `jobs/compliance-cep/` | `cd jobs/compliance-cep && mvn test` |
| CDC enrichment + CEP parsers | both `go test` and `mvn test`; `./scripts/check-kafka-contracts.sh` must pass |
| Seeds / smoke scripts | `bash -n scripts/smoke-test-phase2.sh scripts/smoke-lib-phase2.sh scripts/verify-state-twin-pipeline.sh scripts/wait-outbox-drained.sh`; full smoke needs stack |
| Single smoke scenario | `SMOKE_PHASE2_ONLY=M002 ./scripts/smoke-test-phase2.sh` (optional `SMOKE_PHASE2_SKIP_PREREQS=1` when stack warm) |

CI runs Go + `mvn test` + `check-kafka-contracts.sh` before `docker compose up`; smoke remains the integration gate.

- **Phase 2 CI bring-up order** (`.github/workflows/ci.yml` — do not reorder without updating smoke assumptions):

  1. Unit tests + Kafka contracts (fail fast)
  2. `docker compose up -d --wait`
  3. `./scripts/seed.sh` → restart `alert-service`
  4. Register schemas + Debezium → restart `state-service` → `./scripts/wait-outbox-drained.sh` → `./scripts/verify-state-twin-pipeline.sh`
  5. `mvn package -DskipTests` (host jar; Compose submitter image may be stale until this)
  6. `./scripts/submit-flink-job.sh` with fresh `CEP_CONSUMER_GROUP_SUFFIX`
  7. `./scripts/smoke-test.sh` then `./scripts/smoke-test-phase2.sh`
  8. `./scripts/check-coverage-gates.sh`

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
| `contracts/kafka/` | Golden Kafka payload fixtures (cross-service API contract) |
| `scripts/` | Seed, smoke test, schema registration |
| `.github/workflows/` | CI and schema compatibility |

## Coding rules

- **Kafka payloads are published APIs** — any shape on `twin.state.updated`, `domain.events.public.*`, or `compliance.alerts` that crosses a service boundary must have a golden fixture in [`contracts/kafka/`](contracts/kafka/README.md), a publisher test (producer side), and a consumer test (parser side). Run `./scripts/check-kafka-contracts.sh` before claiming cross-service work done. Do not duplicate fixtures under `services/` or `jobs/`.
- Match patterns in [docs/data-flow.md](docs/data-flow.md) for event envelopes and idempotency keys.
- All entity tables include `tenant_id` (default single tenant per [ADR-007](docs/adr/007-phase1-foundation-decisions.md)).
- Legal entity hierarchy max depth: 3 levels (parent → subsidiary → sub-subsidiary).
- Validate external input at API and consumer boundaries.
- No secrets in code; use `.env` (never commit `.env`).
- Imports at top of file; no inline imports unless documented circular-dependency reason.
- **Verification floor (not waivable)**: any edit to a code file (`.go`, `.java`, `.ts`/`.tsx`) requires at minimum a build/compile or the touched package's tests before claiming done — including comment- or doc-only edits. A user calling a change "trivial" scopes *how much* to verify (a one-line comment → `go build ./...` or `go test ./internal/<pkg>/...`; logic changes → package tests + relevant smoke), never *whether* to verify.

## Scope by phase

### Phase 3 (current)

In scope unless a task is explicitly Phase 2-only or earlier:

- Cedar Policy Service (`services/cedar-service/`), Decision Service (`services/decision-service/`)
- Audit Service + immudb (`services/audit-service/`)
- Audit Explorer UI (`apps/audit-explorer/`)
- `policies/cedar/`, `policies/zen/`, `smoke-test-phase3.sh`, policy CI gates

### Out of scope (Phase 4+)

Do **not** add unless the task explicitly targets a later phase:

- Neo4j / Graph Service
- Simulation Service (Python stress/contagion)
- Regulatory reporting (XBRL/SDMX)
- Keycloak / full OIDC auth middleware (mock principal only in Phase 3 — ADR-009 D20)
- Flink → Zen hot-path migration (Phase 3b stretch; default serial is Phase 3a audit path only)

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

**Phase 2a (integration stable — merge bar):** two consecutive green CI runs with smoke steps 1–9, `check-kafka-contracts.sh` in CI, and PR checklist items covered by smoke (rules fire, API, WS, ack, Redis keys).

**Phase 2b (repo contract complete):** Phase 2a on `main`, behavior eval pillar (≥5/6 scenarios, ≥80% pass rate), `./scripts/token-efficiency.sh --strict` on verification chat, and honest §14 checklist (defer/document soak/p99 if not measured).

Before claiming Phase 2a complete, all must exit 0:

```bash
docker compose -f docker-compose.dev.yml up -d --wait
./scripts/seed.sh
./scripts/smoke-test.sh
./scripts/smoke-test-phase2.sh
cd services/alert-service && go test ./...
```

Done also means:

- [Phase 2 exit criteria checklist](docs/phase2-implementation-spec.md#14-phase-2-exit-criteria-checklist) in spec is satisfied (Phase 2b)
- `./scripts/token-efficiency.sh --strict` passes on the session transcript (`harness_reread_count: 0`, `duplicate_read_count ≤ 3`); paste output in the PR (Phase 2b)
- Behavior eval pillar populated: ≥ 5/6 live scenarios stored under `evals/live-model-phase2/results/` (see [Behavior evals](#behavior-evals-phase-2)) (Phase 2b)

## Behavior evals (Phase 2)

Pass/fail is **invariant-gated** (git diff + command ordering), not prose regex. Prose quality is advisory only. Gate definitions live in `evals/harness/gates.json` — do not read that file or `evals/fixtures/` during a live eval session.

The mechanical and DoD pillars run in CI; the **behavior pillar** (adversarial live scenarios) uses `./scripts/run-behavioral-eval.sh` or manual fresh chats. CI regresses pass/fail fixtures via `./scripts/run-eval-fixtures.sh`.

To populate the pillar before claiming Phase 2 done:

1. For each scenario in `evals/live-model-phase2/scenarios/`, open a **new** Cursor chat and paste its Prompt section (or run `./scripts/run-behavioral-eval.sh --phase2 --scenario <id>`).
2. Let the agent run to completion (or stop after it claims done).
3. Score and persist the result (includes workspace git diff):

```bash
./scripts/score-eval-session.sh \
  --manifest evals/live-model-phase2/manifest.json \
  --scenario <id> \
  --transcript <path-to.jsonl> \
  --baseline-ref HEAD \
  --write-result evals/live-model-phase2/results/<id>/run-$(date +%Y%m%dT%H%M%S).json
```

Pass bar: **≥ 80% pass rate** per scenario over `runs_per_scenario` (default 3) from manifest, and **≥ 5/6** scenarios meeting that bar. An empty `evals/live-model-phase2/results/` means the behavior pillar is unmet and Phase 2 is **not** done.

The behavior pillar scores **process discipline** (verification order, scope boundaries, invariant preservation) and **contract retention** (loads `AGENTS.md` gotchas before deep debugging) — not code correctness. Retention scenario: `debug-int-m001-retention`. See [docs/eval-harness-scope.md](docs/eval-harness-scope.md).

Measure session efficiency separately with `./scripts/token-efficiency.sh --strict` (the Efficiency pillar and DoD gate); do not fold `--fail-on-harness-rereads` into the behavior score.

## Handoff between agents

- **Planning agent** produces specs and ADRs under `docs/`; does not implement services.
- **Implementation agent** builds from [docs/phase1-implementation-spec.md](docs/phase1-implementation-spec.md), [docs/phase2-implementation-spec.md](docs/phase2-implementation-spec.md), or [docs/phase3-implementation-spec.md](docs/phase3-implementation-spec.md).
- **Verification agent** runs [docs/handoff-verification-agent.md](docs/handoff-verification-agent.md); minimal diffs only.
- **Continuation** across sessions uses [docs/handoff-continuation.md](docs/handoff-continuation.md); outcomes only, no transcript archaeology.
- After implementation, run review using [docs/review/phase1-review-checklist.md](docs/review/phase1-review-checklist.md).

## References

- [docs/roadmap.md](docs/roadmap.md) — Phases and exit criteria
- [docs/architecture.md](docs/architecture.md) — C4 and component map
- [docs/domain-model.md](docs/domain-model.md) — Entities and personas
- [docs/adr/](docs/adr/) — Architecture decisions
