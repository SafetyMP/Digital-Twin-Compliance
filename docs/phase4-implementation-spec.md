# Phase 4 Implementation Spec

Executable handoff for the implementation agent. Implements [roadmap.md](./roadmap.md) Phase 4 only.

**Prerequisites**: Phase 3 complete ([review/phase3-exit-checklist.md](./review/phase3-exit-checklist.md), release **v0.1.0+**). [ADR-010](./adr/010-phase4-foundation-decisions.md) (D21–D24).

**Related docs**: [architecture.md](./architecture.md), [domain-model.md](./domain-model.md), [data-flow.md](./data-flow.md), [handoff-phase4-agent.md](./handoff-phase4-agent.md).

---

## 1. Goal

Build graph analytics and deterministic stress simulation so that:

1. **Graph Service** (Go + Neo4j) materializes institution and exposure edges from streaming twin/domain events.
2. **Graph UI** visualizes the exposure network with layer filtering (short-term / long-term / contingent).
3. **Simulation Service** (Python) runs one **ECB Adverse–style** deterministic scenario on the seed graph and returns explainable metrics.
4. **Simulation UI** configures scenario parameters and compares baseline vs stressed outcomes.
5. **Zen COREP-R001/R002** evaluate simulation output; results link to immudb via **`evidenceRef`-style audit entries**.
6. **CI + smoke** block merge on graph ingestion regression and simulation fixture drift.

---

## 2. Scope boundaries

### In scope (Phase 4)

| Deliverable | Technology |
|-------------|------------|
| Graph Service | Go + Neo4j driver (Bolt) |
| Neo4j | Compose single instance |
| Graph UI | Next.js + sigma.js or react-flow |
| Simulation Service | Python 3.11+ + NetworkX |
| Simulation UI | Next.js (parameter form + results) |
| Kafka consumption | `twin.state.updated`, selected `domain.events.public.*` |
| gRPC/HTTP API | Graph query + simulation run |
| `smoke-test-phase4.sh` | Graph node/edge counts + simulation run + audit linkage |
| CI extend | Phase 4 smoke + Python/Go unit tests |

### Out of scope (defer to Phase 5+)

Per [AGENTS.md](../AGENTS.md):

- Regulatory reporting (XBRL/SDMX/ClickHouse)
- Agent-based contagion simulation (Phase 6+)
- Keycloak / full OIDC
- Neo4j Aura / HA cluster (document staging path only)
- Contract NLP / unstructured obligation parsing

### Rule set (Phase 4)

Reuse Phase 3 Zen models — no new Cedar policies required for exit criteria:

| ID | Trigger |
|----|---------|
| `COREP-R001` | CET1 ratio from simulation output below minimum |
| `COREP-R002` | Total capital ratio below 8% |

---

## 3. Repository layout

Add to Phase 3 structure:

```
/
├── docker-compose.dev.yml          # extend: neo4j, graph-service, simulation-service, graph-ui, simulation-ui
├── schemas/avro/
│   └── simulation-run.avsc         # optional; JSON envelope OK in dev per ADR-007
├── services/
│   ├── graph-service/              # Go
│   └── simulation-service/         # Python (FastAPI or gRPC)
├── apps/
│   ├── graph-explorer/             # Next.js
│   └── simulation-console/         # Next.js
├── scripts/
│   ├── smoke-test-phase4.sh
│   └── wait-graph-seeded.sh        # poll Neo4j counts like wait-outbox-drained.sh
└── .github/workflows/
    └── ci.yml                      # extend Phase 4 smoke + pytest
```

**Module paths**:

- `github.com/digital-twin/platform/services/graph-service`
- Python package `simulation_service` under `services/simulation-service/`

Each new service gets `services/<svc>/AGENTS.md` during implementation.

---

## 4. Docker Compose (local dev)

Extend `docker-compose.dev.yml`:

| Service | Image / build | Ports | Notes |
|---------|---------------|-------|-------|
| `neo4j` | `neo4j:5-community` | 7474 (browser), 7687 (Bolt) | Auth from env; persist volume |
| `graph-service` | build `services/graph-service` | 8093 | Consumes Kafka; writes Neo4j |
| `simulation-service` | build `services/simulation-service` | 8094 | Calls graph-service HTTP; publishes audit |
| `graph-explorer` | build `apps/graph-explorer` | 3003 | Proxy to graph-service |
| `simulation-console` | build `apps/simulation-console` | 3004 | Proxy to simulation-service |

Environment additions in `.env.example`:

```bash
# Phase 4
NEO4J_URI=bolt://neo4j:7687
NEO4J_USER=neo4j
NEO4J_PASSWORD=changeme
GRAPH_SERVICE_URL=http://localhost:8093
SIMULATION_SERVICE_URL=http://localhost:8094
GRAPH_EXPLORER_URL=http://localhost:3003
SIMULATION_CONSOLE_URL=http://localhost:3004
```

**Health**: Go/Python services expose `GET /api/v1/health` including Neo4j connectivity (graph-service) and graph-service reachability (simulation-service).

**UI proxy rule** (same as alert-console / audit-explorer): browser must not call `:8093`/`:8094` directly — use Next.js `/api/*` routes.

---

## 5. Event and graph model

### 5.1 Ingestion topics

| Topic | Consumer | Action |
|-------|----------|--------|
| `twin.state.updated` | graph-service | Upsert `LegalEntity` / `Instrument` nodes; refresh exposure properties |
| `domain.events.public.instruments` (or CDC-derived) | graph-service | Create/update `Exposure` edges |

### 5.2 Neo4j schema (minimal)

**Nodes**: `LegalEntity {tenantId, entityId, name, lcr, cet1Ratio, …}`  
**Relationships**: `EXPOSURE {exposureType, notionalEur, layer, updatedAt}`

Seed target for smoke: **≥ 10 institutions**, **≥ 50 exposure edges** (roadmap exit criteria).

### 5.3 SimulationRun audit envelope

Publish to `compliance.audit.pending`:

```json
{
  "entryType": "SimulationRun",
  "correlationId": "smoke-phase4-001",
  "subject": { "subjectId": "run-uuid", "subjectType": "SimulationRun" },
  "actor": { "actorId": "simulation-service", "actorType": "Service" },
  "action": "SimulationRunCompleted",
  "payload": {
    "scenarioId": "ecb-adverse-v1",
    "baselineCet1": 0.12,
    "stressedCet1": 0.068,
    "explainabilityRef": "graph-path:entity-a→entity-b"
  }
}
```

---

## 6. Service specifications

### 6.1 Graph Service

- Kafka consumer group `graph-service`
- Idempotent upserts keyed by `tenantId + entityId` / edge composite key
- REST: `GET /api/v1/graph/summary`, `GET /api/v1/graph/nodes`, `GET /api/v1/graph/edges?layer=`
- Optional Cypher passthrough **disabled** in dev (security); fixed query templates only

### 6.2 Simulation Service

- `POST /api/v1/simulations/run` with `{ "scenarioId": "ecb-adverse-v1", "parameters": {…} }`
- Loads subgraph via graph-service; runs deterministic stress propagation
- Calls Decision Service for COREP-R001/R002 on stressed metrics
- Publishes audit pending event; returns run id + rule decisions

### 6.3 UIs

| App | Minimum UX |
|-----|------------|
| graph-explorer | Force-directed graph, institution search, layer toggle |
| simulation-console | Scenario picker, run button, baseline vs stressed table, link to audit entry |

---

## 7. Smoke test (`smoke-test-phase4.sh`)

Steps (after Phase 1–3 prereqs):

1. Health: neo4j, graph-service, simulation-service, both UIs
2. Wait `./scripts/wait-graph-seeded.sh` — node count ≥ 10, edges ≥ 50
3. Graph API returns seed institution by name
4. Run simulation scenario — completes in < 60s
5. Audit Service verify — new `SimulationRun` entry with valid chain
6. Regression: `./scripts/smoke-test-phase3.sh` (or `SMOKE_PHASE3_SKIP_PREREQS=1` when warm)

---

## 8. CI extension

In `.github/workflows/ci.yml` after Phase 3 smoke:

```bash
./scripts/smoke-test-phase4.sh
cd services/graph-service && go test ./...
cd services/simulation-service && pytest -q
```

Add `bash -n scripts/smoke-test-phase4.sh scripts/wait-graph-seeded.sh`.

Policy gates unchanged unless simulation adds new Zen fixture files.

---

## 9. Implementation order

1. Neo4j Compose + health
2. graph-service consumer + REST
3. `wait-graph-seeded.sh` + graph API tests
4. simulation-service scenario + audit publish
5. graph-explorer + simulation-console (proxy routes)
6. `smoke-test-phase4.sh` + CI
7. `services/*/AGENTS.md` + exit checklist

---

## 10. Phase 4 exit criteria checklist

Copy into PR when Phase 4 is complete:

- [ ] Compose starts Phase 1–3 services **plus** neo4j, graph-service, simulation-service, graph-explorer, simulation-console
- [ ] Graph contains ≥ 10 institutions and ≥ 50 exposure edges from seed/streaming ingestion
- [ ] Graph UI renders interactive exposure graph with layer filter
- [ ] Deterministic stress scenario completes in < 60s on seed dataset
- [ ] Simulation results produce audit entry searchable in Audit Explorer
- [ ] COREP-R001/R002 evaluate against simulation output in smoke run
- [ ] `./scripts/smoke-test-phase4.sh` exits 0
- [ ] `./scripts/smoke-test-phase3.sh` still passes (regression)
- [ ] `go test ./...` (graph-service) and `pytest` (simulation-service) pass
- [ ] No Phase 5+ components (XBRL, ClickHouse reporting pipeline)

---

## 11. Staging notes

- Neo4j backup + Aura migration path
- Simulation horizontal scale (job queue) if runtime exceeds 60s on production graph size
- gRPC between simulation ↔ graph if HTTP payload size becomes limiting

---

## References

- [roadmap.md](./roadmap.md) — Phase 4 duration and deliverables
- [ADR-010](./adr/010-phase4-foundation-decisions.md) — D21–D24
- [review/phase4-readiness.md](./review/phase4-readiness.md) — pre-implementation gate
- [handoff-phase4-agent.md](./handoff-phase4-agent.md) — agent prompt
