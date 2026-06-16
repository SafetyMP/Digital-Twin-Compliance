package events

import (
	"encoding/json"
	"testing"
)

func TestParseAuditPending(t *testing.T) {
	t.Parallel()

	raw := json.RawMessage(`{
		"entryType":"Alert",
		"correlationId":"corr-1",
		"subject":{"subjectId":"sub-1","subjectType":"ComplianceAlert"},
		"actor":{"actorId":"alert-service","actorType":"Service"},
		"action":"ComplianceAlertRaised",
		"payload":{"alertId":"a1","ruleCode":"INT-M001"},
		"metadata":{"regime":"Internal"}
	}`)
	pending, err := ParseAuditPending(raw)
	if err != nil {
		t.Fatalf("ParseAuditPending: %v", err)
	}
	if pending.EntryType != "Alert" || pending.Subject.SubjectID != "sub-1" {
		t.Fatalf("pending = %+v", pending)
	}
}

func TestParseEnvelope(t *testing.T) {
	t.Parallel()

	data := []byte(`{
		"eventId":"e1",
		"eventType":"AuditPending",
		"eventVersion":"1.0",
		"source":"alert-service",
		"correlationId":"corr-1",
		"timestamp":"2026-06-15T12:00:00Z",
		"idempotencyKey":"idem-1",
		"payload":{"entryType":"Alert","subject":{"subjectId":"s1","subjectType":"ComplianceAlert"}}
	}`)
	env, err := ParseEnvelope(data)
	if err != nil {
		t.Fatal(err)
	}
	if env.IdempotencyKey != "idem-1" || env.EventType != EventTypeAuditPending {
		t.Fatalf("env = %+v", env)
	}
}

func TestNewRecordedEnvelope(t *testing.T) {
	t.Parallel()

	entry := AuditEntry{
		EntryID:        "entry-1",
		EntryType:      "Alert",
		SequenceNumber: 1,
		RecordedAt:     "2026-06-15T12:00:00Z",
		CorrelationID:  "corr-1",
		Subject:        SubjectRef{SubjectID: "s1", SubjectType: "ComplianceAlert"},
		PayloadHash:    "sha256:abc",
		PreviousHash:   "",
		IdempotencyKey: "idem-1",
	}
	env, err := NewRecordedEnvelope(entry, "audit-service")
	if err != nil {
		t.Fatal(err)
	}
	if env.EventType != EventTypeAuditRecorded || env.Source != "audit-service" {
		t.Fatalf("env = %+v", env)
	}
}
