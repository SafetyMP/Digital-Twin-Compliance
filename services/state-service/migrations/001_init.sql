CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS twin_personas (
  persona_id       UUID PRIMARY KEY,
  tenant_id        UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001',
  source_entity_id UUID NOT NULL,
  persona_type     TEXT NOT NULL,
  state_version    INT NOT NULL DEFAULT 1,
  current_state    JSONB NOT NULL DEFAULT '{}',
  compliance_status TEXT NOT NULL DEFAULT 'Unknown',
  last_synced_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, source_entity_id, persona_type)
);

CREATE TABLE IF NOT EXISTS accounts (
  account_id       UUID PRIMARY KEY,
  tenant_id        UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001',
  account_number   TEXT NOT NULL,
  account_type     TEXT NOT NULL,
  currency         CHAR(3) NOT NULL,
  owner_entity_id  UUID NOT NULL,
  status           TEXT NOT NULL,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS instruments (
  instrument_id    UUID PRIMARY KEY,
  tenant_id        UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001',
  isin             TEXT,
  instrument_type  TEXT NOT NULL,
  notional_amount  NUMERIC(20,2) NOT NULL,
  currency         CHAR(3) NOT NULL,
  maturity_date    DATE,
  regulatory_class TEXT,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS processed_events (
  idempotency_key  TEXT PRIMARY KEY,
  processed_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS outbox (
  id               BIGSERIAL PRIMARY KEY,
  topic            TEXT NOT NULL,
  partition_key    TEXT NOT NULL,
  payload          JSONB NOT NULL,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  published_at     TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS outbox_unpublished ON outbox (id) WHERE published_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_twin_personas_type ON twin_personas (tenant_id, persona_type);
CREATE INDEX IF NOT EXISTS idx_twin_personas_source ON twin_personas (tenant_id, source_entity_id);
