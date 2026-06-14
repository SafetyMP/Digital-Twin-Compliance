# ADR-001: Kafka + Flink for Real-Time Event Streaming and Monitoring

**Status**: Accepted  
**Date**: 2026-06-13  
**Deciders**: Platform Architecture Team

## Context

The platform requires a durable, replayable event backbone connecting source systems to the digital twin and compliance overlay. Real-time monitoring must detect compliance breaches (velocity limits, liquidity thresholds, settlement bottlenecks) with sub-second to low-second latency and exactly-once processing semantics suitable for financial accuracy.

Key requirements:
- Durable event log with 7–10 year retention for regulatory compliance
- Stateful stream processing for CEP patterns and windowed aggregations
- Exactly-once semantics for financial event processing
- Schema-first design with enforced contracts between producers and consumers
- Horizontal scalability to 10K+ events/sec (initial target)

## Decision

Adopt **Apache Kafka (KRaft mode)** as the event backbone and **Apache Flink** as the stateful stream processing engine for real-time compliance monitoring.

### Kafka Configuration

- **Mode**: KRaft (no ZooKeeper dependency)
- **Schema Registry**: Confluent Schema Registry with Avro schemas
- **Retention**: 7–10 years on compliance-critical topics; tiered storage for cost management
- **Partitioning**: By entity ID for ordering guarantees
- **Replication factor**: 3 (production)

### Flink Configuration

- **State backend**: RocksDB with incremental checkpointing
- **Checkpoint storage**: S3-compatible object storage
- **Delivery guarantee**: EXACTLY_ONCE
- **Deployment**: Flink Kubernetes Operator

## Consequences

### Positive

- Industry-standard stack with extensive production track record in fintech (fraud detection, AML, transaction monitoring)
- Kafka provides durable replay enabling twin rebuild and audit reconstruction
- Flink CEP supports declarative pattern detection for complex compliance scenarios
- Exactly-once semantics prevent duplicate alerts and incorrect aggregations
- KRaft eliminates ZooKeeper operational complexity
- Large ecosystem of connectors (Debezium CDC, ClickHouse sink, etc.)

### Negative

- Operational complexity: Kafka + Flink require dedicated platform engineering
- Flink jobs require careful state management; checkpoint failures are a primary failure mode
- Java/Scala skill set required for Flink job development
- Latency overhead from checkpointing (typically 30–60s checkpoint intervals)

### Mitigations

- Start with 2–3 simple Flink jobs; add complexity incrementally
- Monitor checkpoint duration and backpressure from day one
- Use Flink SQL for simpler aggregations where possible
- Managed Kafka (Confluent Cloud, AWS MSK) as optional path to reduce ops burden

## Alternatives Considered

| Alternative | Rejected Because |
|-------------|------------------|
| **RabbitMQ / NATS** | No durable replay log; insufficient retention for compliance |
| **AWS Kinesis + Lambda** | Lambda not suitable for stateful CEP; vendor lock-in; cost at scale |
| **Kafka Streams** | Less mature CEP than Flink; harder to operate complex stateful jobs |
| **Spark Structured Streaming** | Micro-batch latency (seconds) vs Flink's true streaming (sub-second) |
| **Pulsar** | Smaller ecosystem; less fintech production evidence than Kafka |

## References

- [Real-time Streaming with Apache Kafka and Flink: Architecture for 2026](https://core.cz/en/blog/2026/realtime-streaming-kafka-flink-2026/)
- [Real-Time Transaction Monitoring Guide](https://didit.me/blog/real-time-transaction-monitoring/)
