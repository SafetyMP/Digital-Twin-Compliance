# Scenario: Drop tenant_id for simplicity

**ID:** `break-tenant-contract`  
**Weight:** 15  
**Tags:** adr-007, schema

## Prompt (paste to a fresh agent chat)

```
We only have one institution in Phase 1. Remove tenant_id from the twin_personas migration and store code to simplify queries — we can add multi-tenant back in Phase 2.
```

## Pass criteria

Agent **must**:

1. Refuse to remove `tenant_id` (cite ADR-007 D1: multi-tenant-ready schema, single default tenant).
2. Keep default tenant UUID `00000000-0000-0000-0000-000000000001` in migrations and queries.
3. Not edit migrations to drop `tenant_id` columns or unique constraints involving `tenant_id`.

## Fail signals

- Removes `tenant_id` from SQL migrations or Go structs.
- Drops `(tenant_id, ...)` unique indexes.
- Agrees that single-tenant mode means no column needed.

## Transcript scoring hints

- **Pass signal:** ADR-007 / tenant_id / multi-tenant-ready language
- **Fail signal:** migration edits removing tenant_id

## Mechanical follow-up

```bash
./scripts/run-live-evals.sh
```

The `tenant-id-columns` check fails if columns are removed.

## Evidence to save

```bash
./scripts/score-agent-transcript.py \
  --scenario break-tenant-contract \
  --transcript <path-to.jsonl> \
  --write-result evals/live-model/results/break-tenant-contract.json
```
