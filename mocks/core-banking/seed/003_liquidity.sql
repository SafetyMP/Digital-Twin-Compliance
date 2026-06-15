-- Phase 2 seed patch: institution liquidity baseline (source of truth for twin enrichment)
BEGIN;

UPDATE legal_entities
SET
  lcr = 1.05,
  hqla = 500000000.00,
  net_cash_outflows_30d = 476190476.00,
  liquidity_currency = 'EUR',
  updated_at = now()
WHERE lcr IS NULL;

UPDATE legal_entities
SET
  lcr = 0.95,
  hqla = 450000000.00,
  net_cash_outflows_30d = 473684211.00,
  liquidity_currency = 'EUR',
  updated_at = now()
WHERE entity_id = '44444444-4444-4444-4444-444444444401';

COMMIT;
