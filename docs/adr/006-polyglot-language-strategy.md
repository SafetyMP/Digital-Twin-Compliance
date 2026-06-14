# ADR-006: Polyglot Language Strategy

**Status**: Accepted  
**Date**: 2026-06-13  
**Deciders**: Platform Architecture Team

## Context

The platform spans event streaming, real-time monitoring, policy evaluation, graph analytics, simulation, regulatory reporting, and web dashboards. No single programming language is optimal for all concerns. The team must decide whether to:

1. Standardize on one language (simplicity, hiring)
2. Adopt a polyglot strategy (best tool per concern)

Given the diverse technical requirements — sub-millisecond policy evaluation, exactly-once stream processing, agent-based simulation, XBRL generation, interactive graph visualization — a polyglot approach is warranted.

## Decision

Adopt a **polyglot language strategy** assigning languages by concern:

| Language | Components | Rationale |
|----------|------------|-----------|
| **Go** | API Gateway, State Service, Graph Service, Ingestion Connectors, Audit Service, Cedar/Zen integration | High throughput, simple concurrency, small binaries, strong gRPC/HTTP ecosystem |
| **Java** | Apache Flink CEP jobs | Flink native; exactly-once streaming maturity; largest CEP ecosystem |
| **Python** | Simulation Service, Reporting Service, ML/analytics scripts | Best ecosystem for NetworkX, pandas, PyTorch, XBRL/SDMX libraries |
| **TypeScript** | Next.js web app, shared type definitions | Type-safe frontend; shared schemas with API contracts |
| **Rust** | Cedar engine core, GoRules Zen engine core (via SDK bindings) | Memory safety, sub-ms performance; accessed via Go/Python SDKs, not authored directly |
| **SQL** | PostgreSQL queries, ClickHouse analytics, Neo4j Cypher, immudb SQL | Declarative data access per store |
| **Cedar** | Policy definitions | Purpose-built authorization language with formal verification |
| **JDM (JSON)** | Decision models | Business-readable rule definitions |

### Language Boundaries

```mermaid
flowchart TB
  subgraph goServices [Go Services]
    api[API Gateway]
    state[State Service]
    graph[Graph Service]
    ingest[Connectors]
    audit[Audit Service]
  end

  subgraph javaServices [Java Services]
    flink[Flink CEP Jobs]
  end

  subgraph pythonServices [Python Services]
    sim[Simulation Service]
    report[Reporting Service]
  end

  subgraph tsServices [TypeScript]
    web[Next.js Web App]
  end

  subgraph declarative [Declarative]
    cedar[Cedar Policies]
    jdm[Decision Models]
    sql[SQL / Cypher]
  end

  goServices <-->|gRPC / REST| javaServices
  goServices <-->|gRPC / REST| pythonServices
  goServices <-->|REST / WebSocket| tsServices
  cedar --> goServices
  jdm --> goServices
```

### Inter-Service Communication

| From | To | Protocol | Serialization |
|------|----|----------|---------------|
| Connectors | Kafka | Kafka protocol | Avro |
| Flink | Decision Service | gRPC | Protobuf |
| Flink | Cedar Service | gRPC | Protobuf |
| API Gateway | All services | gRPC / REST | Protobuf / JSON |
| Web App | API Gateway | REST / WebSocket | JSON |
| Simulation | Graph Service | gRPC | Protobuf |
| Reporting | PostgreSQL, immudb | SQL | — |

### Shared Contracts

- **Avro schemas** in a shared Git repository for all Kafka events
- **Protobuf definitions** for gRPC service interfaces
- **OpenAPI spec** for REST endpoints consumed by the web app
- **JSON Schema** for decision model validation

## Consequences

### Positive

- Each component uses the best-suited runtime for its concern
- Go services provide consistent operational profile (small containers, fast startup)
- Python unlocks best-in-class analytics and reporting libraries
- Java required only for Flink (isolated, well-understood boundary)
- TypeScript provides type-safe frontend with shared contract types

### Negative

- Multiple language toolchains in CI/CD
- Cross-language debugging is harder
- Hiring requires broader skill set (or team silos by language)
- Shared type definitions must be generated, not hand-maintained

### Mitigations

- Shared schema repository with code generation (Avro → Go/Java/Python, Protobuf → Go/Python)
- Standardized service template in Go (80%+ of services)
- Flink jobs are self-contained Java modules with minimal cross-language interaction
- Python services expose gRPC interfaces matching Protobuf contracts
- Docker multi-stage builds per language with standardized base images

## Alternatives Considered

| Alternative | Rejected Because |
|-------------|------------------|
| **Java everywhere** | Weak analytics/ML ecosystem; verbose for API services; frontend still needs TypeScript |
| **Python everywhere** | GIL limits concurrency for ingestion/API; Flink still needs Java |
| **TypeScript everywhere** | No mature exactly-once stream processing; weak simulation/ML ecosystem |
| **Rust everywhere** | Steep learning curve; immature web/gRPC ecosystem; overkill for CRUD services |
| **Go everywhere** | No mature CEP/stream processing; weak ML/reporting ecosystem |

## Team Structure Recommendation

| Team | Primary Language | Components |
|------|------------------|------------|
| **Platform** | Go | API, State, Graph, Ingestion, Audit |
| **Streaming** | Java | Flink CEP jobs |
| **Analytics** | Python | Simulation, Reporting |
| **Frontend** | TypeScript | Web app |
| **Compliance Engineering** | Cedar + JDM | Policy and rule authoring (cross-team) |

## References

- [ADR-001: Kafka + Flink](./001-kafka-flink-streaming.md)
- [ADR-002: Cedar + Zen](./002-cedar-decision-engine.md)
- [ADR-004: Datastore Selection](./004-datastore-selection.md)
