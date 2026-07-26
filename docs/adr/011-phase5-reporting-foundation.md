# ADR-011: Phase 5 Reporting Foundation (D25–D29)

**Status**: Accepted (planning)  
**Date**: 2026-07-25  
**Deciders**: Platform Architecture Team  
**Implements**: CECT master-spec §2.1; site [phase5-implementation-spec.md](../phase5-implementation-spec.md) (authored with this ADR)  
**Supersedes (partial)**: ADR-009 D19 for *regulatory report artifacts* only (MinIO Object Lock required for Phase 5 exit; filesystem stub remains for non-report audit blobs in Phase 3 paths)

## Context

Phases 1–4 deliver CDC → twin state → Flink CEP → alerts → Cedar/Zen → immudb → Neo4j graph + deterministic stress. Supervisory **report generation** (XBRL / SDMX / DORA ICT XML), taxonomy versioning, immutable artifact storage, and a report lifecycle UI remain unfinished (roadmap Phase 5; CECT Tranche A).

| ID | Decision | Phase 5 impact |
|----|----------|----------------|
| D25 | Reporting Service runtime | Language / framework |
| D26 | Report artifact store | MinIO Object Lock vs filesystem / cloud S3 |
| D27 | Taxonomy + validation strategy | Full EBA download vs CI fixtures + arelle |
| D28 | Report Console lifecycle | States and enforcement |
| D29 | Taxonomy mapper persistence | Schema and pin semantics |

Open risks at handoff: **UR-CECT-001** (taxonomy packaging in CI).

## Decision

### D25 — Python FastAPI Reporting Service

**Decision**: Add **`services/reporting-service/`** as a **Python FastAPI** service that generates three report types from twin/seed state (not as system of record — CR-6):

1. **FINREP F01** — XBRL balance-sheet template  
2. **AnaCredit Table 2** — SDMX instrument data  
3. **DORA ICT Register** — XML

Inputs: State Service / seed mirrors and taxonomy maps. Outputs: validated artifacts + metadata (`taxonomyVersion`, `evidenceRef`, object-store key). Publish audit intents to **`compliance.audit.pending`** (Audit Service remains sole immudb writer — ADR-009 D16).

**Rationale**: ADR-006 polyglot split; Python ecosystem for XBRL/SDMX tooling (arelle); keeps Go services focused on hot-path twin/compliance.

### D26 — MinIO with Object Lock in Compose for report artifacts

**Decision**: Store submitted (and smoke-validated) regulatory reports in **MinIO** with **Object Lock / retention** configured for a **7-year** retention policy in dev/CI (S3-compatible API). Each stored report MUST have a non-null **`evidenceRef`** verifiable via Audit Service.

**Rationale**: Closes ADR-009 D19 upgrade path for report-sized artifacts without AWS dependency; satisfies AC-A-P5-003 / CR-4 evidence bar.

### D27 — Fixture-equivalent taxonomies + arelle (CI-honest)

**Decision**: Ship **versioned taxonomy fixtures** (published subset or golden instances) under the repo (e.g. `fixtures/taxonomies/` or `mocks/reporting/`). CI and `./scripts/smoke-test-phase5.sh` validate FINREP F01 with **arelle** (or documented fixture-equivalent). AnaCredit T2 and DORA ICT validate against **schema fixtures**. Full commercial EBA/ECB taxonomy parity is **out of scope** (CR-5 / out-of-scope list).

**Rationale**: Resolves UR-CECT-001 without blocking CI; honesty over false “full taxonomy” claims.

### D28 — Report Console: draft → validated → submitted

**Decision**: Add **`apps/report-console/`** (Next.js) with server-side proxies to Reporting Service (same CORS pattern as alert-console). Enforce lifecycle **`draft` → `validated` → `submitted`**; illegal transitions return 4xx. UI may be exercised by smoke API flow if Playwright is not yet wired.

**Rationale**: Matches roadmap Phase 5 exit and AC-A-P5-004.

### D29 — Versioned taxonomy mapper in PostgreSQL

**Decision**: Persist instrument/entity → EBA/ECB (and DORA ICT) code maps in **PostgreSQL** with **`version`**, **`effective_from` / `effective_to`**, and require generated reports to **pin** `taxonomyVersion`. Seed maps for the three report types in Phase 5 smoke.

**Rationale**: AC-A-P5-005 / R-D-001; supports annual taxonomy churn (roadmap R4).

### Stretch (not Phase 5 exit)

**ClickHouse** (or heavy Postgres pre-agg) for report input metrics is **optional stretch**. Prefer Postgres views / twin queries for smoke-green exit unless aggregation becomes a proven bottleneck.

## Consequences

### Positive

- Completes original reporting thesis with falsifiable `smoke-test-phase5.sh`.
- Object Lock + `evidenceRef` ties reports into the existing audit chain.
- Fixture strategy keeps CI hermetic and claim-honest.

### Negative

- New Python service + MinIO increase Compose footprint and CI time.
- Fixture taxonomies are not commercial suite parity — must stay explicit in docs/README claims.
- Report generation from twin/seed can diverge from live core banking; SoR remains mock CDC (CR-6).

## Alternatives Considered

| Decision | Alternative | Why rejected for Phase 5 |
|----------|-------------|--------------------------|
| D25 | Extend State Service (Go) for XBRL | Poor fit for arelle/XBRL ecosystem; couples hot path to report batch work |
| D26 | Filesystem-only artifacts | Fails Object Lock / retention acceptance |
| D26 | Real AWS S3 in CI | External dependency; credentials burden |
| D27 | Download full EBA taxonomy in CI | Fragile, large, often impractical (UR-CECT-001) |
| D28 | API-only without UI | Weakens lifecycle demo; roadmap expects Report UI |
| D29 | Hard-coded maps in Python | No version pin / effective dating |

## Exit wiring

- Script: `./scripts/smoke-test-phase5.sh` (exit 0)  
- Gates: `G-CECT-ADR-011`, `G-CECT-P5-SMOKE`  
- Acceptance: `AC-A-P5-001` … `AC-A-P5-005`

## References

- CECT `master-spec.md` §2.1, §0 CR-4/CR-5/CR-6  
- [ADR-003](./003-immudb-audit-ledger.md), [ADR-006](./006-polyglot-language-strategy.md), [ADR-009](./009-phase3-foundation-decisions.md)  
- [roadmap.md](../roadmap.md) — Phase 5
