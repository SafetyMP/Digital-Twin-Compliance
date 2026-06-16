package audit

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/digital-twin/platform/services/cedar-service/internal/decision"
)

type mockWriter struct {
	messages [][]byte
}

func (m *mockWriter) WriteMessages(ctx context.Context, msgs ...interface{}) error {
	for _, msg := range msgs {
		switch v := msg.(type) {
		case []byte:
			m.messages = append(m.messages, v)
		}
	}
	return nil
}

func TestPublishDenyEnvelope(t *testing.T) {
	t.Parallel()

	p := &Publisher{source: "cedar-service"}
	dec := decision.RuleDecision{
		DecisionID:    "dec-1",
		RuleCode:      "INT-R003",
		Outcome:       "Deny",
		Score:         0.95,
		Rationale:     "denied",
		PolicyVersion: "1.0.0",
		EvaluatedAt:   "2026-06-15T12:00:00Z",
		InputHash:     "sha256:abc",
	}

	// Build envelope without kafka write.
	decisionPayload, err := json.Marshal(dec)
	if err != nil {
		t.Fatal(err)
	}
	metadata, err := json.Marshal(map[string]string{
		"regime":        regimeForRule(dec.RuleCode),
		"policyVersion": dec.PolicyVersion,
	})
	if err != nil {
		t.Fatal(err)
	}
	pending := map[string]any{
		"entryType":     "RuleDecision",
		"correlationId": "corr-1",
		"subject":       map[string]string{"subjectId": "twin-1", "subjectType": "TwinPersona"},
		"actor":         map[string]string{"actorId": "user-1", "actorType": "Principal"},
		"action":        "RuleEvaluated",
		"payload":       json.RawMessage(decisionPayload),
		"metadata":      json.RawMessage(metadata),
	}
	payload, err := json.Marshal(pending)
	if err != nil {
		t.Fatal(err)
	}
	env := map[string]any{
		"eventType":      "AuditPending",
		"source":         p.source,
		"idempotencyKey": "audit-rule-INT-R003-dec-1",
		"payload":        json.RawMessage(payload),
	}
	body, err := json.Marshal(env)
	if err != nil {
		t.Fatal(err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatal(err)
	}
	if parsed["eventType"] != "AuditPending" {
		t.Fatalf("eventType = %#v", parsed["eventType"])
	}
	if parsed["source"] != "cedar-service" {
		t.Fatalf("source = %#v", parsed["source"])
	}

	_ = p.PublishDeny(context.Background(), decision.RuleDecision{Outcome: "Allow"}, "", "", "")
}

func TestNoopPublisher(t *testing.T) {
	t.Parallel()
	var n Noop
	if err := n.PublishDeny(context.Background(), decision.RuleDecision{Outcome: "Deny"}, "u", "s", "c"); err != nil {
		t.Fatal(err)
	}
}
