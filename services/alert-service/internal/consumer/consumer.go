package consumer

import (
	"context"
	"log/slog"

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

func (r *Runner) Run(ctx context.Context) error {
	for {
		msg, err := r.reader.FetchMessage(ctx)
		if err != nil {
			return err
		}
		if err := r.handler.HandleMessage(ctx, msg.Value); err != nil {
			slog.Error("handle alert message", "error", err, "offset", msg.Offset)
			if r.dlq == nil {
				slog.Warn("dlq disabled; committing poison message to avoid consumer stall", "offset", msg.Offset)
				if err := r.reader.CommitMessages(ctx, msg); err != nil {
					slog.Error("commit offset", "error", err)
				}
				continue
			}
		}
			if dlqErr := r.dlq.PublishDLQ(ctx, msg, err); dlqErr != nil {
				slog.Error("publish dlq message", "error", dlqErr, "offset", msg.Offset)
				continue
			}
			slog.Warn("routed poison message to dlq", "offset", msg.Offset)
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
