# ADR 0000: Threat Model — Caller, Trust Boundary, Authentication

**Status:** Accepted  
**Date:** 2026-07-13  
**Product:** Digital Twin

## Context

Digital Twin runs a multi-service compose stack. Cooperative `./scripts/smoke-test*.sh` validates CDC, personas, and streaming paths. Services that enforce **verified principals** must reject anonymous callers in tier-3 adversarial CI — separate from cooperative smoke.

Machine-readable cells: `specs/threat-model.yaml`. Tier-3 negatives: `scripts/adversarial.sh`.

## Decision

### Principals

| Principal | Services | Notes |
|-----------|----------|-------|
| `anonymous` | cedar-service | denied on policy evaluate |
| `service_jwt` | cedar-service | bearer JWT with `sub` + roles |
| `dev_open` | state-service (dev) | mock principal — production auth deferred (ADR-009) |

### Trust boundaries

| Boundary | Route | Authentication | Failure |
|----------|-------|------------------|---------|
| Cedar policy evaluate | `POST /api/v1/evaluate` | `Authorization: Bearer` JWT (`CEDAR_SERVICE_JWT_SECRET`) | `401` missing/invalid token |
| State personas (dev) | `GET /api/v1/personas` | none in dev stack | documented open — adversarial does not assert deny |

### Authentication mechanism (cedar-service)

JWT HS256 bearer verified in `services/cedar-service/internal/auth/jwt.go`. Body/header role claims are ignored once principal is resolved from token.

## Consequences

**Positive:** Cedar auth regressions fail adversarial CI without polluting cooperative smoke.

**Negative:** State-service production auth still TODO; threat model documents current dev posture honestly.

## References

- `specs/threat-model.yaml`, `scripts/adversarial.sh`
- ADR-009 (mock principal / future OIDC)
