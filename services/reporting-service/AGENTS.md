# Reporting Service — Agent Contract

Python FastAPI service for supervisory report generation (Phase 5 / ADR-011).

Parent: [AGENTS.md](../../AGENTS.md) · Spec: [phase5-implementation-spec.md](../../docs/phase5-implementation-spec.md)

## Layout

| Path | Role |
|------|------|
| `reporting_service/main.py` | FastAPI app |
| `reporting_service/generators.py` | FINREP / AnaCredit / DORA generators |
| `reporting_service/taxonomy.py` | Versioned mapper (Postgres) |
| `reporting_service/storage.py` | MinIO Object Lock put/get |
| `reporting_service/audit.py` | `compliance.audit.pending` publisher |
| `reporting_service/validate.py` | Fixture-equivalent validators |

## Commands

```bash
cd services/reporting-service
pip install -r requirements.txt
pytest -q
```

## API

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/health` | Liveness + DB/MinIO |
| GET | `/api/v1/taxonomies` | Versioned taxonomy rows |
| POST | `/api/v1/reports` | Create draft report |
| POST | `/api/v1/reports/{id}/validate` | draft → validated |
| POST | `/api/v1/reports/{id}/submit` | validated → submitted (MinIO + audit) |
| GET | `/api/v1/reports/{id}` | Fetch report metadata |

## Invariants

- Twin/seed is intelligence input; mock core banking remains SoR (CR-6)
- Audit via `compliance.audit.pending` only
- Submitted reports MUST have object key + non-null `evidenceRef`
- Fixture taxonomies — no commercial suite parity claims
