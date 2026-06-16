ALTER TABLE compliance_alerts
  ADD COLUMN IF NOT EXISTS evidence_ref TEXT;

CREATE INDEX IF NOT EXISTS compliance_alerts_evidence_ref ON compliance_alerts (evidence_ref)
  WHERE evidence_ref IS NOT NULL;
