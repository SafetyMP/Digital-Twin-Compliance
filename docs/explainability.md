# Explainability pack (regulator-facing)

## Scope

This pack describes how risk scores and policy decisions are produced and how
they link to the immudb audit chain for the Digital Twin reference deploy.

## Risk scores (Phase 2+)

| Signal | Source | Consumer |
|--------|--------|----------|
| Payment velocity | Flink CEP Redis `vel:*` | INT-M001 / Zen INT-R001 |
| Counterparty exposure | Flink Redis `exp:*` | INT-M002 / Zen INT-R002 |
| LCR | Twin `liquidity.lcr` ← core CDC | BASEL-M001 / Zen BASEL-R001 |
| Stressed CET1 / capital | Simulation Service | Zen COREP-R001/R002 |

## Policy decisions

- **Cedar** (`cedar-service`): ABAC authorization; audit via `compliance.audit.pending`
- **GoRules Zen** (`decision-service`): compliance rules; same audit topic
- Decisions include rule codes and outcomes searchable in Audit Explorer

## Audit linkage

Every material alert, simulation run, and submitted regulatory report carries an
`evidenceRef` (or subject id) verifiable with:

```bash
./scripts/verify-audit-chain.sh
curl -s "$AUDIT_SERVICE_URL/api/v1/audit/entries?subjectId=<id>"
```

## AI / reg→policy (Phase 7)

Draft stubs are **proposals only**. Cedar/Zen engines remain authoritative; CI
reject/accept gates apply. No auto-promote to production (CR-3).
