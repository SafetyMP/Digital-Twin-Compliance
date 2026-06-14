# ADR-004: Polyglot Datastore Selection

**Status**: Accepted  
**Date**: 2026-06-13  
**Deciders**: Platform Architecture Team

## Context

The platform manages diverse data with different access patterns, consistency requirements, and retention policies:

- Transactional entity state (CRUD, relational queries)
- Event streams (append-only, replay, high throughput)
- Graph relationships (traversal, centrality, path finding)
- Time-series metrics (aggregations, rollups, analytics)
- Immutable audit records (append-only, cryptographic verification)
- Cache and online features (low-latency key-value)
- Large binary artifacts (reports, snapshots, retention-locked)

No single database can optimally serve all these patterns.

## Decision

Adopt a **polyglot persistence** strategy with purpose-built stores:

| Store | Technology | Primary Data | Access Pattern |
|-------|------------|--------------|----------------|
| **Entity state** | PostgreSQL 16+ | TwinPersona, Account, Instrument, Contract, Rule metadata | CRUD, relational joins, ACID |
| **Event log** | Apache Kafka | All domain and compliance events | Append-only, replay, pub/sub |
| **Graph** | Neo4j 5+ | Exposures, interconnections, contagion paths | Traversal, Cypher queries |
| **Time-series** | ClickHouse | Monitoring metrics, aggregation inputs, report data | Columnar analytics, rollups |
| **Audit ledger** | immudb | Compliance decisions, alerts, access events | Append-only, verified reads |
| **Cache / features** | Redis 7+ | Online features, rate counters, session cache | Key-value, TTL, sub-ms |
| **Artifacts** | S3-compatible (Object Lock) | Reports, evidence snapshots, savepoints | Object read/write, immutability |

### PostgreSQL

- Primary transactional store for entity state and configuration
- Outbox pattern for reliable event publishing (Debezium CDC → Kafka)
- Extensions: `pgcrypto` (hashing), `pg_audit` (DB-level audit supplement)
- HA: Patroni or cloud-managed (RDS, Cloud SQL, Neon)

### Neo4j

- Exposure graph with typed relationships and layer attributes
- GDS library for centrality, community detection, path algorithms
- Periodic snapshot export to S3 for simulation service
- Alternative: Memgraph for lower-latency in-memory graph (evaluate if Neo4j latency insufficient)

### ClickHouse

- Monitoring metrics: alert counts, evaluation latency, breach rates by regime
- Pre-aggregated tables for regulatory report inputs
- Materialized views for dashboard queries
- 3-year hot retention; older data archived to S3 Parquet

### Redis

- Online feature store for Flink jobs (rolling counters, last-N events)
- API response cache (persona state, graph summaries)
- Rate limiting counters for API gateway
- No persistent compliance data — ephemeral only

## Consequences

### Positive

- Each store optimized for its access pattern
- No forced compromises (e.g., graph queries in SQL, time-series in PostgreSQL)
- Independent scaling per store
- Clear ownership boundaries per bounded context

### Negative

- Operational complexity: 6+ data stores to deploy, monitor, backup
- Data consistency across stores requires event-driven synchronization
- Team needs expertise across multiple technologies
- Cross-store queries require application-level joins or materialized views

### Mitigations

- Event-driven sync via Kafka — single source of truth is the event log
- PostgreSQL holds current state; other stores are derived/indexed views
- Terraform modules for each store with standardized backup/monitoring
- Start with managed services where possible (RDS, Confluent Cloud, Neo4j Aura)
- ClickHouse can be deferred to Phase 5 if PostgreSQL aggregations suffice initially

## Alternatives Considered

| Alternative | Rejected Because |
|-------------|------------------|
| **PostgreSQL only** | Poor graph traversal; not suitable for time-series at scale; not tamper-evident |
| **MongoDB only** | Weak graph queries; no event log; no tamper-evident audit |
| **Single cloud data warehouse (Snowflake/BigQuery)** | Batch-oriented; unsuitable for real-time twin state and sub-second monitoring |
| **EventStoreDB as primary store** | Event sourcing for everything adds complexity; PostgreSQL + outbox is simpler for CRUD state |
| **TimescaleDB instead of ClickHouse** | Good for time-series but weaker for analytical aggregations at report scale |

## References

- [PostgreSQL Best Practices (Supabase)](https://supabase.com/docs/guides/database)
- [Neo4j Graph Data Science](https://neo4j.com/docs/graph-data-science/)
- [ClickHouse for Real-Time Analytics](https://clickhouse.com/docs/en/intro)
