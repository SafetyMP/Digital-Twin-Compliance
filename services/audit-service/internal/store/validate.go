package store

import (
	"fmt"
	"time"

	"github.com/digital-twin/platform/services/audit-service/internal/events"
	"github.com/google/uuid"
)

// ValidateEntryForIndex checks index-row constraints before writing to immudb.
func ValidateEntryForIndex(entry events.AuditEntry) error {
	if entry.EntryID == "" {
		return fmt.Errorf("entryId required")
	}
	if _, err := uuid.Parse(entry.EntryID); err != nil {
		return fmt.Errorf("entryId must be UUID: %w", err)
	}
	if entry.EntryType == "" {
		return fmt.Errorf("entryType required")
	}
	if entry.Subject.SubjectID == "" {
		return fmt.Errorf("subjectId required")
	}
	if entry.Subject.SubjectType == "" {
		return fmt.Errorf("subjectType required")
	}
	if entry.IdempotencyKey == "" {
		return fmt.Errorf("idempotencyKey required")
	}
	if entry.PayloadHash == "" {
		return fmt.Errorf("payloadHash required")
	}
	if entry.SequenceNumber < 1 {
		return fmt.Errorf("sequenceNumber must be >= 1")
	}
	if entry.SequenceNumber == 1 && entry.PreviousHash != "" {
		return fmt.Errorf("genesis entry must have empty previousHash")
	}
	if entry.SequenceNumber > 1 && entry.PreviousHash == "" {
		return fmt.Errorf("previousHash required for sequence > 1")
	}
	if _, err := parseRecordedAt(entry.RecordedAt); err != nil {
		return fmt.Errorf("recordedAt: %w", err)
	}
	return nil
}

func parseRecordedAt(raw string) (time.Time, error) {
	if raw == "" {
		return time.Time{}, fmt.Errorf("missing value")
	}
	if ts, err := time.Parse(time.RFC3339Nano, raw); err == nil {
		return ts, nil
	}
	return time.Parse(time.RFC3339, raw)
}
