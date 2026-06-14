CREATE TABLE IF NOT EXISTS payments (
  payment_id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  source_account_id      UUID NOT NULL REFERENCES accounts(account_id),
  destination_account_id UUID NOT NULL REFERENCES accounts(account_id),
  amount                 NUMERIC(20,2) NOT NULL,
  currency               CHAR(3) NOT NULL DEFAULT 'EUR',
  status                 TEXT NOT NULL DEFAULT 'Pending',
  initiated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at             TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE payments REPLICA IDENTITY FULL;

-- Owner institution for exposure aggregation (INT-M002)
ALTER TABLE instruments ADD COLUMN IF NOT EXISTS owner_entity_id UUID REFERENCES legal_entities(entity_id);
UPDATE instruments SET owner_entity_id = counterparty_id WHERE owner_entity_id IS NULL;

DO $$
BEGIN
  ALTER PUBLICATION dbz_publication ADD TABLE payments;
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_payments_source_account ON payments(source_account_id);
CREATE INDEX IF NOT EXISTS idx_payments_initiated_at ON payments(initiated_at);
CREATE INDEX IF NOT EXISTS idx_instruments_owner ON instruments(owner_entity_id);
CREATE INDEX IF NOT EXISTS idx_instruments_counterparty ON instruments(counterparty_id);
