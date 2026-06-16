package store

import (
	"testing"
	"time"

	"github.com/digital-twin/platform/services/audit-service/internal/events"
)

func TestValidateEntryForIndexGenesis(t *testing.T) {
	t.Parallel()

	entry := events.AuditEntry{
		EntryID:        "11111111-1111-1111-1111-111111111101",
		EntryType:      "Alert",
		SequenceNumber: 1,
		RecordedAt:     time.Now().UTC().Format(time.RFC3339Nano),
		Subject:        events.SubjectRef{SubjectID: "22222222-2222-2222-2222-222222222202", SubjectType: "ComplianceAlert"},
		PayloadHash:    "sha256:abc",
		IdempotencyKey: "audit-alert-1",
	}
	if err := ValidateEntryForIndex(entry); err != nil {
		t.Fatalf("ValidateEntryForIndex: %v", err)
	}
}

func TestValidateEntryForIndexRejectsBadGenesis(t *testing.T) {
	t.Parallel()

	entry := events.AuditEntry{
		EntryID:        "11111111-1111-1111-1111-111111111101",
		EntryType:      "Alert",
		SequenceNumber: 1,
		PreviousHash:   "sha256:dead",
		RecordedAt:     time.Now().UTC().Format(time.RFC3339Nano),
		Subject:        events.SubjectRef{SubjectID: "t1", SubjectType: "TwinPersona"},
		PayloadHash:    "sha256:abc",
		IdempotencyKey: "k1",
	}
	if err := ValidateEntryForIndex(entry); err == nil {
		t.Fatal("expected genesis previousHash error")
	}
}
