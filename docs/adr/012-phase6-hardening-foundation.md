# ADR-012: Phase 6 Hardening Foundation (D30–D35)

**Status**: Accepted (planning)  
**Date**: 2026-07-25  
**Deciders**: Platform Architecture Team  
**Implements**: CECT master-spec §2.2  
**Supersedes (partial)**: ADR-009 D20 for **public Compose/staging edges** (Keycloak/OIDC required; mock principals may remain only on explicitly documented internal/dev-only paths)

## Context

Phase 3 deferred Keycloak (D20). Phase 5 (ADR-011) adds reporting surfaces. Phase 6 must raise the **reference-deploy hardening bar** without claiming full multi-tenant supervisory SaaS (CR-5): OIDC, TLS edges, secrets hygiene, OpenTelemetry, load/DR honesty, and regulator-facing explainability.

| ID | Decision | Phase 6 impact |
|----|----------|----------------|
| D30 | Identity on public edges | Keycloak/OIDC vs mock headers |
| D31 | TLS termination | Edge proxy vs per-service certs |
| D32 | Secrets injection | Compose plaintext vs injected secrets |
| D33 | Observability | OpenTelemetry exporters |
| D34 | Load evidence | 10K evt/s vs CI-sized substitute |
| D35 | Ops docs | DR runbooks + explainability + residuals |

Open risk: **UR-CECT-002** (10K evt/s may not fit default CI hosts).

## Decision

### D30 — Keycloak OIDC on public API/UI edges

**Decision**: Run **Keycloak** in Compose (staging/hardening profile acceptable). Protect public HTTP/WebSocket edges (State, Alert, Audit, Graph, Simulation, Reporting APIs and Next.js consoles) with **OIDC bearer validation**. Unauthenticated requests return **401**. Valid token path must succeed in the documented profile. **Remove mock principal bypass** from protected routes (AC-A-P6-001).

Cedar/Zen continue to receive a resolved principal/roles from validated tokens (or a thin gateway mapping) — engines are **not** replaced (CR-3).

**Rationale**: Roadmap D10; closes D20 deferral for public edges; R-S-001.

### D31 — TLS at the public edge

**Decision**: Terminate **TLS** at a single Compose edge (e.g. Caddy/Traefik/nginx) in front of public UI/API. Document the path in `docs/deployment.md`. Prove with `curl` (proper CA or documented `-k` for local self-signed) — AC-A-P6-002.

**Rationale**: Reference hardening without per-service mTLS complexity in Phase 6.

### D32 — No plaintext production secrets in Compose

**Decision**: Staging/hardening profile MUST NOT commit production credentials. Use env files / Docker secrets / documented injection. `.env.example` stays non-secret. Call out residual gaps in `docs/deployment.md` (CR-5).

**Rationale**: Security baseline without requiring Hashicorp Vault as a Phase 6 exit dependency (Vault may be documented as staging upgrade).

### D33 — OpenTelemetry across polyglot services

**Decision**: Instrument Go / Java (Flink where practical) / Python / TS services with **OpenTelemetry**. Ship a collector + Jaeger (or Grafana Tempo) in Compose. Exit requires **≥1 end-to-end trace spanning ≥2 services** (AC-A-P6-003).

**Rationale**: Matches roadmap observability deliverable; supports incident DR and explainability.

### D34 — Load profile with CI-sized honesty

**Decision**: Document a **10K evt/s** target profile and provide either (a) a scaled load script + evidence, or (b) a **CI-sized substitute** profile with explicit residual that full 10K was not run on CI hosts (UR-CECT-002). Do not silently claim 10K from a tiny smoke.

**Rationale**: CR-5 honesty.

### D35 — DR runbooks + explainability pack + deployment residuals

**Decision**:

1. DR runbooks under `docs/` for **Kafka**, **Flink**, and **immudb** recovery procedures (tested enough to be executable, not prose-only).  
2. Regulator-facing **`docs/explainability.md`** covering risk scores and policy decisions with audit linkage.  
3. **`docs/deployment.md`** lists residual gaps (HA immudb, full K8s/Terraform, multi-tenant SaaS, etc.).

Full Kubernetes/Terraform/immudb HA from the original roadmap Phase 6 wish-list are **reference-documented residuals**, not silent pass criteria for CECT AC-A-P6-* unless separately evidenced.

**Rationale**: AC-A-P6-004 / AC-A-P6-005; CR-5.

## Consequences

### Positive

- Public edges become authn-real; mock bypass debt shrinks.  
- Traces and DR docs make the reference deploy operable.  
- Residuals stay visible — marketing cannot imply SaaS/HA completion.

### Negative

- Keycloak + TLS + OTel increase Compose complexity and smoke runtime.  
- Touching every public service for OIDC risks Phase 1–4 regressions — regression gate is mandatory.  
- 10K evt/s may remain residual on laptop/CI.

## Alternatives Considered

| Decision | Alternative | Why rejected for Phase 6 |
|----------|-------------|--------------------------|
| D30 | Keep X-Principal on all routes | Fails AC-A-P6-001 / CR-5 |
| D30 | Auth0/cloud IdP only | Vendor lock-in vs roadmap Keycloak preference |
| D31 | Per-service TLS only | Operational sprawl; harder local DX |
| D34 | Claim 10K without evidence | Violates CR-4/CR-5 |
| D35 | Require full K8s HA for exit | Out of CECT honesty scope unless evidenced |

## Exit wiring

- Evidence: `evidence/phase6-oidc-*.txt`, `evidence/phase6-tls-*.txt`, `evidence/phase6-otel-*.txt`, docs checks  
- Gates: `G-CECT-ADR-012`, `G-CECT-P6-OIDC`, `G-CECT-P6-HARDEN`, `G-CECT-REGRESSION`  
- Acceptance: `AC-A-P6-001` … `AC-A-P6-005`, `AC-A-REG-001`

## References

- CECT `master-spec.md` §2.2, §0 CR-3/CR-4/CR-5  
- [ADR-009](./009-phase3-foundation-decisions.md) D20  
- [roadmap.md](../roadmap.md) — Phase 6
