## Summary

<!-- 1–3 bullets: what changed and why -->

-

## Test plan

<!-- Commands run and results -->

- [ ] `cd services/state-service && go test ./...`
- [ ] `./scripts/smoke-test.sh` (requires Compose stack)
- [ ] `./scripts/run-live-evals.sh` (optional, mechanical checks)

## Phase 1 checklist

<!-- Complete for Phase 1 changes; mark N/A for docs-only PRs -->

- [ ] `docker compose -f docker-compose.dev.yml up` starts all services
- [ ] `scripts/seed.sh` loads seed data (≥10 institutions, ≥100 accounts, 500 instruments)
- [ ] `GET /api/v1/personas?personaType=Institution` returns ≥ 10 records
- [ ] Core-banking UPDATE propagates to state within 5s; `state_version` increments
- [ ] Outbox publishes to `twin.state.updated`
- [ ] Schema compat CI passes (BACKWARD-compatible Avro changes only)
- [ ] No Phase 2+ components added ([AGENTS.md § Out of scope](../AGENTS.md#out-of-scope-phase-1))

## Review

- [ ] Self-reviewed against [phase1-review-checklist.md](../docs/review/phase1-review-checklist.md)
