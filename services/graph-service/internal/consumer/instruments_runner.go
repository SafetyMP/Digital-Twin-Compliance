package consumer

import (
	"context"
	"log/slog"

	"github.com/segmentio/kafka-go"
)

type instrumentHandler interface {
	HandleInstrumentMessage(ctx context.Context, data []byte) error
}

type InstrumentsRunner struct {
	reader  *kafka.Reader
	handler instrumentHandler
}

func NewInstrumentsRunner(brokers []string, group, topic string, handler instrumentHandler) *InstrumentsRunner {
	return &InstrumentsRunner{
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

func (r *InstrumentsRunner) Run(ctx context.Context) error {
	for {
		msg, err := r.reader.FetchMessage(ctx)
		if err != nil {
			return err
		}
		if err := r.handler.HandleInstrumentMessage(ctx, msg.Value); err != nil {
			slog.Error("handle instrument cdc", "error", err, "offset", msg.Offset)
		}
		if err := r.reader.CommitMessages(ctx, msg); err != nil {
			slog.Error("commit instrument offset", "error", err)
		}
	}
}

func (r *InstrumentsRunner) Close() error {
	return r.reader.Close()
}
