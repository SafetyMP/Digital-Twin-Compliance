package events

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNewEnvelope(t *testing.T) {
	t.Parallel()

	payload := EntityStateUpdated{
		PersonaID:      "11111111-1111-1111-1111-111111111101",
		PersonaType:    "Institution",
		SourceEntityID: "11111111-1111-1111-1111-111111111101",
		StateVersion:   1,
		ChangedFields:  []string{"name"},
		CurrentState:   json.RawMessage(`{"name":"Alpha Bank"}`),
		SourceSystem:   "core-banking",
	}
	envelope, err := NewEnvelope("EntityStateUpdated", "state-service", "cdc-1", payload)
	if err != nil {
		t.Fatalf("NewEnvelope: %v", err)
	}
	if envelope.EventType != "EntityStateUpdated" {
		t.Fatalf("EventType = %q", envelope.EventType)
	}
	if envelope.EventVersion != EventVersion {
		t.Fatalf("EventVersion = %q", envelope.EventVersion)
	}
	if envelope.Source != "state-service" {
		t.Fatalf("Source = %q", envelope.Source)
	}
	if envelope.IdempotencyKey != "cdc-1" {
		t.Fatalf("IdempotencyKey = %q", envelope.IdempotencyKey)
	}
	if envelope.EventID == "" || envelope.CorrelationID == "" || envelope.Timestamp == "" {
		t.Fatal("expected generated envelope metadata")
	}
	if !strings.Contains(envelope.Payload, "Alpha Bank") {
		t.Fatalf("payload = %q", envelope.Payload)
	}
}

func TestNewEnvelopeMarshalError(t *testing.T) {
	t.Parallel()

	_, err := NewEnvelope("Bad", "state-service", "key", make(chan int))
	if err == nil {
		t.Fatal("expected marshal error")
	}
}
