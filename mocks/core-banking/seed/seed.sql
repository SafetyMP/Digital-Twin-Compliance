-- Phase 1 seed: 10 institutions, 100 accounts, 500 instruments
-- Hierarchy: 3 parent groups with subsidiaries and sub-subsidiaries (max depth 3)

BEGIN;

-- Parent groups
INSERT INTO legal_entities (entity_id, legal_name, lei, entity_type, jurisdiction, parent_entity_id, consolidation_scope) VALUES
  ('11111111-1111-1111-1111-111111111101', 'Group Alpha Holdings', 'LEI000000000000000001', 'Bank', 'DE', NULL, 'Group'),
  ('22222222-2222-2222-2222-222222222201', 'Group Beta Financial', 'LEI000000000000000002', 'Bank', 'FR', NULL, 'Group'),
  ('33333333-3333-3333-3333-333333333301', 'Group Gamma Capital', 'LEI000000000000000003', 'Bank', 'NL', NULL, 'Group');

-- Group Alpha subsidiaries
INSERT INTO legal_entities (entity_id, legal_name, lei, entity_type, jurisdiction, parent_entity_id, consolidation_scope) VALUES
  ('11111111-1111-1111-1111-111111111102', 'Alpha Bank Berlin', 'LEI000000000000000011', 'Bank', 'DE', '11111111-1111-1111-1111-111111111101', 'Group'),
  ('11111111-1111-1111-1111-111111111103', 'Alpha Bank Munich', 'LEI000000000000000012', 'Bank', 'DE', '11111111-1111-1111-1111-111111111101', 'Group');

-- Group Alpha sub-subsidiary
INSERT INTO legal_entities (entity_id, legal_name, lei, entity_type, jurisdiction, parent_entity_id, consolidation_scope) VALUES
  ('11111111-1111-1111-1111-111111111104', 'Alpha Branch Kreuzberg', 'LEI000000000000000013', 'Bank', 'DE', '11111111-1111-1111-1111-111111111102', 'Solo');

-- Group Beta subsidiaries
INSERT INTO legal_entities (entity_id, legal_name, lei, entity_type, jurisdiction, parent_entity_id, consolidation_scope) VALUES
  ('22222222-2222-2222-2222-222222222202', 'Beta Bank Paris', 'LEI000000000000000021', 'Bank', 'FR', '22222222-2222-2222-2222-222222222201', 'Group'),
  ('22222222-2222-2222-2222-222222222203', 'Beta Asset Management', 'LEI000000000000000022', 'Fund', 'FR', '22222222-2222-2222-2222-222222222201', 'Group');

-- Group Beta sub-subsidiary
INSERT INTO legal_entities (entity_id, legal_name, lei, entity_type, jurisdiction, parent_entity_id, consolidation_scope) VALUES
  ('22222222-2222-2222-2222-222222222204', 'Beta Branch Lyon', 'LEI000000000000000023', 'Bank', 'FR', '22222222-2222-2222-2222-222222222202', 'Solo');

-- Group Gamma subsidiaries
INSERT INTO legal_entities (entity_id, legal_name, lei, entity_type, jurisdiction, parent_entity_id, consolidation_scope) VALUES
  ('33333333-3333-3333-3333-333333333302', 'Gamma Bank Amsterdam', 'LEI000000000000000031', 'Bank', 'NL', '33333333-3333-3333-3333-333333333301', 'Group'),
  ('33333333-3333-3333-3333-333333333303', 'Gamma SPV Treasury', 'LEI000000000000000032', 'SPV', 'NL', '33333333-3333-3333-3333-333333333301', 'Group');

-- Standalone institution (10th)
INSERT INTO legal_entities (entity_id, legal_name, lei, entity_type, jurisdiction, parent_entity_id, consolidation_scope) VALUES
  ('44444444-4444-4444-4444-444444444401', 'Delta Independent Bank', 'LEI000000000000000041', 'Bank', 'IE', NULL, 'Solo');

-- 100 accounts: 10 per institution
DO $$
DECLARE
  inst RECORD;
  i INT;
  acct_types TEXT[] := ARRAY['Customer','Nostro','Vostro','Suspense','Regulatory'];
  currencies TEXT[] := ARRAY['EUR','USD','GBP'];
BEGIN
  FOR inst IN SELECT entity_id FROM legal_entities ORDER BY legal_name LOOP
    FOR i IN 1..10 LOOP
      INSERT INTO accounts (account_number, account_type, currency, owner_entity_id, status)
      VALUES (
        'ACC-' || replace(inst.entity_id::text, '-', '') || '-' || lpad(i::text, 3, '0'),
        acct_types[1 + (i % array_length(acct_types, 1))],
        currencies[1 + (i % array_length(currencies, 1))],
        inst.entity_id,
        'Active'
      );
    END LOOP;
  END LOOP;
END $$;

-- 500 instruments
DO $$
DECLARE
  inst_types TEXT[] := ARRAY['Loan','Bond','Deposit','Derivative','Repo'];
  i INT;
  owner UUID;
BEGIN
  FOR i IN 1..500 LOOP
    SELECT entity_id INTO owner FROM legal_entities ORDER BY random() LIMIT 1;
    INSERT INTO instruments (isin, instrument_type, counterparty_id, notional_amount, currency, maturity_date, regulatory_class)
    VALUES (
      'XS' || lpad(i::text, 10, '0'),
      inst_types[1 + (i % array_length(inst_types, 1))],
      owner,
      (random() * 10000000 + 10000)::numeric(20,2),
      CASE (i % 3) WHEN 0 THEN 'EUR' WHEN 1 THEN 'USD' ELSE 'GBP' END,
      (CURRENT_DATE + (random() * 3650)::int),
      'F0610'
    );
  END LOOP;
END $$;

COMMIT;
