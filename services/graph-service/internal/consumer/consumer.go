package consumer

import (
	"context"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

type twinHandler interface {
	HandleTwinMessage(ctx context.Context, data []byte) error
}

type Runner struct {
	reader  *kafka.Reader
	handler twinHandler
	dlq     dlqPublisher
}

func NewRunner(brokers []string, group, topic, dlqTopic string, handler twinHandler) *Runner {
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
		if err := r.handler.HandleTwinMessage(ctx, msg.Value); err != nil {
			slog.Error("handle twin message", "error", err, "offset", msg.Offset)
			if !isPoison(err) {
				// Transient (e.g. Neo4j): do not commit — retry after brief backoff.
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(500 * time.Millisecond):
				}
				continue
			}
			if r.dlq == nil {
				slog.Warn("dlq disabled; committing poison twin message", "offset", msg.Offset)
			} else if dlqErr := r.dlq.PublishDLQ(ctx, msg, err); dlqErr != nil {
				slog.Error("publish twin dlq failed; not committing", "dlq_error", dlqErr, "offset", msg.Offset)
				continue
			} else {
				slog.Warn("routed poison twin message to dlq", "offset", msg.Offset)
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
