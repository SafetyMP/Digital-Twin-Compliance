-- Institution liquidity metrics for twin.state.updated / BASEL-M001
ALTER TABLE legal_entities ADD COLUMN IF NOT EXISTS lcr NUMERIC(10, 4);
ALTER TABLE legal_entities ADD COLUMN IF NOT EXISTS hqla NUMERIC(20, 2);
ALTER TABLE legal_entities ADD COLUMN IF NOT EXISTS net_cash_outflows_30d NUMERIC(20, 2);
ALTER TABLE legal_entities ADD COLUMN IF NOT EXISTS liquidity_currency CHAR(3) DEFAULT 'EUR';
