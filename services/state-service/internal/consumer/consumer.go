package consumer

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/digital-twin/platform/services/state-service/internal/store"
	"github.com/segmentio/kafka-go"
)

type Handler struct {
	store  *store.Store
	source string
}

func NewHandler(s *store.Store, source string) *Handler {
	return &Handler{store: s, source: source}
}

func (h *Handler) HandleMessage(ctx context.Context, msg kafka.Message) error {
	payload, err := ParseDebeziumMessage(msg.Value)
	if err != nil {
		return err
	}
	return h.processPayload(ctx, payload)
}

func (h *Handler) processPayload(ctx context.Context, p DebeziumPayload) error {
	input, parentID, err := MapDebeziumToCDCInput(p)
	if err != nil {
		return err
	}
	if input.PersonaID == "" {
		return nil
	}

	if input.SourceTable == "legal_entities" {
		if err := h.store.ValidateInstitutionDepth(ctx, input.PersonaID, parentID); err != nil {
			slog.Warn("hierarchy validation failed", "entity_id", input.PersonaID, "error", err)
			return err
		}
	}

	_, err = h.store.ApplyCDCEvent(ctx, input)
	return err
}

type tableMap struct {
	pkColumn    string
	personaType string
}

var tableMapping = map[string]tableMap{
	"legal_entities": {pkColumn: "entity_id", personaType: "Institution"},
	"accounts":       {pkColumn: "account_id", personaType: "Account"},
	"instruments":    {pkColumn: "instrument_id", personaType: "Instrument"},
}

func stringField(row map[string]any, key string) string {
	v, ok := row[key]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatInt(int64(t), 10)
	case map[string]any:
		if s, ok := t["string"].(string); ok {
			return s
		}
		return fmt.Sprint(v)
	default:
		return fmt.Sprint(v)
	}
}

func parseTimestamp(v any) time.Time {
	if v == nil {
		return time.Now().UTC()
	}
	switch t := v.(type) {
	case string:
		for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05.999999-07"} {
			if parsed, err := time.Parse(layout, t); err == nil {
				return parsed.UTC()
			}
		}
		if strings.Contains(t, "T") {
			if parsed, err := time.Parse("2006-01-02T15:04:05.999999Z", t); err == nil {
				return parsed.UTC()
			}
		}
	case float64:
		return time.UnixMicro(int64(t)).UTC()
	}
	return time.Now().UTC()
}

type Runner struct {
	reader  *kafka.Reader
	handler *Handler
	dlq     dlqPublisher
}

func NewRunner(brokers []string, groupID string, topics []string, dlqTopic string, handler *Handler) *Runner {
	var dlq dlqPublisher
	if dlqTopic != "" {
		dlq = newKafkaDLQ(brokers, dlqTopic)
	}
	return &Runner{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:     brokers,
			GroupID:     groupID,
			GroupTopics: topics,
			MinBytes:    1,
			MaxBytes:    10e6,
		}),
		handler: handler,
		dlq:     dlq,
	}
}

func (r *Runner) Run(ctx context.Context) error {
	for {
		msg, err := r.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		if err := r.handler.HandleMessage(ctx, msg); err != nil {
			slog.Error("handle message failed", "topic", msg.Topic, "error", err, "offset", msg.Offset)
			if r.dlq == nil {
				slog.Warn("dlq disabled; committing poison message to avoid consumer stall", "offset", msg.Offset)
			} else if dlqErr := r.dlq.PublishDLQ(ctx, msg, err); dlqErr != nil {
				slog.Error("publish dlq message", "error", dlqErr, "offset", msg.Offset)
				continue
			} else {
				slog.Warn("routed poison message to dlq", "offset", msg.Offset)
			}
		}
		if err := r.reader.CommitMessages(ctx, msg); err != nil {
			return err
		}
	}
}

func (r *Runner) Close() error {
	if dlq, ok := r.dlq.(*kafkaDLQ); ok {
		_ = dlq.Close()
	}
	return r.reader.Close()
}
