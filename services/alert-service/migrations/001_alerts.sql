CREATE TABLE IF NOT EXISTS compliance_alerts (
  alert_id        UUID PRIMARY KEY,
  tenant_id       UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001',
  rule_code       TEXT NOT NULL,
  regime          TEXT NOT NULL,
  severity        TEXT NOT NULL,
  status          TEXT NOT NULL DEFAULT 'Open',
  persona_id      UUID NOT NULL,
  persona_type    TEXT NOT NULL,
  summary         TEXT NOT NULL,
  details         JSONB NOT NULL DEFAULT '{}',
  detected_at     TIMESTAMPTZ NOT NULL,
  acknowledged_at TIMESTAMPTZ,
  acknowledged_by TEXT,
  idempotency_key TEXT NOT NULL UNIQUE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS compliance_alerts_status_detected ON compliance_alerts (status, detected_at DESC);
CREATE INDEX IF NOT EXISTS compliance_alerts_persona ON compliance_alerts (persona_id);
