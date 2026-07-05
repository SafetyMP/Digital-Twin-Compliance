# Simulation Service — Agent Contract

Python FastAPI service for deterministic stress simulation (Phase 4).

Parent contract: [AGENTS.md](../../AGENTS.md). Audit envelope: [docs/phase4-implementation-spec.md](../../docs/phase4-implementation-spec.md) §5.3.

## Layout

| Path | Role |
|------|------|
| `simulation_service/main.py` | FastAPI app + run endpoint |
| `simulation_service/scenario.py` | ECB Adverse deterministic scenario (NetworkX) |
| `simulation_service/graph_client.py` | HTTP client to graph-service |
| `simulation_service/decision.py` | COREP-R001/R002 via decision-service |
| `simulation_service/audit.py` | `compliance.audit.pending` publisher |

## Commands

```bash
cd services/simulation-service
pip install -r requirements.txt
pytest -q
python -m simulation_service.main
```

**Verification floor:** any edit requires `pytest -q` before claiming done.

## API

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/health` | Liveness + graph-service reachability |
| POST | `/api/v1/simulations/run` | Run scenario; audit + Zen evaluation |

## Invariants

- One exit scenario: `ecb-adverse-v1` (ADR-010 D23)
- Audit via `compliance.audit.pending` only (ADR-010 D24)
- Browser UIs proxy via Next.js — no CORS on `:8094`

## Gotchas

- Graph must be seeded before run — use `./scripts/wait-graph-seeded.sh`
- Stressed CET1 floored at 0.068 for deterministic COREP-R001 Deny in smoke
- `KAFKA_BROKERS` comma-separated for audit publish
