# Phase 4 local demo runbook

Warm-stack walkthrough for **Graph & simulation · Phase 4**: Neo4j exposure graph + deterministic stress simulation.

**Prerequisites**: Docker, `curl`, `jq`. Phases 1–3 stack healthy (Compose up, seed, Debezium, Flink optional for this demo).

---

## Port map

| Service | URL | Demo use |
|---------|-----|----------|
| Graph Explorer | http://localhost:3003 | Institution exposure network |
| Simulation Console | http://localhost:3004 | ECB Adverse stress run |
| Graph Service | http://localhost:8093 | REST summary / nodes / edges |
| Simulation Service | http://localhost:8094 | `POST /api/v1/simulations/run` |
| Neo4j Browser | http://localhost:7474 | Optional graph inspection |
| Audit Explorer | http://localhost:3002 | Stress decisions / evidence links |

Use the shared console app switcher to move between UIs.

---

## Before the room (~5 min if stack warm)

```bash
cd "/path/to/Digital Twin"
docker compose -f docker-compose.dev.yml up -d --wait
./scripts/seed.sh
./scripts/wait-graph-seeded.sh   # if present; else wait for graph-service healthy
SMOKE_PHASE4_SKIP_PREREQS=1 ./scripts/smoke-test-phase4.sh
```

Cold start: follow [README Quick start](../README.md#quick-start) steps 1–3 first, then the commands above.

---

## Suggested narrative (~10 minutes)

### 1. Exposure graph (4 min)

Open **Graph Explorer** (http://localhost:3003). Show node/edge counts, filter by layer, and the SVG network. Explain twin institutions mirrored into Neo4j.

### 2. Stress simulation (4 min)

Open **Simulation Console** (http://localhost:3004). Run **ECB Adverse v1**. Point at baseline vs stressed CET1 and policy decisions. Follow **Open Audit Explorer →** when evidence is linked.

### 3. Prove with smoke (2 min)

```bash
SMOKE_PHASE4_SKIP_PREREQS=1 ./scripts/smoke-test-phase4.sh
```

---

## Related

- Journey map: [phase-journeys.md](phase-journeys.md)
- ADR: [010-phase4-foundation-decisions.md](adr/010-phase4-foundation-decisions.md) (if present) · Phase 4 spec
