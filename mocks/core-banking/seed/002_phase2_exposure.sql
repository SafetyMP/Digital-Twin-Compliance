-- Phase 2 seed patch: deterministic INT-M002 exposure pair + institution liquidity baseline
BEGIN;

-- Assign owner_entity_id from institution row numbers when still null
UPDATE instruments i
SET owner_entity_id = le.entity_id
FROM (
  SELECT entity_id, row_number() OVER (ORDER BY entity_id) AS rn
  FROM legal_entities
) le
WHERE i.owner_entity_id IS NULL
  AND le.rn = (abs(hashtext(i.instrument_id::text)) % 10) + 1;

-- Deterministic exposure pair: two instruments for Alpha Bank Berlin -> same counterparty
UPDATE instruments
SET owner_entity_id = '11111111-1111-1111-1111-111111111102',
    counterparty_id = '22222222-2222-2222-2222-222222222202',
    notional_amount = 6000000.00,
    updated_at = now()
WHERE instrument_id IN (
  SELECT instrument_id FROM instruments ORDER BY instrument_id LIMIT 2
);

COMMIT;
