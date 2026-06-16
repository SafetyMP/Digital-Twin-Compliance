package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/digital-twin/platform/services/decision-service/internal/decision"
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

type PublishInput struct {
	Decision      decision.RuleDecision
	Input         map[string]any
	CorrelationID string
}

func (p *PendingPublisher) PublishRuleDecision(ctx context.Context, in PublishInput) error {
	if p == nil || p.writer == nil {
		return nil
	}
	if !decision.RequiresAudit(in.Decision.Outcome) {
		return nil
	}

	subjectID, subjectType := decision.SubjectFromInput(in.Input)
	correlationID := in.CorrelationID
	if correlationID == "" {
		correlationID = in.Decision.DecisionID
	}

	decisionPayload, err := json.Marshal(in.Decision)
	if err != nil {
		return err
	}

	pending := map[string]any{
		"entryType":     "RuleDecision",
		"correlationId": correlationID,
		"subject": map[string]string{
			"subjectId":   subjectID,
			"subjectType": subjectType,
		},
		"actor": map[string]string{
			"actorId":   p.source,
			"actorType": "Service",
		},
		"action":  "RuleEvaluated",
		"payload": json.RawMessage(decisionPayload),
		"metadata": map[string]string{
			"regime":         decision.RegimeForRuleCode(in.Decision.RuleCode),
			"policyVersion":  in.Decision.PolicyVersion,
			"retentionUntil": time.Now().AddDate(7, 0, 0).Format("2006-01-02"),
		},
	}
	payload, err := json.Marshal(pending)
	if err != nil {
		return err
	}

	idempotencyKey := fmt.Sprintf("audit-rule-%s-%s", in.Decision.RuleCode, in.Decision.DecisionID)
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
	return p.writer.WriteMessages(ctx, kafka.Message{Key: []byte(in.Decision.DecisionID), Value: body})
}

func (p *PendingPublisher) Close() error {
	if p == nil || p.writer == nil {
		return nil
	}
	return p.writer.Close()
}
