-- Phase 2 seed patch: deterministic INT-M002 exposure pair
BEGIN;

-- Two instruments for Alpha Bank Berlin -> same counterparty (smoke INT-M002)
UPDATE instruments
SET owner_entity_id = '11111111-1111-1111-1111-111111111102',
    counterparty_id = '22222222-2222-2222-2222-222222222202',
    notional_amount = 6000000.00,
    updated_at = now()
WHERE instrument_id IN (
  SELECT instrument_id FROM instruments ORDER BY instrument_id LIMIT 2
);

COMMIT;
