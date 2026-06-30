# Phase 4 Readiness Checklist

Gate before opening Phase 4 implementation work. Maps [roadmap.md § Phase 4](../roadmap.md#phase-4-graph-model-and-simulation).

**Status**: Ready for implementation (2026-06-29)

---

## Prerequisites (must be true)

- [x] Phase 3 mechanical exit criteria satisfied — [phase3-exit-checklist.md](./phase3-exit-checklist.md)
- [x] Release **v0.1.0** tagged on `main` with CHANGELOG and Flink 1.20 runtime alignment
- [x] Phase 3b complete — Flink CEP → Decision Service for INT-M001, INT-M002, BASEL-M001 when `CEP_DECISION_SERVICE_URL` set
- [x] GHCR images published for Phase 1–3 services ([docs/deployment.md](../deployment.md))
- [x] [ADR-010](../adr/010-phase4-foundation-decisions.md) accepted (D21–D24)
- [x] [phase4-implementation-spec.md](../phase4-implementation-spec.md) authored
- [x] [handoff-phase4-agent.md](../handoff-phase4-agent.md) authored

---

## Planning artifacts

| Artifact | Purpose |
|----------|---------|
| ADR-010 | Neo4j, ingestion, simulation scope, audit linkage |
| phase4-implementation-spec.md | Executable layout, APIs, smoke, exit criteria |
| handoff-phase4-agent.md | Fresh-chat implementation prompt |
| domain-model.md § Exposure | Entity/relationship definitions (read-only reference) |

---

## Not blocking Phase 4 start (track separately)

| Item | Owner / when |
|------|----------------|
| Branch protection on `main` | Repo admin — [docs/github-setup.md](../github-setup.md) |
| Production TLS / OIDC | Phase 6+ hardening |
| Soak metrics (Flink checkpoint >99%, alert p99 <2s) | Ops benchmark; `./scripts/measure-phase3-latency.sh` samples rule eval only |
| Agent-based contagion | Phase 6+ per ADR-010 D23 |

---

## First implementation PR should include

1. Neo4j service in `docker-compose.dev.yml` + `.env.example` keys
2. `services/graph-service/` skeleton with health + Kafka consumer stub
3. `services/simulation-service/` skeleton with health
4. Placeholder `scripts/smoke-test-phase4.sh` (exit 1 until wired) — optional if spec prefers all-in-one PR

---

## Verification command (after Phase 4 claimed done)

```bash
docker compose -f docker-compose.dev.yml up -d --wait
./scripts/seed.sh
# ... Phase 1–3 warm steps per AGENTS.md ...
./scripts/smoke-test-phase4.sh
```

Regression floor: Phase 1–3 smoke suites remain green in CI.
