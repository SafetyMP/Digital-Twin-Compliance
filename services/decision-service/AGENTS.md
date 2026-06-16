# Decision Service — Agent Contract

Go REST API and Zen JDM evaluator for Phase 3 quantitative compliance rules.

Parent contract: [AGENTS.md](../../AGENTS.md). RuleDecision shape: [schemas/avro/rule-decision.avsc](../../schemas/avro/rule-decision.avsc).

## Package map

| Path | Role |
|------|------|
| `cmd/server/` | HTTP server startup, policy load, Kafka audit publisher |
| `internal/api/` | REST handlers (health, rules, evaluate) |
| `internal/engine/` | zen-go loader, rule-code mapping, evaluation |
| `internal/decision/` | RuleDecision types, input hash, audit helpers |
| `internal/audit/` | `compliance.audit.pending` publisher |
| `internal/config/` | Environment configuration |

## Commands

From `services/decision-service/`:

```bash
go test ./...
go run ./cmd/server
```

From repo root (policies mounted at `ZEN_POLICY_DIR`):

```bash
ZEN_POLICY_DIR=policies/zen go run ./services/decision-service/cmd/server
```

**Verification floor:** any edit requires `go test ./...` before claiming done.

## Rule code mapping

| ruleCode | JDM file |
|----------|----------|
| INT-R001 | int-r001.json |
| INT-R002 | int-r002.json |
| BASEL-R001 | basel-r001.json |
| COREP-R001 | corep-r001.json |
| COREP-R002 | corep-r002.json |

Fixtures: `policies/zen/fixtures/*.json` — table-driven tests in `internal/engine/fixtures_test.go`.

## API

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/health` | Liveness |
| GET | `/api/v1/rules` | Loaded models + versions |
| POST | `/api/v1/evaluate` | `{ "ruleCode", "input" }` → RuleDecision |

## Invariants

- **Audit on non-Allow** — `Deny`, `Flag`, `Escalate` publish `AuditPending` (`entryType: RuleDecision`) to `compliance.audit.pending`
- **RuleDecision fields** — `decisionId`, `evaluatedAt`, `inputHash` generated per evaluation
- **REST error shape** — `{"error":"...", "code":"NOT_FOUND"}` matching alert-service
- Default HTTP port `8092`; policies from `ZEN_POLICY_DIR` (default `policies/zen`)

## Gotchas

- Run tests from repo root or service dir — `PolicyDirFromRepoRoot()` walks up to find `policies/zen`
- Docker Compose mounts `./policies/zen:/policies/zen:ro` and sets `ZEN_POLICY_DIR=/policies/zen`
- zen-go `Engine` must be disposed on shutdown; loader reads JDM by filename key
