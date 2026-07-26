-- Institution capital metrics for twin.state.updated / COREP CET1
ALTER TABLE legal_entities ADD COLUMN IF NOT EXISTS cet1_ratio NUMERIC(10, 4);
