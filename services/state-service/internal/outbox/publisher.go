package outbox

import (
	"context"
	"log/slog"
	"time"

	"github.com/digital-twin/platform/services/state-service/internal/store"
	"github.com/segmentio/kafka-go"
)

type Publisher struct {
	store   *store.Store
	writer  *kafka.Writer
	source  string
	interval time.Duration
}

func NewPublisher(s *store.Store, brokers []string, source string, interval time.Duration) *Publisher {
	return &Publisher{
		store: s,
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Balancer: &kafka.Hash{},
		},
		source:   source,
		interval: interval,
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
	rows, err := p.store.FetchUnpublishedOutbox(ctx, 100)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if err := p.publishOne(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

func (p *Publisher) publishOne(ctx context.Context, row store.OutboxRow) error {
	_, body, err := buildTwinStateEnvelope(p.source, row)
	if err != nil {
		return err
	}

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Topic: row.Topic,
		Key:   []byte(row.PartitionKey),
		Value: body,
	})
	if err != nil {
		return err
	}

	return p.store.MarkOutboxPublished(ctx, row.ID)
}

func (p *Publisher) Close() error {
	return p.writer.Close()
}
