package consumer

import (
	"context"
	"testing"
	"time"

	"github.com/digital-twin/platform/services/alert-service/internal/events"
)

func TestHandleMessage_FlinkVelocityAlertEnvelope(t *testing.T) {
	t.Parallel()

	st := &fakeAlertStore{}
	h := &fakeHub{}
	handler := &Handler{store: st, hub: h}

	accountID := "23cced6e-1907-4f3f-b70b-1680418a9dd7"
	raw := []byte(`{
		"eventId":"11111111-1111-1111-1111-111111111111",
		"eventType":"ComplianceAlertRaised",
		"eventVersion":"1.0",
		"source":"flink-compliance-cep",
		"correlationId":"22222222-2222-2222-2222-222222222222",
		"causationId":null,
		"timestamp":"2026-06-15T01:30:36.123456789Z",
		"idempotencyKey":"INT-M001-` + accountID + `-2026-06-15T01:00:00Z",
		"payload":{
			"alertId":"550e8400-e29b-41d4-a716-446655440000",
			"ruleCode":"INT-M001",
			"regime":"Internal",
			"severity":"Warning",
			"status":"Open",
			"personaId":"` + accountID + `",
			"personaType":"Account",
			"summary":"Transaction velocity exceeded threshold",
			"detectedAt":"2026-06-15T01:30:36.123456789Z",
			"details":{"count":"51","threshold":"50","window":"1h"}
		}
	}`)

	if err := handler.HandleMessage(context.Background(), raw); err != nil {
		t.Fatalf("HandleMessage: %v", err)
	}
	if !st.created || st.saved.RuleCode != "INT-M001" {
		t.Fatalf("store = %+v created=%v", st.saved, st.created)
	}
}

func TestParseDetectedAt_JavaInstantFormats(t *testing.T) {
	t.Parallel()

	cases := []string{
		"2026-06-15T01:30:36.123456789Z",
		"2026-06-15T01:00:00Z",
		"2026-06-15T01:30:36Z",
	}
	for _, tc := range cases {
		if _, err := events.ParseDetectedAt(tc); err != nil {
			t.Fatalf("ParseDetectedAt(%q): %v", tc, err)
		}
	}
}

func TestParseFlinkEnvelopePayloadObject(t *testing.T) {
	t.Parallel()

	raw := []byte(`{"eventType":"ComplianceAlertRaised","idempotencyKey":"k1","payload":{"alertId":"550e8400-e29b-41d4-a716-446655440000","ruleCode":"INT-M001","regime":"Internal","severity":"Warning","status":"Open","personaId":"23cced6e-1907-4f3f-b70b-1680418a9dd7","personaType":"Account","summary":"x","detectedAt":"2026-06-15T01:30:36.123456789Z","details":{"count":"51"}}}`)
	env, err := events.ParseEnvelope(raw)
	if err != nil {
		t.Fatalf("ParseEnvelope: %v", err)
	}
	alert, err := events.ParseAlertPayload(env.Payload)
	if err != nil {
		t.Fatalf("ParseAlertPayload: %v", err)
	}
	if _, err := events.ParseDetectedAt(alert.DetectedAt); err != nil {
		t.Fatalf("ParseDetectedAt: %v", err)
	}
	if alert.RuleCode != "INT-M001" {
		t.Fatalf("ruleCode = %q", alert.RuleCode)
	}
	_ = time.Now()
}
