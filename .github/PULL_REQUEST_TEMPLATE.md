## Summary

<!-- 1–3 bullets: what changed and why -->

-

## Test plan (three pillars)

### Product (mechanical + DoD)

- [ ] `cd services/state-service && go test ./...`
- [ ] `cd services/alert-service && go test ./...` (Phase 2 changes)
- [ ] `cd services/cedar-service && go test ./...` (Phase 3 policy changes)
- [ ] `cd services/decision-service && go test ./...` (Phase 3 policy changes)
- [ ] `cd services/audit-service && go test ./...` (Phase 3 audit changes)
- [ ] `./scripts/run-policy-ci.sh` (Cedar/Zen policy changes)
- [ ] `./scripts/check-kafka-contracts.sh` (Kafka payload contract; required for cross-service changes)
- [ ] `./scripts/run-live-evals.sh` (mechanical)
- [ ] `./scripts/run-live-evals.sh --full` or `./scripts/check-coverage-gates.sh` (coverage ≥35% state-service)
- [ ] `./scripts/smoke-test.sh` (requires Compose stack)
- [ ] `./scripts/smoke-test-phase2.sh` (Phase 2 changes; Flink RUNNING; subset: `SMOKE_PHASE2_ONLY=M002`; gates twin mirror → Redis → alert for twin-path rules)
- [ ] `./scripts/smoke-test-phase3.sh` (Phase 3 changes; Cedar/Zen evaluate, audit chain, Audit Explorer)
- [ ] `./scripts/wait-outbox-drained.sh` then `./scripts/verify-state-twin-pipeline.sh` (after state-service restart / Debezium register)
- [ ] `./scripts/run-live-evals-phase2.sh` (Phase 2 mechanical)

### Behavior (live scenarios — fresh chat)

- [ ] Relevant scenario scored with `--fail-on-harness-rereads --fail-on-efficiency` (if adversarial session run)
- [ ] `./scripts/run-eval-fixtures.sh` (CI fixture regression)

### Efficiency (verification chats)

- [ ] `./scripts/token-efficiency.sh --strict` on the session transcript — `harness_reread_count: 0`, `duplicate_read_count ≤ 3`. Paste output below:

```
<!-- paste ./scripts/token-efficiency.sh --strict <session.jsonl> output here -->
```

Optional report: `./scripts/report-eval-scorecard.sh --all [--full]`

## Phase 1 checklist

<!-- Complete for Phase 1 changes; mark N/A for docs-only PRs -->

- [ ] `docker compose -f docker-compose.dev.yml up` starts all services
- [ ] `scripts/seed.sh` loads seed data (≥10 institutions, ≥100 accounts, 500 instruments)
- [ ] `GET /api/v1/personas?personaType=Institution` returns ≥ 10 records
- [ ] Core-banking UPDATE propagates to state within 5s; `state_version` increments
- [ ] Outbox publishes to `twin.state.updated`
- [ ] Schema compat CI passes (BACKWARD-compatible Avro changes only)
- [ ] No Phase 4+ components added ([AGENTS.md § Scope by phase](../AGENTS.md#scope-by-phase))

## Phase 2 checklist

<!-- Complete for Phase 2 changes; mark N/A otherwise -->

- [ ] Flink job `RUNNING` on `:8082`
- [ ] `INT-M001` / `INT-M002` / `BASEL-M001` alerts visible via alert-service API (smoke gates intermediate twin/Redis keys, not API alone)
- [ ] Alert-service package tests present (`run-live-evals-phase2.sh` mechanical)

## Phase 3 checklist

<!-- Complete for Phase 3 changes; mark N/A otherwise -->

- [ ] `./scripts/run-policy-ci.sh` passes (Cedar Analyzer + Zen tests)
- [ ] Cedar evaluate deny→allow and Zen BASEL-R001 paths covered by `./scripts/smoke-test-phase3.sh`
- [ ] `./scripts/verify-audit-chain.sh` passes; alerts carry `evidenceRef` where applicable
- [ ] Audit Explorer loads via same-origin `/api/audit/*` proxies (no browser fetch to `:8090`)

## Review

- [ ] Self-reviewed against [phase1-review-checklist.md](../docs/review/phase1-review-checklist.md)
- [ ] Self-reviewed against [phase2-exit-checklist.md](../docs/review/phase2-exit-checklist.md) (Phase 2 paths)
- [ ] Self-reviewed against [phase3-exit-checklist.md](../docs/review/phase3-exit-checklist.md) (Phase 3 paths)
