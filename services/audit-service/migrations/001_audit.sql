CREATE TABLE IF NOT EXISTS audit_idempotency_keys (
  idempotency_key TEXT PRIMARY KEY,
  entry_id        UUID NOT NULL,
  tenant_id       UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001',
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS audit_entry_index (
  entry_id         UUID PRIMARY KEY,
  tenant_id        UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001',
  sequence_number  BIGINT NOT NULL UNIQUE,
  entry_type       TEXT NOT NULL,
  rule_code        TEXT,
  subject_id       TEXT,
  subject_type     TEXT,
  correlation_id   TEXT,
  recorded_at      TIMESTAMPTZ NOT NULL,
  payload_hash     TEXT NOT NULL,
  previous_hash    TEXT NOT NULL,
  idempotency_key  TEXT NOT NULL UNIQUE,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS audit_entry_index_rule_recorded
  ON audit_entry_index (rule_code, recorded_at DESC);
CREATE INDEX IF NOT EXISTS audit_entry_index_subject_recorded
  ON audit_entry_index (subject_id, recorded_at DESC);
CREATE INDEX IF NOT EXISTS audit_entry_index_recorded
  ON audit_entry_index (recorded_at DESC);

CREATE TABLE IF NOT EXISTS audit_outbox_dlq (
  id              BIGSERIAL PRIMARY KEY,
  tenant_id       UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001',
  idempotency_key TEXT,
  original_topic  TEXT NOT NULL,
  partition       INT NOT NULL,
  kafka_offset    BIGINT NOT NULL,
  error_message   TEXT NOT NULL,
  payload         JSONB NOT NULL,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS audit_outbox_dlq_created ON audit_outbox_dlq (created_at DESC);
