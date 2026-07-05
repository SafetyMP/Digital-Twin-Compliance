-- Phase 4: emit Debezium CDC for all instruments so graph-service can materialize exposure edges.
UPDATE instruments
SET updated_at = NOW()
WHERE owner_entity_id IS NOT NULL AND counterparty_id IS NOT NULL;
