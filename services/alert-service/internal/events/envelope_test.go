package events

import "testing"

func TestParseFlinkBaselAlert(t *testing.T) {
	raw := readContractFile(t, "compliance.alerts/basel-alert-raised.envelope.json")

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
