package consumer

import (
	"context"
	"log/slog"

	"github.com/segmentio/kafka-go"
)

type twinHandler interface {
	HandleTwinMessage(ctx context.Context, data []byte) error
}

type Runner struct {
	reader  *kafka.Reader
	handler twinHandler
}

func NewRunner(brokers []string, group, topic string, handler twinHandler) *Runner {
	return &Runner{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			GroupID:  group,
			Topic:    topic,
			MinBytes: 1,
			MaxBytes: 10e6,
		}),
		handler: handler,
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
		}
		if err := r.reader.CommitMessages(ctx, msg); err != nil {
			slog.Error("commit offset", "error", err)
		}
	}
}

func (r *Runner) Close() error {
	return r.reader.Close()
}
