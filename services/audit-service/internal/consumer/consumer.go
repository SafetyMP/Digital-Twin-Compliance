package consumer

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

type Runner struct {
	reader  *kafka.Reader
	handler *Handler
	dlq     dlqPublisher
	store   dlqStore
}

type dlqStore interface {
	InsertDLQ(ctx context.Context, idempotencyKey, topic string, partition int, offset int64, errMsg string, payload json.RawMessage) error
}

func NewRunner(brokers []string, group, topic, dlqTopic string, handler *Handler, st dlqStore) *Runner {
	var dlq dlqPublisher
	if dlqTopic != "" {
		dlq = newKafkaDLQ(brokers, dlqTopic)
	}
	return &Runner{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			GroupID:  group,
			Topic:    topic,
			MinBytes: 1,
			MaxBytes: 10e6,
		}),
		handler: handler,
		dlq:     dlq,
		store:   st,
	}
}

func (r *Runner) Run(ctx context.Context) error {
	for {
		msg, err := r.reader.FetchMessage(ctx)
		if err != nil {
			return err
		}
		if err := r.handler.HandleMessage(ctx, msg.Value); err != nil {
			slog.Error("handle audit message", "error", err, "offset", msg.Offset)
			idempotencyKey := extractIdempotencyKey(msg.Value)
			if r.store != nil {
				if dlqErr := r.store.InsertDLQ(ctx, idempotencyKey, msg.Topic, msg.Partition, msg.Offset, err.Error(), msg.Value); dlqErr != nil {
					slog.Error("persist audit dlq row", "error", dlqErr)
				}
			}
			if r.dlq == nil {
				slog.Warn("dlq disabled; committing poison message to avoid consumer stall", "offset", msg.Offset)
			} else if dlqErr := r.dlq.PublishDLQ(ctx, msg, err); dlqErr != nil {
				slog.Error("publish dlq message; committing offset to avoid consumer stall",
					"dlq_error", dlqErr, "handle_error", err, "offset", msg.Offset)
			} else {
				slog.Warn("routed poison message to dlq", "offset", msg.Offset)
			}
		}
		if err := r.reader.CommitMessages(ctx, msg); err != nil {
			slog.Error("commit offset", "error", err)
		}
	}
}

func (r *Runner) Close() error {
	if dlq, ok := r.dlq.(*kafkaDLQ); ok {
		_ = dlq.Close()
	}
	return r.reader.Close()
}

func extractIdempotencyKey(data []byte) string {
	var env struct {
		IdempotencyKey string `json:"idempotencyKey"`
	}
	if err := json.Unmarshal(data, &env); err != nil {
		return ""
	}
	return env.IdempotencyKey
}

type dlqPublisher interface {
	PublishDLQ(ctx context.Context, msg kafka.Message, handleErr error) error
}

type kafkaDLQ struct {
	writer *kafka.Writer
}

func newKafkaDLQ(brokers []string, topic string) *kafkaDLQ {
	return &kafkaDLQ{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			BatchTimeout: 10 * time.Millisecond,
		},
	}
}

func (k *kafkaDLQ) PublishDLQ(ctx context.Context, msg kafka.Message, handleErr error) error {
	if k == nil || k.writer == nil {
		return fmt.Errorf("dlq publisher not configured")
	}
	body, err := json.Marshal(map[string]any{
		"originalTopic": msg.Topic,
		"partition":     msg.Partition,
		"offset":        msg.Offset,
		"error":         handleErr.Error(),
		"payloadBase64": base64.StdEncoding.EncodeToString(msg.Value),
		"timestamp":     time.Now().UTC().Format(time.RFC3339Nano),
	})
	if err != nil {
		return err
	}
	return k.writer.WriteMessages(ctx, kafka.Message{Value: body})
}

func (k *kafkaDLQ) Close() error {
	if k == nil || k.writer == nil {
		return nil
	}
	return k.writer.Close()
}
