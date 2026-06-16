# Phase 3 Implementation Spec

Executable handoff for the implementation agent. Implements [roadmap.md](./roadmap.md) Phase 3 only.

**Prerequisites**: Phase 2 complete ([phase2-implementation-spec.md](./phase2-implementation-spec.md) §14 exit criteria). [ADR-002](./adr/002-cedar-decision-engine.md) (Cedar + Zen), [ADR-003](./adr/003-immudb-audit-ledger.md) (immudb), [ADR-009](./adr/009-phase3-foundation-decisions.md) (D15–D20).

**Related docs**: [architecture.md](./architecture.md), [domain-model.md](./domain-model.md), [data-flow.md](./data-flow.md), [compliance-mapping.md](./compliance-mapping.md).

---

## 1. Goal

Build rules evaluation and a tamper-evident audit ledger so that:

1. **Cedar Policy Service** evaluates 5 access/obligation policies and returns `RuleDecision`.
2. **Decision Service** evaluates 5 Zen decision models (Internal + COREP) with CI JSON fixtures.
3. **Audit Service** consumes `compliance.audit.pending`, writes **hash-chained** entries to **immudb**, and exposes verification + search APIs.
4. **All Phase 2 alerts** gain a non-null **`evidenceRef`** pointing to an immudb `Alert` entry.
5. **Audit Explorer UI** (Next.js) searches by rule, date, entity and shows integrity badges.
6. **CI blocks merge** on Cedar Analyzer failure or Zen fixture regression.

---

## 2. Scope boundaries

### In scope (Phase 3)

| Deliverable | Technology |
|-------------|------------|
| Cedar Policy Service | Go + [cedar-go](https://github.com/cedar-policy/cedar-go) |
| Decision Service | Go + GoRules Zen engine |
| Audit Service | Go + immudb gRPC client |
| immudb | Compose (single node) |
| Audit Explorer UI | Next.js 14+ (App Router) |
| Kafka topics | `compliance.audit.pending`, `compliance.audit.recorded` |
| Avro/JSON schemas | `RuleDecision`, audit envelope extensions |
| Policy/rule repos | `policies/cedar/`, `policies/zen/` |
| CI gates | Cedar Analyzer + Zen fixture tests |
| `smoke-test-phase3.sh` | End-to-end audit + evaluate + verify |

### Out of scope (defer to Phase 4+)

Per [AGENTS.md](../AGENTS.md) and [roadmap.md](./roadmap.md):

- Neo4j / Graph Service
- Simulation Service (Python)
- Regulatory reporting (XBRL/SDMX)
- Keycloak / full OIDC middleware (use mock principal per ADR-009 D20)
- S3 Object Lock in dev (filesystem artifact stub per D19)
- Flink hot-path Zen migration (Phase 3b stretch only)
- ClickHouse time-series store

### Rule set (Phase 3)

Five **Cedar** policies (access / obligation) and five **Zen** models (Internal + COREP), aligned with [compliance-mapping.md](./compliance-mapping.md):

| ID | Engine | Summary | Phase 3 trigger |
|----|--------|---------|-----------------|
| `INT-R003` | Cedar | Sensitive twin data requires `role=Analyst` or `role=Reporter` | Audit Explorer API probe |
| `INT-R004` | Cedar | Payment > €500K requires `role=Approver` | REST evaluate fixture |
| `COREP-R005` | Cedar | Capital adjustment requires `role=CapitalManager` | REST evaluate fixture |
| `EMIR-R004` | Cedar | Trade reporting data requires `role=TradeReporter` | REST evaluate fixture |
| `DORA-R001` | Cedar | Critical ICT contract change requires `role=ICTRiskManager` | REST evaluate fixture |
| `INT-R001` | Zen | Velocity > 50 payments / account / 1h | Mirrors `INT-M001` threshold |
| `INT-R002` | Zen | Counterparty exposure > €10M | Mirrors `INT-M002` threshold |
| `BASEL-R001` | Zen | LCR < 100% | Mirrors `BASEL-M001` threshold |
| `COREP-R001` | Zen | CET1 ratio below regulatory minimum | REST evaluate + fixture |
| `COREP-R002` | Zen | Total capital ratio below 8% | REST evaluate + fixture |

**Note**: Phase 2 Flink CEP (`INT-M001`, `INT-M002`, `BASEL-M001`) remains inline per ADR-008 D12. Phase 3 Zen models must produce **equivalent outcomes** on shared seed fixtures so migration is possible later.

---

## 3. Repository layout

Add to Phase 2 structure:

```
/
├── docker-compose.dev.yml          # extend: immudb, cedar-service, decision-service, audit-service, audit-explorer
├── schemas/avro/
│   ├── rule-decision.avsc
│   └── audit-entry.avsc
├── policies/
│   ├── cedar/
│   │   ├── schema.cedarschema
│   │   └── *.cedar                 # 5 policies
│   └── zen/
│       ├── int-r001.json           # JDM files
│       └── fixtures/               # input → expected RuleDecision
├── services/
│   ├── cedar-service/
│   ├── decision-service/
│   └── audit-service/
├── apps/
│   └── audit-explorer/             # Next.js
├── scripts/
│   ├── smoke-test-phase3.sh
│   ├── verify-audit-chain.sh
│   └── run-policy-ci.sh            # Cedar Analyzer + Zen tests
└── .github/workflows/
    ├── ci.yml                      # extend Phase 3 smoke
    └── policy-gates.yml            # Cedar + Zen on PR
```

**Go module paths**:

- `github.com/digital-twin/platform/services/cedar-service`
- `github.com/digital-twin/platform/services/decision-service`
- `github.com/digital-twin/platform/services/audit-service`

---

## 4. Docker Compose (local dev)

Extend `docker-compose.dev.yml`:

| Service | Image / build | Ports | Notes |
|---------|---------------|-------|-------|
| `immudb` | `codenotary/immudb:latest` | 3322 (gRPC), 9497 (metrics) | Init DB `digitaltwin_audit`; persist volume |
| `audit-service` | build `services/audit-service` | 8090 | Sole immudb writer |
| `cedar-service` | build `services/cedar-service` | 8091 | Mount `policies/cedar` read-only |
| `decision-service` | build `services/decision-service` | 8092 | Mount `policies/zen` read-only |
| `audit-explorer` | build `apps/audit-explorer` | 3001 | `ALERT_API` unchanged; add `AUDIT_API` |

Environment additions in `.env.example`:

```bash
# Phase 3
IMMUDB_HOST=immudb
IMMUDB_PORT=3322
IMMUDB_DATABASE=digitaltwin_audit
AUDIT_SERVICE_URL=http://localhost:8090
CEDAR_SERVICE_URL=http://localhost:8091
DECISION_SERVICE_URL=http://localhost:8092
AUDIT_EXPLORER_URL=http://localhost:3001
AUDIT_ARTIFACT_DIR=/data/audit-artifacts
KAFKA_AUDIT_PENDING_TOPIC=compliance.audit.pending
KAFKA_AUDIT_RECORDED_TOPIC=compliance.audit.recorded
```

**Health**: each Go service exposes `GET /api/v1/health` including immudb connectivity (audit-service) and policy load status (cedar/decision).

---

## 5. Event and schema contracts

### 5.1 Kafka topics

| Topic | Producer | Consumer | Payload |
|-------|----------|----------|---------|
| `compliance.audit.pending` | alert-service, cedar-service, decision-service | audit-service | `AuditPending` envelope |
| `compliance.audit.recorded` | audit-service | reporting (future), audit-explorer (optional) | `AuditEntry` summary |

Create topics in `scripts/create-kafka-topics.sh` (or `seed.sh`) before consumers start.

### 5.2 AuditPending envelope

Inner payload (JSON) written to Kafka; Audit Service maps to immudb row:

```json
{
  "entryType": "Alert",
  "correlationId": "smoke-phase3-001",
  "subject": { "subjectId": "uuid", "subjectType": "ComplianceAlert" },
  "actor": { "actorId": "alert-service", "actorType": "Service" },
  "action": "ComplianceAlertRaised",
  "payload": {
    "alertId": "uuid",
    "ruleCode": "INT-M001",
    "severity": "Warning",
    "summary": "Velocity threshold exceeded"
  },
  "metadata": {
    "regime": "Internal",
    "policyVersion": "phase2-inline",
    "sourceEventId": "uuid",
    "retentionUntil": "2033-06-15"
  }
}
```

### 5.3 RuleDecision (shared output)

```json
{
  "decisionId": "uuid",
  "ruleCode": "COREP-R001",
  "outcome": "Deny",
  "score": 0.92,
  "rationale": "CET1 ratio 6.8% below minimum 7.0%",
  "policyVersion": "1.0.0",
  "evaluatedAt": "2026-06-15T12:00:00.000Z",
  "inputHash": "sha256:..."
}
```

Add golden fixtures under `contracts/kafka/` for cross-service parsing (same pattern as Phase 2).

### 5.4 Hash chain

Implement per [data-flow.md](./data-flow.md) §5.3:

- `payloadHash = SHA-256(canonical JSON of payload + metadata)`
- `previousHash` = prior entry's `payloadHash` (empty string for genesis)
- `verify-audit-chain.sh` calls `GET /api/v1/audit/verify` and exits non-zero on break

---

## 6. Service specifications

### 6.1 Audit Service

**Responsibilities**:

1. Consume `compliance.audit.pending` (idempotent on `idempotencyKey`).
2. Write `AuditEntry` to immudb; emit `compliance.audit.recorded`.
3. REST:
   - `GET /api/v1/audit/entries?ruleCode=&from=&to=&subjectId=`
   - `GET /api/v1/audit/entries/{entryId}`
   - `GET /api/v1/audit/verify` — full chain or scoped range
4. Return `entryId` / `evidenceRef` to callers (sync path optional for smoke).

**Migrations**: PostgreSQL side table `audit_outbox_dlq` for failed immudb writes (mirror alert-service DLQ pattern).

**Tests**: unit tests for hash chain, consumer idempotency, verify API; testcontainers or Compose for immudb integration.

### 6.2 Cedar Policy Service

**Responsibilities**:

1. Load `policies/cedar/schema.cedarschema` + `*.cedar` at startup.
2. `POST /api/v1/evaluate` — body: `{ "principal", "action", "resource", "context" }` → `RuleDecision`.
3. On deny/flag: publish `RuleDecision` audit intent to `compliance.audit.pending`.
4. Dev auth: read `X-Principal`, `X-Roles` headers when principal omitted.

**CI**: `cedar analyze policies/cedar/` must exit 0 in `policy-gates.yml`.

### 6.3 Decision Service

**Responsibilities**:

1. Load Zen JDM from `policies/zen/*.json`.
2. `POST /api/v1/evaluate` — body: `{ "ruleCode", "input": { ... } }` → `RuleDecision`.
3. Publish audit intent on `Flag`, `Deny`, `Escalate`.
4. Expose `GET /api/v1/rules` listing loaded models + versions.

**CI**: table-driven tests — each file in `policies/zen/fixtures/` → expected outcome.

### 6.4 Alert Service changes (minimal)

1. After persisting a new alert, publish `AuditPending` (`entryType: Alert`) to `compliance.audit.pending`.
2. Subscribe to `compliance.audit.recorded` (or poll audit-service) to set `evidence_ref` on the alert row.
3. Include `evidenceRef` in API + WebSocket payloads.

Do **not** remove PostgreSQL alert storage (ADR-008 D11).

---

## 7. Audit Explorer UI

Next.js app at `apps/audit-explorer/`:

| Screen | Behavior |
|--------|----------|
| Search | Filter by `ruleCode`, date range, `subjectId`, `entryType` |
| Entry detail | Show payload, `payloadHash`, `previousHash`, integrity badge |
| Verify | Button runs client-side chain check display (server authoritative) |

Use `AUDIT_SERVICE_URL` server-side (Next.js route handlers). Reuse alert-console Tailwind/layout patterns.

---

## 8. Smoke test (`scripts/smoke-test-phase3.sh`)

Prerequisites: Phase 2 stack healthy (`smoke-test-phase2.sh` passes or `SMOKE_PHASE3_SKIP_PREREQS=1`).

| Step | Action | Pass condition |
|------|--------|----------------|
| 1 | Health | audit, cedar, decision, immudb healthy |
| 2 | Cedar evaluate | `INT-R003` deny without role → `Deny`; with `Reporter` → `Allow` |
| 3 | Zen evaluate | `BASEL-R001` with LCR 0.90 → `Deny`/`Flag` per fixture |
| 4 | Alert audit | Trigger or reuse open alert; `evidenceRef` non-null within 10s |
| 5 | Chain verify | `verify-audit-chain.sh` exits 0 |
| 6 | UI | Audit Explorer lists ≥1 entry from step 4 |

On failure, dump immudb entry count, last 3 `compliance.audit.pending` offsets, and alert `evidence_ref` column.

---

## 9. CI extensions

### 9.1 `policy-gates.yml` (new workflow)

```yaml
# On PR: policies/** changes
- cedar analyze policies/cedar/
- go test ./... in decision-service (Zen fixtures)
```

### 9.2 Extend `ci.yml`

After Phase 2 smoke:

1. `docker compose up -d --wait immudb audit-service cedar-service decision-service`
2. `./scripts/smoke-test-phase3.sh`

Unit job (fail fast):

```bash
cd services/audit-service && go test ./...
cd services/cedar-service && go test ./...
cd services/decision-service && go test ./...
./scripts/run-policy-ci.sh
```

---

## 10. Phase 2 → Phase 3 migration notes

| Phase 2 artifact | Phase 3 change |
|------------------|----------------|
| `evidenceRef` null on alerts | Backfill optional; new alerts must have ref |
| Inline Flink thresholds | Unchanged; Zen models mirror for CI parity |
| `compliance.alerts` topic | Unchanged; audit is sidecar via alert-service |
| Avro alert schemas | BACKWARD compatible; add optional `evidenceRef` enforcement in consumer |

---

## 11. Implementation order

1. ADR-009 + `.env.example` Phase 3 vars
2. Avro schemas + `contracts/kafka/` fixtures
3. `policies/cedar/` + `policies/zen/` seed policies/models
4. immudb in Compose + `verify-audit-chain.sh` stub
5. Audit Service (store, consumer, immudb writer, verify API) + tests
6. Cedar Service + Cedar Analyzer CI
7. Decision Service + Zen fixture CI
8. Alert Service `evidenceRef` + audit publish
9. Audit Explorer UI
10. `smoke-test-phase3.sh` + CI extension
11. `services/*/AGENTS.md` per service
12. Verify §13 exit criteria checklist

---

## 12. Testing rules

- Unit tests for every `services/{audit,cedar,decision}-service/internal/*` package.
- Integration tests may use testcontainers (immudb) or Compose (match CI).
- `./scripts/smoke-test.sh` and `./scripts/smoke-test-phase2.sh` must still pass (regression).
- Do not weaken Phase 2 tests to make Phase 3 green.

---

## 13. Phase 3 exit criteria checklist

Copy into PR description when Phase 3 is complete:

- [ ] Compose starts Phase 1–2 services **plus** immudb, audit, cedar, decision, audit-explorer
- [ ] 5 Cedar policies pass `cedar analyze` in CI
- [ ] 5 Zen models pass fixture tests in CI
- [ ] `compliance.audit.pending` → immudb → `compliance.audit.recorded` pipeline works
- [ ] Hash chain verification API returns valid for 100% of entries in smoke run
- [ ] New alerts have non-null `evidenceRef` within 10s
- [ ] Audit Explorer search returns alert audit entries with integrity badge
- [ ] `./scripts/smoke-test-phase3.sh` exits 0
- [ ] `./scripts/smoke-test-phase2.sh` still passes (regression)
- [ ] `go test ./...` passes for audit, cedar, decision services
- [ ] No Phase 4+ components added (Neo4j, simulation, XBRL)
- [ ] Keycloak not added (mock principal only per D20)

---

## 14. Staging notes

- immudb HA cluster + backup to S3 (ADR-003)
- Replace filesystem artifacts with S3 Object Lock
- OIDC (Keycloak) feeding Cedar principal attributes
- Optional: Flink → Decision Service HTTP for `BASEL-M001` (Phase 3b)

---

## References

- [roadmap.md](./roadmap.md) — Phase 3 duration and deliverables
- [ADR-009](./adr/009-phase3-foundation-decisions.md) — D15–D20
- [handoff-phase3-agent.md](./handoff-phase3-agent.md) — parallel implementation prompt
