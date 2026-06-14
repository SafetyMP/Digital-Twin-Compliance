package consumer

import (
	"context"
	"testing"

	"github.com/digital-twin/platform/services/alert-service/internal/store"
)

type fakeAlertStore struct {
	saved   store.Alert
	created bool
	err     error
}

func (f *fakeAlertStore) UpsertAlert(ctx context.Context, input store.UpsertInput) (store.Alert, bool, error) {
	if f.err != nil {
		return store.Alert{}, false, f.err
	}
	f.created = true
	f.saved = store.Alert{
		AlertID:        input.AlertID,
		RuleCode:       input.RuleCode,
		IdempotencyKey: input.IdempotencyKey,
		DetectedAt:     input.DetectedAt,
	}
	return f.saved, true, nil
}

type fakeHub struct {
	msgType string
	alert   store.Alert
}

func (f *fakeHub) Broadcast(msgType string, alert store.Alert) {
	f.msgType = msgType
	f.alert = alert
}

func TestHandleMessage_IgnoresNonAlertEvents(t *testing.T) {
	t.Parallel()

	st := &fakeAlertStore{}
	h := &fakeHub{}
	handler := &Handler{store: st, hub: h}

	raw := []byte(`{"eventType":"OtherEvent","idempotencyKey":"k1","payload":{}}`)
	if err := handler.HandleMessage(context.Background(), raw); err != nil {
		t.Fatalf("HandleMessage: %v", err)
	}
	if st.created {
		t.Fatal("expected no upsert for non-alert event")
	}
}

func TestHandleMessage_InvalidPayloadReturnsError(t *testing.T) {
	t.Parallel()

	handler := &Handler{store: &fakeAlertStore{}, hub: &fakeHub{}}
	raw := []byte(`{"eventType":"ComplianceAlertRaised","idempotencyKey":"k1","payload":{"alertId":""}}`)
	if err := handler.HandleMessage(context.Background(), raw); err == nil {
		t.Fatal("expected error for invalid alert payload")
	}
}

func TestHandleMessage_UpsertsAndBroadcasts(t *testing.T) {
	t.Parallel()

	st := &fakeAlertStore{}
	h := &fakeHub{}
	handler := &Handler{store: st, hub: h}

	raw := []byte(`{
		"eventType":"ComplianceAlertRaised",
		"idempotencyKey":"idem-consumer-1",
		"payload":{
			"alertId":"550e8400-e29b-41d4-a716-446655440000",
			"ruleCode":"INT-M001",
			"regime":"Internal",
			"severity":"Warning",
			"status":"Open",
			"personaId":"660e8400-e29b-41d4-a716-446655440001",
			"personaType":"Account",
			"summary":"Velocity breach",
			"detectedAt":"2026-06-13T18:45:00Z",
			"details":{"count":"51"}
		}
	}`)
	if err := handler.HandleMessage(context.Background(), raw); err != nil {
		t.Fatalf("HandleMessage: %v", err)
	}
	if !st.created || st.saved.RuleCode != "INT-M001" {
		t.Fatalf("store = %+v created=%v", st.saved, st.created)
	}
	if h.msgType != "alert.raised" || h.alert.AlertID != st.saved.AlertID {
		t.Fatalf("broadcast = %q %+v", h.msgType, h.alert)
	}
}
