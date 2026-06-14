# ADR Index

Architecture Decision Records for the Financial Digital Twin + Compliance Platform.

| ADR | Title | Status |
|-----|-------|--------|
| [001](./001-kafka-flink-streaming.md) | Kafka + Flink for Real-Time Event Streaming and Monitoring | Accepted |
| [002](./002-cedar-decision-engine.md) | Cedar + GoRules Zen for Policy and Decision Engine | Accepted |
| [003](./003-immudb-audit-ledger.md) | immudb for Tamper-Evident Audit Ledger | Accepted |
| [004](./004-datastore-selection.md) | Polyglot Datastore Selection | Accepted |
| [005](./005-gorules-zen-vs-drools.md) | GoRules Zen vs Drools for Business Rules | Accepted |
| [006](./006-polyglot-language-strategy.md) | Polyglot Language Strategy | Accepted |
| [007](./007-phase1-foundation-decisions.md) | Phase 1 Foundation Decisions (D1, D4, D9) | Accepted |
| [008](./008-phase2-foundation-decisions.md) | Phase 2 Foundation Decisions (D10–D13) | Accepted |

## ADR Template

Each ADR follows this structure:
- **Status**: Proposed | Accepted | Deprecated | Superseded
- **Context**: What problem are we solving?
- **Decision**: What did we decide?
- **Consequences**: What are the trade-offs?
- **Alternatives Considered**: What else was evaluated?
