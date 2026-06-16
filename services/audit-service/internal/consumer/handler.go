package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/digital-twin/platform/services/audit-service/internal/chain"
	"github.com/digital-twin/platform/services/audit-service/internal/events"
	"github.com/digital-twin/platform/services/audit-service/internal/immudb"
	"github.com/digital-twin/platform/services/audit-service/internal/store"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type ledger interface {
	GetHead(ctx context.Context) (immudb.HeadState, error)
	AppendEntry(ctx context.Context, entry events.AuditEntry) error
}

type entryStore interface {
	IsProcessed(ctx context.Context, idempotencyKey string) (string, bool, error)
	RecordEntry(ctx context.Context, entry events.AuditEntry, ruleCode string) error
	InsertDLQ(ctx context.Context, idempotencyKey, topic string, partition int, offset int64, errMsg string, payload json.RawMessage) error
}

type recordedPublisher interface {
	PublishRecorded(ctx context.Context, entry events.AuditEntry) error
}

type Handler struct {
	ledger    ledger
	store     entryStore
	publisher recordedPublisher
	source    string
}

func NewHandler(ledger ledger, st entryStore, pub recordedPublisher, source string) *Handler {
	return &Handler{ledger: ledger, store: st, publisher: pub, source: source}
}

func (h *Handler) HandleMessage(ctx context.Context, data []byte) error {
	env, err := events.ParseEnvelope(data)
	if err != nil {
		return err
	}
	if env.EventType != events.EventTypeAuditPending {
		return nil
	}
	if env.IdempotencyKey == "" {
		return fmt.Errorf("missing idempotencyKey")
	}

	if _, processed, err := h.store.IsProcessed(ctx, env.IdempotencyKey); err != nil {
		return err
	} else if processed {
		return nil
	}

	pending, err := events.ParseAuditPending(env.Payload)
	if err != nil {
		return err
	}

	head, err := h.ledger.GetHead(ctx)
	if err != nil {
		return err
	}

	payloadHash, err := chain.PayloadHash(pending.Payload, pending.Metadata)
	if err != nil {
		return err
	}

	previousHash := ""
	sequence := int64(1)
	if head.LastSequence > 0 {
		previousHash = head.LastPayloadHash
		sequence = head.LastSequence + 1
	}

	correlationID := pending.CorrelationID
	if correlationID == "" {
		correlationID = env.CorrelationID
	}

	entry := events.AuditEntry{
		EntryID:        uuid.NewString(),
		EntryType:      pending.EntryType,
		SequenceNumber: sequence,
		RecordedAt:     time.Now().UTC().Format(time.RFC3339Nano),
		CorrelationID:  correlationID,
		Subject:        pending.Subject,
		Actor:          pending.Actor,
		Action:         pending.Action,
		Payload:        pending.Payload,
		PayloadHash:    payloadHash,
		PreviousHash:   previousHash,
		Metadata:       pending.Metadata,
		IdempotencyKey: env.IdempotencyKey,
	}

	if err := store.ValidateEntryForIndex(entry); err != nil {
		return fmt.Errorf("validate audit entry: %w", err)
	}

	if err := h.ledger.AppendEntry(ctx, entry); err != nil {
		return err
	}

	ruleCode := store.ExtractRuleCode(pending.Payload)
	if err := h.store.RecordEntry(ctx, entry, ruleCode); err != nil {
		return err
	}

	return h.publisher.PublishRecorded(ctx, entry)
}

type RecordedProducer struct {
	writer *kafka.Writer
	source string
}

func NewRecordedProducer(brokers []string, topic, source string) *RecordedProducer {
	return &RecordedProducer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			BatchTimeout: 10 * time.Millisecond,
		},
		source: source,
	}
}

func (p *RecordedProducer) PublishRecorded(ctx context.Context, entry events.AuditEntry) error {
	env, err := events.NewRecordedEnvelope(entry, p.source)
	if err != nil {
		return err
	}
	body, err := json.Marshal(env)
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{Key: []byte(entry.EntryID), Value: body})
}

func (p *RecordedProducer) Close() error {
	if p == nil || p.writer == nil {
		return nil
	}
	return p.writer.Close()
}
