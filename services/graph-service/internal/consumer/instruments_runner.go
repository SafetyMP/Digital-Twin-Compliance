package consumer

import (
	"context"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

type instrumentHandler interface {
	HandleInstrumentMessage(ctx context.Context, data []byte) error
}

type InstrumentsRunner struct {
	reader  *kafka.Reader
	handler instrumentHandler
	dlq     dlqPublisher
}

func NewInstrumentsRunner(brokers []string, group, topic, dlqTopic string, handler instrumentHandler) *InstrumentsRunner {
	var dlq dlqPublisher
	if dlqTopic != "" {
		dlq = newKafkaDLQ(brokers, dlqTopic)
	}
	return &InstrumentsRunner{
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

func (r *InstrumentsRunner) Run(ctx context.Context) error {
	for {
		msg, err := r.reader.FetchMessage(ctx)
		if err != nil {
			return err
		}
		if err := r.handler.HandleInstrumentMessage(ctx, msg.Value); err != nil {
			slog.Error("handle instrument cdc", "error", err, "offset", msg.Offset)
			if !isPoison(err) {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(500 * time.Millisecond):
				}
				continue
			}
			if r.dlq == nil {
				slog.Warn("dlq disabled; committing poison instrument message", "offset", msg.Offset)
			} else if dlqErr := r.dlq.PublishDLQ(ctx, msg, err); dlqErr != nil {
				slog.Error("publish instrument dlq failed; not committing", "dlq_error", dlqErr, "offset", msg.Offset)
				continue
			} else {
				slog.Warn("routed poison instrument message to dlq", "offset", msg.Offset)
			}
		}
		if err := r.reader.CommitMessages(ctx, msg); err != nil {
			slog.Error("commit instrument offset", "error", err)
		}
	}
}

func (r *InstrumentsRunner) Close() error {
	if dlq, ok := r.dlq.(*kafkaDLQ); ok {
		_ = dlq.Close()
	}
	return r.reader.Close()
}
