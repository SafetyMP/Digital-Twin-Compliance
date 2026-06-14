package events

import (
	"encoding/json"
	"fmt"
	"time"
)

type Envelope struct {
	EventID        string          `json:"eventId"`
	EventType      string          `json:"eventType"`
	EventVersion   string          `json:"eventVersion"`
	Source         string          `json:"source"`
	CorrelationID  string          `json:"correlationId"`
	CausationID    *string         `json:"causationId"`
	Timestamp      string          `json:"timestamp"`
	IdempotencyKey string          `json:"idempotencyKey"`
	Payload        json.RawMessage `json:"payload"`
}

type ComplianceAlertRaised struct {
	AlertID     string            `json:"alertId"`
	RuleCode    string            `json:"ruleCode"`
	Regime      string            `json:"regime"`
	Severity    string            `json:"severity"`
	Status      string            `json:"status"`
	PersonaID   string            `json:"personaId"`
	PersonaType string            `json:"personaType"`
	Summary     string            `json:"summary"`
	Details     map[string]string `json:"details"`
	DetectedAt  string            `json:"detectedAt"`
	EvidenceRef *string           `json:"evidenceRef"`
}

func ParseEnvelope(data []byte) (Envelope, error) {
	var env Envelope
	if err := json.Unmarshal(data, &env); err != nil {
		return Envelope{}, err
	}
	return env, nil
}

func ParseAlertPayload(raw json.RawMessage) (ComplianceAlertRaised, error) {
	var alert ComplianceAlertRaised
	if err := json.Unmarshal(raw, &alert); err != nil {
		return ComplianceAlertRaised{}, err
	}
	if alert.AlertID == "" || alert.RuleCode == "" {
		return ComplianceAlertRaised{}, fmt.Errorf("missing required alert fields")
	}
	if alert.Status == "" {
		alert.Status = "Open"
	}
	return alert, nil
}

func ParseDetectedAt(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, fmt.Errorf("empty detectedAt")
	}
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return time.Parse(time.RFC3339, s)
	}
	return t, nil
}
