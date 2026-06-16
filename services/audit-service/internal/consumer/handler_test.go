package consumer

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/digital-twin/platform/services/audit-service/internal/events"
	"github.com/digital-twin/platform/services/audit-service/internal/immudb"
)

type fakeLedger struct {
	head    immudb.HeadState
	entries []events.AuditEntry
}

func (f *fakeLedger) Ping(ctx context.Context) error { return nil }

func (f *fakeLedger) GetHead(ctx context.Context) (immudb.HeadState, error) {
	return f.head, nil
}

func (f *fakeLedger) AppendEntry(ctx context.Context, entry events.AuditEntry) error {
	f.entries = append(f.entries, entry)
	f.head = immudb.HeadState{LastSequence: entry.SequenceNumber, LastPayloadHash: entry.PayloadHash}
	return nil
}

func (f *fakeLedger) GetEntry(ctx context.Context, entryID string) (events.AuditEntry, error) {
	for _, e := range f.entries {
		if e.EntryID == entryID {
			return e, nil
		}
	}
	return events.AuditEntry{}, context.Canceled
}

type fakeStore struct {
	processed map[string]string
	recorded  []events.AuditEntry
}

func (f *fakeStore) IsProcessed(ctx context.Context, key string) (string, bool, error) {
	id, ok := f.processed[key]
	return id, ok, nil
}

func (f *fakeStore) RecordEntry(ctx context.Context, entry events.AuditEntry, ruleCode string) error {
	f.recorded = append(f.recorded, entry)
	f.processed[entry.IdempotencyKey] = entry.EntryID
	return nil
}

func (f *fakeStore) InsertDLQ(ctx context.Context, idempotencyKey, topic string, partition int, offset int64, errMsg string, payload json.RawMessage) error {
	return nil
}

type fakePublisher struct {
	entries []events.AuditEntry
}

func (f *fakePublisher) PublishRecorded(ctx context.Context, entry events.AuditEntry) error {
	f.entries = append(f.entries, entry)
	return nil
}

func buildPendingEnvelope(idempotencyKey string) []byte {
	body, _ := json.Marshal(events.Envelope{
		EventID:        "evt-1",
		EventType:      events.EventTypeAuditPending,
		EventVersion:   "1.0",
		Source:         "alert-service",
		CorrelationID:  "corr-1",
		Timestamp:      "2026-06-15T12:00:00Z",
		IdempotencyKey: idempotencyKey,
		Payload: json.RawMessage(`{
			"entryType":"Alert",
			"correlationId":"corr-1",
			"subject":{"subjectId":"sub-1","subjectType":"ComplianceAlert"},
			"actor":{"actorId":"alert-service","actorType":"Service"},
			"action":"ComplianceAlertRaised",
			"payload":{"alertId":"a1","ruleCode":"INT-M001"},
			"metadata":{"regime":"Internal"}
		}`),
	})
	return body
}

func TestHandlerIdempotency(t *testing.T) {
	t.Parallel()

	ledger := &fakeLedger{}
	st := &fakeStore{processed: map[string]string{}}
	pub := &fakePublisher{}
	h := NewHandler(ledger, st, pub, "audit-service")

	data := buildPendingEnvelope("idem-1")
	if err := h.HandleMessage(context.Background(), data); err != nil {
		t.Fatalf("first handle: %v", err)
	}
	if len(ledger.entries) != 1 || len(pub.entries) != 1 {
		t.Fatalf("expected one entry, ledger=%d pub=%d", len(ledger.entries), len(pub.entries))
	}

	if err := h.HandleMessage(context.Background(), data); err != nil {
		t.Fatalf("second handle: %v", err)
	}
	if len(ledger.entries) != 1 || len(pub.entries) != 1 {
		t.Fatalf("duplicate should be skipped, ledger=%d pub=%d", len(ledger.entries), len(pub.entries))
	}
}

func TestHandlerChainsPreviousHash(t *testing.T) {
	t.Parallel()

	ledger := &fakeLedger{head: immudb.HeadState{LastSequence: 1, LastPayloadHash: "sha256:first"}}
	st := &fakeStore{processed: map[string]string{}}
	pub := &fakePublisher{}
	h := NewHandler(ledger, st, pub, "audit-service")

	if err := h.HandleMessage(context.Background(), buildPendingEnvelope("idem-2")); err != nil {
		t.Fatal(err)
	}
	if ledger.entries[0].PreviousHash != "sha256:first" {
		t.Fatalf("previousHash = %q", ledger.entries[0].PreviousHash)
	}
	if ledger.entries[0].SequenceNumber != 2 {
		t.Fatalf("sequence = %d", ledger.entries[0].SequenceNumber)
	}
}
