# ADR-003: immudb for Tamper-Evident Audit Ledger

**Status**: Accepted  
**Date**: 2026-06-13  
**Deciders**: Platform Architecture Team

## Context

Financial compliance requires an immutable, tamper-evident audit trail for:
- Every compliance rule evaluation (inputs, policy version, decision, rationale)
- Every compliance alert and its lifecycle
- Every twin state transition
- Every regulatory report generation
- Every access to sensitive compliance data

Amazon QLDB — previously the leading managed immutable ledger — was deprecated with end-of-support in July 2025. A replacement must provide:
- Cryptographic verification (hash chain / Merkle tree)
- Append-only semantics
- SQL or SDK access for audit queries
- Self-hostable (avoid CLOUD Act / GDPR jurisdiction concerns)
- FIPS-compliant algorithms
- Long-term retention (7–10 years)

## Decision

Adopt **immudb** (Apache 2.0, Codenotary) as the primary tamper-evident audit ledger.

Supplement with **S3 Object Lock (Compliance mode)** for large evidence artifacts (generated reports, state snapshots, Flink savepoints).

### immudb Configuration

- **Deployment**: Self-hosted on Kubernetes (or managed via Codenotary Cloud if jurisdiction permits)
- **Interface**: SQL + gRPC SDK (Go client in Audit Service)
- **Verification**: Built-in client-side cryptographic verification on every read
- **Retention**: Configurable per-table; compliance entries retained 7–10 years
- **Backup**: Continuous backup to S3 with Object Lock

### S3 Object Lock (Evidence Artifacts)

- **Mode**: Compliance (no deletion by anyone, including root, until retention expires)
- **Contents**: XBRL/SDMX reports, evidence snapshots, decision log exports
- **Indexing**: DynamoDB or PostgreSQL metadata table for artifact lookup
- **Integrity**: KMS signatures + `previous_hash` field in metadata

### Audit Entry Flow

1. Compliance event occurs (decision, alert, state change, report)
2. Audit Service serializes event to `AuditEntry` schema
3. Write to immudb (hash chain updated atomically)
4. If artifact attached, write to S3 Object Lock with metadata reference
5. Publish `AuditRecorded` event to Kafka for downstream consumers

## Consequences

### Positive

- Direct functional replacement for QLDB with open-source license
- Cryptographic verification without trusting the vendor
- Self-hostable on EU or any jurisdiction infrastructure (GDPR-friendly)
- SQL interface familiar to compliance officers and auditors
- FIPS-compliant algorithms
- Built-in consistency and integrity checks (server + client + gateway)
- Active development and community post-QLDB deprecation

### Negative

- Self-hosted operational responsibility (backup, HA, monitoring)
- Smaller managed-service ecosystem than QLDB had
- DELETE semantics record deletion in hash chain (not true erasure) — GDPR Art. 17 requires careful key expiration configuration
- Performance for high-volume transactional writes needs benchmarking (QLDB had latency concerns at scale)

### Mitigations

- Buffer audit writes to Kafka topic (`audit.pending`) if immudb is temporarily unavailable; replay on recovery
- immudb supports TTL-based key expiration for GDPR compliance
- Start with single-node immudb; scale to cluster for HA in Phase 6
- Benchmark write throughput early in Phase 3

## Alternatives Considered

| Alternative | Rejected Because |
|-------------|------------------|
| **Amazon QLDB** | Deprecated July 2025; not accepting new customers |
| **S3 Object Lock only** | No native query interface; requires custom indexing; no built-in hash chain |
| **PostgreSQL + pgaudit** | Not cryptographically verifiable; mutable by DBA |
| **Hyperledger Fabric** | Over-engineered for audit log use case; high operational complexity |
| **Dolt (version-controlled SQL)** | Git-style versioning useful but lacks built-in tamper detection |
| **Azure SQL Ledger Tables** | Vendor lock-in; single-vendor jurisdiction risk |
| **EventStoreDB** | Event sourcing DB, not tamper-evident ledger; different use case |

## References

- [immudb — Immutable Database](https://immudb.io/)
- [QLDB Deprecated — Alternatives (DoltHub)](https://www.dolthub.com/blog/2024-08-12-qldb-deprecated-alternatives/)
- [AWS QLDB EU Alternative — GDPR/CLOUD Act (sota.io)](https://sota.io/blog/aws-qldb-eu-alternative-gdpr-cloud-act-2026)
- [S3 Object Lock for Immutable Audit Ledger (TrustWarden)](https://www.trustwarden.ai/blog/why-s3-object-lock-over-qldb/)
