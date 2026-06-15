-- Phase 2 seed patch: exposure limit breach + institution liquidity baseline
BEGIN;

-- Assign owner_entity_id from random institution for existing instruments
UPDATE instruments i
SET owner_entity_id = le.entity_id
FROM (
  SELECT entity_id, row_number() OVER (ORDER BY entity_id) AS rn
  FROM legal_entities
) le
WHERE i.owner_entity_id IS NULL
  AND le.rn = (abs(hashtext(i.instrument_id::text)) % 10) + 1;

-- Create exposure breach: Alpha Bank Berlin holds >10M EUR to same counterparty
UPDATE instruments
SET owner_entity_id = '11111111-1111-1111-1111-111111111102',
    counterparty_id = '22222222-2222-2222-2222-222222222202',
    notional_amount = 6000000.00,
    updated_at = now()
WHERE instrument_id IN (
  SELECT instrument_id FROM instruments
  WHERE owner_entity_id = '11111111-1111-1111-1111-111111111102'
  LIMIT 2
);

COMMIT;
