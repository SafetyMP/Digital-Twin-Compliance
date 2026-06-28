## Summary

<!-- 1–3 bullets: what changed and why -->

-

## Component

<!-- Which area does this PR touch? -->

- [ ] state-service / ingestion
- [ ] alert-service / alert-console
- [ ] compliance-cep (Flink)
- [ ] cedar-service / decision-service / policies
- [ ] audit-service / audit-explorer
- [ ] schemas / Kafka contracts
- [ ] infra / CI / docs only

## Test plan

<!-- Minimum for human contributors; CI runs the full matrix on merge -->

- [ ] Unit tests for touched packages (`go test`, `mvn test`, or `npm run build` as applicable)
- [ ] `./scripts/check-kafka-contracts.sh` (if cross-service Kafka payloads changed)
- [ ] `./scripts/run-policy-ci.sh` (if `policies/**` or policy services changed)
- [ ] `./scripts/smoke-test.sh` (ingestion paths)
- [ ] `./scripts/smoke-test-phase2.sh` (monitoring paths; subset: `SMOKE_PHASE2_ONLY=M002`)
- [ ] `./scripts/smoke-test-phase3.sh` (policy/audit paths)
- [ ] Schema compat / Avro BACKWARD compatibility considered

**Commands run locally (paste or summarize):**

```
<!-- e.g. cd services/audit-service && go test ./... -->
```

## Review

- [ ] Self-reviewed against the relevant checklist in `docs/review/` for touched areas
- [ ] [CHANGELOG.md](../CHANGELOG.md) updated under `[Unreleased]` (if user-visible)
- [ ] No scope creep into [ROADMAP.md planned work](../ROADMAP.md#planned-not-built-yet) without issue discussion

---

<details>
<summary>Maintainer / agent eval checklist (optional)</summary>

### Mechanical evals

- [ ] `./scripts/run-live-evals.sh` and `./scripts/run-live-evals-phase2.sh`
- [ ] `./scripts/run-eval-fixtures.sh`
- [ ] `./scripts/check-coverage-gates.sh` or `./scripts/run-live-evals.sh --full`

### Behavior & efficiency (fresh agent session)

- [ ] `./scripts/token-efficiency.sh --strict` — paste output if applicable

</details>
