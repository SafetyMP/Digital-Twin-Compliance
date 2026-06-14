package events

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const EventVersion = "1.0"

type Envelope struct {
	EventID        string  `json:"eventId"`
	EventType      string  `json:"eventType"`
	EventVersion   string  `json:"eventVersion"`
	Source         string  `json:"source"`
	CorrelationID  string  `json:"correlationId"`
	CausationID    *string `json:"causationId"`
	Timestamp      string  `json:"timestamp"`
	IdempotencyKey string  `json:"idempotencyKey"`
	Payload        string  `json:"payload"`
}

type EntityStateUpdated struct {
	PersonaID       string          `json:"personaId"`
	PersonaType     string          `json:"personaType"`
	SourceEntityID  string          `json:"sourceEntityId"`
	StateVersion    int             `json:"stateVersion"`
	ChangedFields   []string        `json:"changedFields"`
	CurrentState    json.RawMessage `json:"currentState"`
	SourceSystem    string          `json:"sourceSystem"`
	SourceTimestamp string          `json:"sourceTimestamp"`
}

type TwinStateUpdated struct {
	PersonaID        string          `json:"personaId"`
	PersonaType      string          `json:"personaType"`
	SourceEntityID   string          `json:"sourceEntityId"`
	StateVersion     int             `json:"stateVersion"`
	ComplianceStatus string          `json:"complianceStatus"`
	LastSyncedAt     string          `json:"lastSyncedAt"`
	CurrentState     json.RawMessage `json:"currentState,omitempty"`
}

func NewEnvelope(eventType, source, idempotencyKey string, payload any) (Envelope, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return Envelope{}, err
	}
	return Envelope{
		EventID:        uuid.NewString(),
		EventType:      eventType,
		EventVersion:   EventVersion,
		Source:         source,
		CorrelationID:  uuid.NewString(),
		Timestamp:      time.Now().UTC().Format(time.RFC3339Nano),
		IdempotencyKey: idempotencyKey,
		Payload:        string(payloadBytes),
	}, nil
}
