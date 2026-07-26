# Phase 5 local demo runbook

Warm-stack walkthrough for **Regulatory reporting · Phase 5**: FINREP / AnaCredit / DORA draft → validate → submit with MinIO Object Lock.

**Prerequisites**: Docker, `curl`, `jq`. Dev Compose with `reporting-service`, MinIO, and Report Console healthy.

---

## Port map

| Service | URL | Demo use |
|---------|-----|----------|
| Report Console | http://localhost:3005 | Lifecycle UI |
| Reporting Service | http://localhost:8095 | REST report API |
| MinIO Console | http://localhost:9001 | Object Lock artifacts (optional) |
| Audit Explorer | http://localhost:3002 | `evidenceRef` after submit |

Use the shared console app switcher to jump from Alert/Audit into Report Console.

---

## Before the room (~5 min if stack warm)

```bash
cd "/path/to/Digital Twin"
docker compose -f docker-compose.dev.yml up -d --wait reporting-service report-console minio
SMOKE_PHASE5_SKIP_PREREQS=1 ./scripts/smoke-test-phase5.sh
```

If reporting services were never brought up, use full `docker compose -f docker-compose.dev.yml up -d --wait` then re-run smoke.

---

## Suggested narrative (~10 minutes)

### 1. Create draft (3 min)

Open **Report Console** (http://localhost:3005). Choose **FINREP F01 (XBRL)** (or AnaCredit / DORA). Click **Create draft** — show `id` + `status=draft`.

### 2. Validate → submit (4 min)

**Validate** then **Submit**. Show `taxonomyVersion`, `objectKey`, and `evidenceRef`. Optionally open MinIO and Audit Explorer for the locked object / ledger entry.

### 3. Prove with smoke (2 min)

```bash
SMOKE_PHASE5_SKIP_PREREQS=1 ./scripts/smoke-test-phase5.sh
```

Honest scope: fixture taxonomies for the reference stack — not commercial EBA XBRL suite parity. See [ADR-011](adr/011-phase5-reporting-foundation.md).

---

## Related

- Journey map: [phase-journeys.md](phase-journeys.md)
- [ADR-011](adr/011-phase5-reporting-foundation.md)
