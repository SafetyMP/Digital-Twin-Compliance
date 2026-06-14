# ADR-002: Cedar + GoRules Zen for Policy and Decision Engine

**Status**: Accepted  
**Date**: 2026-06-13  
**Deciders**: Platform Architecture Team

## Context

The compliance overlay requires two distinct evaluation capabilities:

1. **Authorization and obligation checks** — Fine-grained access control and structural obligation verification (e.g., "can this user view counterparty exposure data?", "does this ICT contract meet DORA criticality requirements?"). These require formal correctness guarantees.

2. **Complex business rules and risk scoring** — Frequently changing regulatory logic, decision tables, scorecards, and multi-step rule chains (e.g., AML risk scoring, capital ratio calculations, EMIR clearing thresholds). These require business-user accessibility and rapid iteration.

A single engine cannot optimally serve both concerns.

## Decision

Adopt a **two-tier policy architecture**:

| Tier | Engine | Use Case |
|------|--------|----------|
| **Tier 1: Policy** | Cedar (AWS open-source) | Access control, obligation checks, structural compliance |
| **Tier 2: Decision** | GoRules Zen (open-source) | Complex regulatory logic, risk scoring, decision tables |

Both tiers produce `RuleDecision` outputs recorded in the audit ledger.

### Cedar (Tier 1)

- In-process evaluation via Cedar SDK (Rust core, Go bindings)
- Principal-action-resource model with typed schemas
- Formal verification via Cedar Analyzer in CI pipeline
- Sub-millisecond evaluation latency
- Policies stored as `.cedar` files in Git, deployed via CI/CD

### GoRules Zen (Tier 2)

- JSON Decision Model (JDM) format for rules
- Embeddable in Go services via Zen engine
- Decision tables, rule chains, expressions, and scorecards
- Versioned in Git with CI test fixtures
- Called synchronously from Flink jobs and Simulation Service

See [ADR-005](./005-gorules-zen-vs-drools.md) for the Zen vs Drools comparison.

## Consequences

### Positive

- Cedar provides formal verification — policies can be proven correct before deployment
- Cedar's deny-overrides semantics align with compliance "fail closed" requirements
- Zen enables business-readable decision models without sacrificing testability
- Separation of concerns: security/policy vs business logic
- Both engines support policy-as-code with Git-based versioning and CI gates

### Negative

- Two engines to operate, monitor, and train teams on
- Cedar learning curve (~15–20 hours for Rego-experienced developers)
- Zen is newer than Drools with smaller community (mitigated by active development and Rust core performance)
- Rule routing logic needed to dispatch evaluations to the correct tier

### Mitigations

- Clear rule classification guide: Cedar for access/obligation, Zen for scoring/calculation
- Shared `RuleDecision` output schema regardless of engine
- Unified CI pipeline running Cedar Analyzer + Zen test fixtures
- ADR-005 documents fallback to Drools if RETE depth is required

## Alternatives Considered

| Alternative | Rejected Because |
|-------------|------------------|
| **OPA/Rego only** | General-purpose but no formal verification; higher learning curve; slower for auth-specific patterns |
| **Drools only** | JVM-only; no formal verification; overkill for simple access checks; weak authorization model |
| **Custom rule engine** | Unacceptable risk for financial compliance; no formal guarantees |
| **Hardcoded rules in application code** | Not version-controlled, not testable in CI, not auditable |

## References

- [OPA vs Cedar: Policy-as-Code Comparison](https://www.cybersrely.com/opa-vs-cedar-ship-policy-as-code/)
- [Policy Language Comparison: Cedar, Rego, OpenFGA](https://sph.sh/en/posts/policy-language-comparison-cedar-rego-openfga/)
- [GoRules Zen Engine](https://gorules.io/)
