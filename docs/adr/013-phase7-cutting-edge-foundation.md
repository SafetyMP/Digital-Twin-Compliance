# ADR-013: Phase 7 Cutting-Edge Foundation (D36–D39)

**Status**: Accepted (planning)  
**Date**: 2026-07-25  
**Deciders**: Platform Architecture Team  
**Implements**: CECT master-spec §3  
**Related**: ADR-010 D23 (deterministic stress remains; agentic/continuous contagion added here under audit linkage)

## Context

Tranche B ships only after Tranche A smoke-green for **cutting-edge claims** (CR-2). ADRs may be authored earlier. Phase 7 differentiators: control-effectiveness twin, continuous/agentic contagion→audit, reg→policy drafts (proposal-only), and graph path/centrality analytics. Industrial OT / RTGS twins remain out of scope (CR-1).

| ID | Decision | Phase 7 impact |
|----|----------|----------------|
| D36 | Control-effectiveness twin | Replay + metrics + audit |
| D37 | Continuous / agentic contagion | Loop over Neo4j + audit |
| D38 | Reg→policy assist | Proposal-only; CI reject/accept |
| D39 | Graph supervisory analytics | Path + centrality APIs |

Open risk: **UR-CECT-003** (LLM path optional; CI fixtures).

## Decision

### D36 — Control-effectiveness twin with audit-linked replay

**Decision**: Provide a **CLI and/or API** (under `scripts/` and/or `services/` / `apps/`) that replays labeled breach scenarios against detection controls, emits **coverage** and **miss rate** metrics, and records each run with a verifiable **`evidenceRef`** on the audit chain (AC-B-P7-001). Sandbox threshold / Cedar/Zen experiment hooks may be included but must not auto-promote policies (CR-3).

**Rationale**: July 2026 “prove detection works” overlay; falsifiable vs prose demos.

### D37 — Continuous/agentic contagion loop → audit

**Decision**: Lift deferred **agent-based / continuous contagion** (roadmap D5 / ADR-010 D23 stretch) into a **scheduled or streaming loop** over the Neo4j exposure graph. Deterministic ECB Adverse path from Phase 4 **remains available**. Each continuous/agentic run writes an audit entry with path/seed explainability (AC-B-P7-002).

**Rationale**: Cutting-edge differentiator; keeps Simulation Service + Graph Service as owners; audit-only write path unchanged.

### D38 — Reg→policy drafts are proposals only (CI fixture path)

**Decision**: Add a pipeline that produces **Cedar/Zen stub drafts** from structured regulatory snippets. **CI MUST demonstrate reject (invalid stub) and accept (valid stub)** paths. **No auto-deploy / auto-promote** to production policy sets (CR-3). LLM assistance is **optional**; default CI path is **deterministic fixtures** (UR-CECT-003 / R-D-002). Human review + existing `./scripts/run-policy-ci.sh` (or successor) remain authoritative.

**Rationale**: AC-B-P7-003; preserves Cedar + Zen as engines of record.

### D39 — Graph path and centrality supervisory APIs

**Decision**: Extend **Graph Service** with APIs returning **multi-hop paths** and **centrality** metrics for seed institutions (financial contagion paths — not ICS ATT&CK). Smoke under `evidence/phase7-graph-*.txt` (AC-B-P7-004). Static viz-only changes without API metrics **fail** the gate.

**Rationale**: Supervisory analytics beyond Phase 4 explorer summary.

### Claims discipline

README/ROADMAP MUST NOT claim “cutting-edge OSS supervisory twin” until critical `AC-A-*` and `AC-B-*` have executable PASS evidence (AC-B-CLAIM-001 / CR-4). Prefer a claim-lint check in verify.

## Consequences

### Positive

- Differentiating demos are smoke/evidence gated.  
- Policy engines stay authoritative; AI assist cannot silently ship policy.  
- Contagion gains continuity without discarding deterministic Phase 4 path.

### Negative

- Agentic loops can be flaky/non-deterministic — require seed control + audit explainability.  
- Effectiveness replay needs curated labeled breaches (fixture maintenance).  
- Graph GDS/centrality may need Neo4j plugin/config growth in Compose.

## Alternatives Considered

| Decision | Alternative | Why rejected for Phase 7 |
|----------|-------------|--------------------------|
| D36 | Prose-only effectiveness narrative | Violates CR-4 |
| D37 | Replace deterministic ECB Adverse | Breaks Phase 4 regression value |
| D38 | Auto-merge policy stubs to main | Violates CR-3 |
| D38 | Require live LLM in CI | UR-CECT-003; nondeterministic |
| D39 | UI-only centrality charts | Fails API smoke / G-CECT-P7-GRAPH |
| — | Industrial OT ATT&CK twin | CR-1 out of scope |

## Exit wiring

- Evidence: `evidence/phase7-effectiveness-*.txt`, `phase7-contagion-*.txt`, `phase7-reg2policy-*.txt`, `phase7-graph-*.txt`  
- Gates: `G-CECT-ADR-013`, `G-CECT-P7-EFFECT`, `G-CECT-P7-CONTAGION`, `G-CECT-P7-REG2POLICY`, `G-CECT-P7-GRAPH`  
- Acceptance: `AC-B-P7-001` … `AC-B-P7-004`, `AC-B-CLAIM-001`

## References

- CECT `master-spec.md` §3, §0 CR-1/CR-2/CR-3/CR-4  
- [ADR-010](./010-phase4-foundation-decisions.md) D23–D24  
- [ADR-002](./002-cedar-decision-engine.md), [ADR-005](./005-gorules-zen-vs-drools.md)
