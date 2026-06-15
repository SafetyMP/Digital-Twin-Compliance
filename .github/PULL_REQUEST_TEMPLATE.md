## Summary

<!-- 1–3 bullets: what changed and why -->

-

## Test plan (three pillars)

### Product (mechanical + DoD)

- [ ] `cd services/state-service && go test ./...`
- [ ] `cd services/alert-service && go test ./...` (Phase 2 changes)
- [ ] `./scripts/run-live-evals.sh` (mechanical)
- [ ] `./scripts/run-live-evals.sh --full` or `./scripts/check-coverage-gates.sh` (coverage ≥35% state-service)
- [ ] `./scripts/smoke-test.sh` (requires Compose stack)
- [ ] `./scripts/smoke-test-phase2.sh` (Phase 2 changes; Flink RUNNING)
- [ ] `./scripts/run-live-evals-phase2.sh` (Phase 2 mechanical)

### Behavior (live scenarios — fresh chat)

- [ ] Relevant scenario scored with `--fail-on-harness-rereads --fail-on-efficiency` (if adversarial session run)
- [ ] `./scripts/run-efficiency-fixtures.sh` (CI fixture regression)

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
- [ ] No Phase 3+ components added ([AGENTS.md § Scope by phase](../AGENTS.md#scope-by-phase))

## Phase 2 checklist

<!-- Complete for Phase 2 changes; mark N/A otherwise -->

- [ ] Flink job `RUNNING` on `:8082`
- [ ] `INT-M001` / `BASEL-M001` alerts visible via alert-service API
- [ ] Alert-service package tests present (`run-live-evals-phase2.sh` mechanical)

## Review

- [ ] Self-reviewed against [phase1-review-checklist.md](../docs/review/phase1-review-checklist.md)
