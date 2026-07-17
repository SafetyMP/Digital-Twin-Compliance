package consumer

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

type Runner struct {
	reader  *kafka.Reader
	handler *Handler
	dlq     dlqPublisher
}

func NewRunner(brokers []string, group, topic, dlqTopic string, handler *Handler) *Runner {
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
	}
}

func isPoison(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	for _, needle := range []string{
		"invalid character",
		"unexpected end of json",
		"cannot unmarshal",
		"missing event",
		"unknown event",
	} {
		if strings.Contains(msg, needle) {
			return true
		}
	}
	return false
}

func (r *Runner) Run(ctx context.Context) error {
	for {
		msg, err := r.reader.FetchMessage(ctx)
		if err != nil {
			return err
		}
		if err := r.handler.HandleMessage(ctx, msg.Value); err != nil {
			slog.Error("handle alert message", "error", err, "offset", msg.Offset)
			if !isPoison(err) {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(500 * time.Millisecond):
				}
				continue
			}
			if r.dlq == nil {
				slog.Warn("dlq disabled; committing poison message to avoid consumer stall", "offset", msg.Offset)
			} else if dlqErr := r.dlq.PublishDLQ(ctx, msg, err); dlqErr != nil {
				slog.Error("publish dlq message; not committing",
					"dlq_error", dlqErr, "handle_error", err, "offset", msg.Offset)
				continue
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
