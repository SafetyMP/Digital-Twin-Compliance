-- Baseline CET1 ratios for graph/simulation (COREP path).
UPDATE legal_entities
SET cet1_ratio = 0.14
WHERE cet1_ratio IS NULL;

UPDATE legal_entities
SET cet1_ratio = 0.11
WHERE entity_id = '44444444-4444-4444-4444-444444444401';
