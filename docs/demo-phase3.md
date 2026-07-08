# Phase 3 local demo runbook

Executable walkthrough for showing **rules evaluation** (Cedar + Zen) and the **tamper-evident audit ledger** (immudb) on a laptop.

**Prerequisites**: Docker, `curl`, `jq`, `psql`. Phase 2 stack healthy (Kafka, Flink CEP, alert-service).

---

## Port map

| Service | URL | Demo use |
|---------|-----|----------|
| Alert Console | http://localhost:3000 | Compliance alerts; audit evidence link on detail |
| Audit Explorer | http://localhost:3002 | Search ledger, verify chain, entry detail |
| Alert API (scripts/curl) | http://localhost:8085 | REST; UI proxies via `/api/alerts` |
| Audit Service | http://localhost:8090 | Hash-chained entries, verify API |
| Cedar Service | http://localhost:8091 | Access / obligation policies |
| Decision Service | http://localhost:8092 | Zen regulatory models |
| Grafana (optional) | http://localhost:3030 | Metrics dashboards |

---

## Before the room (30–40 min cold, ~5 min if stack warm)

```bash
cd "/path/to/Digital Twin"
cp .env.example .env   # first time only

docker compose -f docker-compose.dev.yml up -d --wait
./scripts/seed.sh
./scripts/register-debezium-connector.sh
docker compose -f docker-compose.dev.yml restart state-service alert-service
./scripts/wait-outbox-drained.sh
./scripts/submit-flink-job.sh
```

**5 minutes before showtime:**

```bash
./scripts/demo-phase3.sh --restart-policies
```

`--restart-policies` avoids empty Cedar/Zen volume mounts (common when the repo path contains spaces).

Confirm automated proof (optional but recommended):

```bash
SMOKE_PHASE3_SKIP_PREREQS=1 ./scripts/smoke-test-phase3.sh
```

---

## Demo helper script

```bash
chmod +x scripts/demo-phase3.sh

# Health + sample policy evals + chain verify + URLs
./scripts/demo-phase3.sh

# Include a live INT-M001 alert moment (~30–60s)
./scripts/demo-phase3.sh --trigger-alert --restart-policies
```

---

## Suggested narrative (~15 minutes)

### 1. Phase 2 context (2 min)

Open **Alert Console** (http://localhost:3000). Explain that Flink CEP already fires `INT-M001`, `INT-M002`, `BASEL-M001` from twin state.

### 2. Phase 3 audit path (4 min)

Open an alert detail page. Show **Audit evidence** with `evidenceRef` and **View in Audit Explorer →**.

In **Audit Explorer** (http://localhost:3002):

- List entries (Alert + RuleDecision types)
- Click **Verify chain** (green = hash chain intact)
- Open an entry: `payloadHash`, `previousHash`, payload JSON

**Story**: alert persisted → Kafka `compliance.audit.pending` → Audit Service → immudb → `evidenceRef` on alert row.

### 3. Live policy engines (4 min)

Terminal — Cedar deny/allow:

```bash
# Deny: sensitive twin data, no role
curl -s -X POST http://localhost:8091/api/v1/evaluate \
  -H 'Content-Type: application/json' \
  -d '{"ruleCode":"INT-R003","principal":{"id":"demo","roles":[]},"resource":{"type":"TwinData","id":"t1","attributes":{"sensitivity":"high"}}}' | jq .

# Allow: same request with Reporter role (mock principal — no Keycloak in Phase 3)
curl -s -X POST http://localhost:8091/api/v1/evaluate \
  -H 'Content-Type: application/json' -H 'X-Roles: Reporter' \
  -d '{"ruleCode":"INT-R003","principal":{"id":"demo"},"resource":{"type":"TwinData","id":"t1","attributes":{"sensitivity":"high"}}}' | jq .
```

Zen Basel LCR:

```bash
curl -s -X POST http://localhost:8092/api/v1/evaluate \
  -H 'Content-Type: application/json' \
  -d '{"ruleCode":"BASEL-R001","input":{"lcr":0.9,"personaId":"44444444-4444-4444-4444-444444444401"}}' | jq .
```

Refresh Audit Explorer — new `RuleDecision` entries may appear.

### 4. Live alert moment (3 min, optional)

```bash
./scripts/demo-phase3.sh --trigger-alert
```

Or manually:

```bash
./mocks/simulators/payment-burst.sh
```

Watch Alert Console for new `INT-M001`; within ~30s the detail page should show `evidenceRef` and the Explorer link.

### 5. Integrity proof (2 min)

```bash
./scripts/verify-audit-chain.sh
```

Or use **Verify chain** in Audit Explorer.

---

## What to say is out of scope

- Keycloak / full OIDC (mock `X-Roles` only)
- Flink **INT-M001**, **INT-M002**, and **BASEL-M001** call Decision Service on the hot path when `CEP_DECISION_SERVICE_URL` is set (Phase 3b); inline thresholds on HTTP failure
- Neo4j, simulation, XBRL (Phase 4+)
- S3 Object Lock artifacts (filesystem stub in dev)

---

## Troubleshooting

| Symptom | Fix |
|---------|-----|
| Cedar/Zen `no such file or directory` for policies | `./scripts/demo-phase3.sh --restart-policies` |
| No alerts / empty Explorer | Re-run seed + Flink submit; check `docker compose ps` |
| `evidenceRef` pending on alert | Wait 10–30s; check audit-service logs and Kafka topic |
| Cold start too slow | Pre-warm stack; do not `docker compose up` live in the demo |
| Full regression | `./scripts/smoke-test-phase3.sh` |

---

## Screenshots for maintainers

Regenerate [README](../README.md) hero images when alert-console or audit-explorer UI changes materially.

**Prerequisites:** warm Phase 3 stack with a linked alert (`evidenceRef` populated).

```bash
./scripts/demo-phase3.sh --trigger-alert --restart-policies
```

Note the printed URLs (alert detail + audit entry), then capture at ~1280px width:

| Output file | URL pattern |
|-------------|-------------|
| `docs/assets/alert-console.png` | `http://localhost:3000/alerts/{alertId}` — INT-M001 detail with **View in Audit Explorer** link |
| `docs/assets/audit-explorer.png` | `http://localhost:3002/entries/{entryId}` — entry with **integrity OK** badge and payload JSON |

Optional: export `docs/assets/social-preview.png` from [social-preview.svg](./assets/social-preview.svg) (1280×640) for GitHub Settings → Social preview.

**Demo GIF:** after capturing PNGs above, run `npm run screenshots` (live stack) or `npm run screenshots:rebuild-gif` (from existing PNGs). Output: `docs/assets/demo.gif` (2s per frame, same timing as PSA).

---

## References

- [phase3-implementation-spec.md](./phase3-implementation-spec.md) — scope and exit criteria
- [AGENTS.md](../AGENTS.md) — stack commands and gotchas
