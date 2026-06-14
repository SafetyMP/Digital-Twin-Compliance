# ADR-005: GoRules Zen vs Drools for Business Rules

**Status**: Accepted  
**Date**: 2026-06-13  
**Deciders**: Platform Architecture Team

## Context

Tier 2 of the policy architecture (see [ADR-002](./002-cedar-decision-engine.md)) requires a decision engine for complex regulatory logic and risk scoring. Candidates:

- **GoRules Zen** — Open-source, Rust core, JSON Decision Model, embeddable in Go/Node/Python
- **Drools** — Mature JVM business rule engine, RETE algorithm, 20+ years of production use

Both support policy-as-code, but differ in ecosystem maturity, language binding, and operational model.

## Decision

Adopt **GoRules Zen** as the primary Tier 2 decision engine.

**Drools remains the documented fallback** if the team encounters requirements that Zen cannot satisfy (complex RETE-based forward chaining across large fact bases, existing Drools rule libraries).

## Comparison

| Dimension | GoRules Zen | Drools |
|-----------|-------------|--------|
| **License** | Open-source (MIT) | Open-source (Apache 2.0) |
| **Core language** | Rust | Java |
| **Embedding** | Go, Node, Python, Rust | JVM only |
| **Rule format** | JSON Decision Model (JDM) | DRL, decision tables, DSL |
| **Business user access** | Web editor (GoRules Cloud) | Business Central (Red Hat) |
| **Performance** | Rust core, sub-ms | JVM, ms-range |
| **Maturity** | Newer (2023+) | 20+ years |
| **Audit trail** | Built-in decision logging | Requires custom integration |
| **CI testing** | JSON fixtures + expected outputs | JUnit + scenario tests |
| **RETE algorithm** | No (decision graph) | Yes (forward chaining) |
| **Operational model** | Embedded library | JVM process or KIE Server |

## Rationale for Zen

1. **Polyglot alignment** — Embeds in Go services (API layer, Flink sidecar) without JVM dependency
2. **Modern rule format** — JDM is JSON-native, diff-friendly, and readable by non-developers
3. **Performance** — Rust core provides sub-millisecond evaluation suitable for Flink hot paths
4. **Built-in audit** — Decision logging aligns with immudb audit requirements
5. **Lower operational overhead** — Embedded library vs JVM process management

## Rationale for Drools Fallback

Choose Drools instead if:
- Team has existing Drools rule libraries to migrate
- Requirements need RETE forward chaining across large, dynamic fact bases
- Regulatory rules are already authored in DRL format
- JVM ecosystem is already established in the organization

## Consequences

### Positive

- No JVM dependency for the primary decision path
- JSON Decision Models are Git-friendly and CI-testable
- Business users can author rules via GoRules visual editor
- Consistent with polyglot strategy (Go services + Python analytics)

### Negative

- Zen is less mature than Drools; smaller community and fewer production case studies
- Complex RETE-style reasoning not available (decision graph model instead)
- Team may need to learn JDM format
- GoRules Cloud (visual editor) is a SaaS dependency for business user rule authoring

### Mitigations

- Maintain Drools fallback path documented and tested (spike in Phase 3)
- Zen decision models stored in Git regardless of authoring tool
- CI pipeline validates all decision models with fixture inputs/outputs
- If Zen proves insufficient, swap Decision Service backend to Drools with same `RuleDecision` output schema

## Alternatives Considered

| Alternative | Rejected Because |
|-------------|------------------|
| **Drools (primary)** | JVM-only; adds operational complexity to Go-centric services |
| **DecisionRules (SaaS)** | Vendor lock-in; data residency concerns for financial compliance |
| **Custom Python rules** | Not auditable, not version-controlled as policy-as-code |
| **OpenFGA** | Relationship-based access control, not business rule evaluation |

## References

- [Top 10 Business Rule Engines 2026 (DecisionRules)](https://www.decisionrules.io/en/articles/top-10-business-rule-engines/)
- [GoRules Zen Engine](https://gorules.io/)
- [Drools Documentation](https://www.drools.org/)
