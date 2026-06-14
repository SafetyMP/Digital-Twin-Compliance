package events

import "testing"

func TestParseFlinkBaselAlert(t *testing.T) {
	raw := []byte(`{"eventId":"27f3c9e9-9ca7-4de9-86d4-8a3f5ad5616b","eventType":"ComplianceAlertRaised","eventVersion":"1.0","source":"flink-compliance-cep","correlationId":"0f088a16-7cd3-41ad-8668-09b38235a49e","causationId":null,"timestamp":"2026-06-14T01:34:56.093806460Z","idempotencyKey":"BASEL-M001-44444444-4444-4444-4444-444444444401-95","payload":{"alertId":"79ca894e-6347-474b-b602-e864c510c8e2","ruleCode":"BASEL-M001","regime":"Basel","severity":"Critical","status":"Open","personaId":"44444444-4444-4444-4444-444444444401","personaType":"Institution","summary":"LCR below minimum threshold","detectedAt":"2026-06-14T01:34:56.093806460Z","details":{"lcr":"0.95","metric":"lcr","threshold":"1.0"}}}`)

	env, err := ParseEnvelope(raw)
	if err != nil {
		t.Fatalf("ParseEnvelope: %v", err)
	}
	alert, err := ParseAlertPayload(env.Payload)
	if err != nil {
		t.Fatalf("ParseAlertPayload: %v", err)
	}
	if _, err := ParseDetectedAt(alert.DetectedAt); err != nil {
		t.Fatalf("ParseDetectedAt: %v", err)
	}
	if alert.RuleCode != "BASEL-M001" {
		t.Fatalf("ruleCode=%s", alert.RuleCode)
	}
}
