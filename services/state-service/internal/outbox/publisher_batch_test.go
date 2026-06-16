package outbox

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/digital-twin/platform/services/state-service/internal/store"
	"github.com/segmentio/kafka-go"
)

type fakeOutboxStore struct {
	rows []store.OutboxRow
	err  error
}

func (f *fakeOutboxStore) FetchUnpublishedOutbox(ctx context.Context, limit int) ([]store.OutboxRow, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.rows, nil
}

func (f *fakeOutboxStore) MarkOutboxPublished(ctx context.Context, id int64) error {
	return nil
}

type fakeKafkaWriter struct {
	err       error
	batchLens []int
}

func (f *fakeKafkaWriter) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	f.batchLens = append(f.batchLens, len(msgs))
	return f.err
}

func (f *fakeKafkaWriter) Close() error { return nil }

func TestPublishBatch_Empty(t *testing.T) {
	t.Parallel()

	p := &Publisher{store: &fakeOutboxStore{}, writer: &fakeKafkaWriter{}, source: "state-service"}
	if err := p.publishBatch(context.Background()); err != nil {
		t.Fatalf("publishBatch: %v", err)
	}
}

func TestPublishBatch_FetchError(t *testing.T) {
	t.Parallel()

	p := &Publisher{
		store:  &fakeOutboxStore{err: errors.New("db down")},
		writer: &fakeKafkaWriter{},
		source: "state-service",
	}
	if err := p.publishBatch(context.Background()); err == nil {
		t.Fatal("expected fetch error")
	}
}

func TestPublishBatch_PublishesRow(t *testing.T) {
	t.Parallel()

	payload, _ := json.Marshal(map[string]any{"personaId": "abc"})
	p := &Publisher{
		store: &fakeOutboxStore{
			rows: []store.OutboxRow{{
				ID: 1, Topic: "twin.state.updated", PartitionKey: "abc", Payload: payload,
			}},
		},
		writer: &fakeKafkaWriter{},
		source: "state-service",
	}
	if err := p.publishBatch(context.Background()); err != nil {
		t.Fatalf("publishBatch: %v", err)
	}
}

func TestPublishBatch_SingleWriteForMultipleRows(t *testing.T) {
	t.Parallel()

	payload, _ := json.Marshal(map[string]any{"personaId": "abc"})
	writer := &fakeKafkaWriter{}
	p := &Publisher{
		store: &fakeOutboxStore{
			rows: []store.OutboxRow{
				{ID: 1, Topic: "twin.state.updated", PartitionKey: "abc", Payload: payload},
				{ID: 2, Topic: "twin.state.updated", PartitionKey: "def", Payload: payload},
				{ID: 3, Topic: "twin.state.updated", PartitionKey: "ghi", Payload: payload},
			},
		},
		writer:    writer,
		source:    "state-service",
		batchSize: 100,
	}
	if err := p.publishBatch(context.Background()); err != nil {
		t.Fatalf("publishBatch: %v", err)
	}
	if len(writer.batchLens) != 1 {
		t.Fatalf("WriteMessages calls = %d, want 1", len(writer.batchLens))
	}
	if writer.batchLens[0] != 3 {
		t.Fatalf("batch size = %d, want 3", writer.batchLens[0])
	}
}
