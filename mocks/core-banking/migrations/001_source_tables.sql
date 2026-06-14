CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS legal_entities (
  entity_id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  legal_name           TEXT NOT NULL,
  lei                  TEXT,
  entity_type          TEXT NOT NULL CHECK (entity_type IN ('Bank','Fund','SPV','CCP','ICTProvider','InternalUnit')),
  jurisdiction         CHAR(2) NOT NULL,
  parent_entity_id     UUID REFERENCES legal_entities(entity_id),
  consolidation_scope  TEXT NOT NULL DEFAULT 'Solo',
  created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS accounts (
  account_id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  account_number   TEXT NOT NULL,
  account_type     TEXT NOT NULL,
  currency         CHAR(3) NOT NULL,
  owner_entity_id  UUID NOT NULL REFERENCES legal_entities(entity_id),
  status           TEXT NOT NULL DEFAULT 'Active',
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS instruments (
  instrument_id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  isin               TEXT,
  instrument_type    TEXT NOT NULL,
  counterparty_id    UUID,
  notional_amount    NUMERIC(20,2) NOT NULL,
  currency           CHAR(3) NOT NULL,
  maturity_date      DATE,
  regulatory_class   TEXT,
  created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE legal_entities REPLICA IDENTITY FULL;
ALTER TABLE accounts REPLICA IDENTITY FULL;
ALTER TABLE instruments REPLICA IDENTITY FULL;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_publication WHERE pubname = 'dbz_publication') THEN
    CREATE PUBLICATION dbz_publication FOR TABLE legal_entities, accounts, instruments;
  END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_legal_entities_parent ON legal_entities(parent_entity_id);
CREATE INDEX IF NOT EXISTS idx_accounts_owner ON accounts(owner_entity_id);
