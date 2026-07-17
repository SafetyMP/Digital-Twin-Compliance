package consumer

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

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

// isPoison reports parse/validation errors that will not succeed on retry.
func isPoison(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	for _, needle := range []string{
		"missing eventtype",
		"invalid character",
		"unexpected end of json",
		"cannot unmarshal",
		"missing entity id",
		"institution missing entity id",
	} {
		if strings.Contains(msg, needle) {
			return true
		}
	}
	return false
}
