package outbox

import (
	"context"
	"log/slog"
	"time"

	"github.com/digital-twin/platform/services/state-service/internal/store"
	"github.com/segmentio/kafka-go"
)

type kafkaMessageWriter interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

type OutboxStore interface {
	FetchUnpublishedOutbox(ctx context.Context, limit int) ([]store.OutboxRow, error)
	MarkOutboxPublished(ctx context.Context, id int64) error
}

type Publisher struct {
	store     OutboxStore
	writer    kafkaMessageWriter
	source    string
	interval  time.Duration
	batchSize int
}

func NewPublisher(s OutboxStore, brokers []string, source string, interval, batchTimeout time.Duration, batchSize int) *Publisher {
	if batchSize <= 0 {
		batchSize = 100
	}
	if batchTimeout <= 0 {
		batchTimeout = 10 * time.Millisecond
	}
	return &Publisher{
		store: s,
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Balancer:     &kafka.Hash{},
			BatchSize:    batchSize,
			BatchTimeout: batchTimeout,
		},
		source:    source,
		interval:  interval,
		batchSize: batchSize,
	}
}

func (p *Publisher) Run(ctx context.Context) error {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := p.publishBatch(ctx); err != nil {
				slog.Error("outbox publish batch failed", "error", err)
			}
		}
	}
}

func (p *Publisher) publishBatch(ctx context.Context) error {
	rows, err := p.store.FetchUnpublishedOutbox(ctx, p.batchSize)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}

	type pending struct {
		id  int64
		msg kafka.Message
	}
	batch := make([]pending, 0, len(rows))

	for _, row := range rows {
		_, body, err := buildTwinStateEnvelope(p.source, row)
		if err != nil {
			slog.Error("outbox envelope build failed", "id", row.ID, "error", err)
			continue
		}
		batch = append(batch, pending{
			id: row.ID,
			msg: kafka.Message{
				Topic: row.Topic,
				Key:   []byte(row.PartitionKey),
				Value: body,
			},
		})
	}
	if len(batch) == 0 {
		return nil
	}

	msgs := make([]kafka.Message, len(batch))
	for i, item := range batch {
		msgs[i] = item.msg
	}
	if err := p.writer.WriteMessages(ctx, msgs...); err != nil {
		return err
	}
	for _, item := range batch {
		if err := p.store.MarkOutboxPublished(ctx, item.id); err != nil {
			return err
		}
	}
	return nil
}

func (p *Publisher) Close() error {
	return p.writer.Close()
}
