CREATE TABLE twin_personas (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001'
);

-- simplified for eval fixture
ALTER TABLE twin_personas DROP COLUMN tenant_id;
