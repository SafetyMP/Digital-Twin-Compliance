package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type PendingPublisher struct {
	writer *kafka.Writer
	source string
}

func NewPendingPublisher(brokers []string, topic, source string) *PendingPublisher {
	return &PendingPublisher{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			BatchTimeout: 10 * time.Millisecond,
		},
		source: source,
	}
}

type AlertAuditInput struct {
	AlertID        string
	RuleCode       string
	Regime         string
	Severity       string
	Summary        string
	SourceEventID  string
	IdempotencyKey string
	CorrelationID  string
}

func (p *PendingPublisher) PublishAlertRaised(ctx context.Context, in AlertAuditInput) error {
	correlationID := in.CorrelationID
	if correlationID == "" {
		correlationID = in.AlertID
	}
	pending := map[string]any{
		"entryType":     "Alert",
		"correlationId": correlationID,
		"subject": map[string]string{
			"subjectId":   in.AlertID,
			"subjectType": "ComplianceAlert",
		},
		"actor": map[string]string{
			"actorId":   p.source,
			"actorType": "Service",
		},
		"action": "ComplianceAlertRaised",
		"payload": map[string]string{
			"alertId":  in.AlertID,
			"ruleCode": in.RuleCode,
			"severity": in.Severity,
			"summary":  in.Summary,
		},
		"metadata": map[string]string{
			"regime":         in.Regime,
			"policyVersion":  "phase2-inline",
			"sourceEventId":  in.SourceEventID,
			"retentionUntil": time.Now().AddDate(7, 0, 0).Format("2006-01-02"),
		},
	}
	payload, err := json.Marshal(pending)
	if err != nil {
		return err
	}
	idempotencyKey := fmt.Sprintf("audit-alert-%s", in.AlertID)
	env := map[string]any{
		"eventId":        uuid.NewString(),
		"eventType":      "AuditPending",
		"eventVersion":   "1.0",
		"source":         p.source,
		"correlationId":  correlationID,
		"timestamp":      time.Now().UTC().Format(time.RFC3339Nano),
		"idempotencyKey": idempotencyKey,
		"payload":        json.RawMessage(payload),
	}
	body, err := json.Marshal(env)
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{Key: []byte(in.AlertID), Value: body})
}

func (p *PendingPublisher) Close() error {
	if p == nil || p.writer == nil {
		return nil
	}
	return p.writer.Close()
}
