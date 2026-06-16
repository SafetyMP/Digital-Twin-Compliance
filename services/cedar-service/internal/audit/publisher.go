package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/digital-twin/platform/services/cedar-service/internal/decision"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type Publisher struct {
	writer *kafka.Writer
	source string
}

func NewPublisher(brokers []string, topic, source string) *Publisher {
	return &Publisher{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			BatchTimeout: 10 * time.Millisecond,
			RequiredAcks: kafka.RequireOne,
		},
		source: source,
	}
}

func (p *Publisher) PublishDeny(ctx context.Context, d decision.RuleDecision, principalID, subjectID, correlationID string) error {
	if p == nil || p.writer == nil || d.Outcome != "Deny" {
		return nil
	}
	if correlationID == "" {
		correlationID = d.DecisionID
	}
	if principalID == "" {
		principalID = p.source
	}
	if subjectID == "" {
		subjectID = d.DecisionID
	}

	decisionPayload, err := json.Marshal(d)
	if err != nil {
		return err
	}
	metadata, err := json.Marshal(map[string]string{
		"regime":        regimeForRule(d.RuleCode),
		"policyVersion": d.PolicyVersion,
	})
	if err != nil {
		return err
	}
	pending := map[string]any{
		"entryType":     "RuleDecision",
		"correlationId": correlationID,
		"subject": map[string]string{
			"subjectId":   subjectID,
			"subjectType": "TwinPersona",
		},
		"actor": map[string]string{
			"actorId":   principalID,
			"actorType": "Principal",
		},
		"action":   "RuleEvaluated",
		"payload":  json.RawMessage(decisionPayload),
		"metadata": json.RawMessage(metadata),
	}
	payload, err := json.Marshal(pending)
	if err != nil {
		return err
	}
	idempotencyKey := fmt.Sprintf("audit-rule-%s-%s", d.RuleCode, d.DecisionID)
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
	return p.writer.WriteMessages(ctx, kafka.Message{Key: []byte(idempotencyKey), Value: body})
}

func (p *Publisher) Close() error {
	if p == nil || p.writer == nil {
		return nil
	}
	return p.writer.Close()
}

type Noop struct{}

func (Noop) PublishDeny(context.Context, decision.RuleDecision, string, string, string) error {
	return nil
}
func (Noop) Close() error { return nil }

func regimeForRule(ruleCode string) string {
	switch {
	case len(ruleCode) >= 4 && ruleCode[:4] == "INT-":
		return "Internal"
	case len(ruleCode) >= 6 && ruleCode[:6] == "COREP-":
		return "COREP"
	case len(ruleCode) >= 5 && ruleCode[:5] == "EMIR-":
		return "EMIR"
	case len(ruleCode) >= 5 && ruleCode[:5] == "DORA-":
		return "DORA"
	default:
		return "Unknown"
	}
}
