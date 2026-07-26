CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS taxonomy_maps (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL,
  report_type TEXT NOT NULL,
  source_code TEXT NOT NULL,
  target_code TEXT NOT NULL,
  version TEXT NOT NULL,
  effective_from DATE NOT NULL,
  effective_to DATE,
  UNIQUE (tenant_id, report_type, source_code, version)
);

CREATE TABLE IF NOT EXISTS reports (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  report_type TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('draft', 'validated', 'submitted')),
  taxonomy_version TEXT NOT NULL,
  artifact_xml TEXT,
  object_key TEXT,
  evidence_ref TEXT,
  correlation_id TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS reports_status_idx ON reports (tenant_id, status);
