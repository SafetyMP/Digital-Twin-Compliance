package events

import (
	"encoding/json"
	"fmt"
)

type Envelope struct {
	EventID        string `json:"eventId"`
	EventType      string `json:"eventType"`
	EventVersion   string `json:"eventVersion"`
	Source         string `json:"source"`
	CorrelationID  string `json:"correlationId"`
	Timestamp      string `json:"timestamp"`
	IdempotencyKey string `json:"idempotencyKey"`
	Payload        string `json:"payload"`
}

type TwinStatePayload struct {
	PersonaID      string          `json:"personaId"`
	PersonaType    string          `json:"personaType"`
	SourceEntityID string          `json:"sourceEntityId"`
	StateVersion   int             `json:"stateVersion"`
	CurrentState   json.RawMessage `json:"currentState"`
}

type InstitutionState struct {
	EntityID   string         `json:"entity_id"`
	LegalName  string         `json:"legal_name"`
	EntityType string         `json:"entity_type"`
	Liquidity  map[string]any `json:"liquidity"`
}

type InstrumentState struct {
	InstrumentID   string  `json:"instrument_id"`
	OwnerEntityID  string  `json:"owner_entity_id"`
	CounterpartyID string  `json:"counterparty_id"`
	NotionalAmount float64 `json:"notional_amount"`
	Currency       string  `json:"currency"`
	InstrumentType string  `json:"instrument_type"`
}

func ParseEnvelope(data []byte) (Envelope, error) {
	var env Envelope
	if err := json.Unmarshal(data, &env); err != nil {
		return Envelope{}, err
	}
	if env.EventType == "" {
		return Envelope{}, fmt.Errorf("missing eventType")
	}
	return env, nil
}

func ParseTwinPayload(raw string) (TwinStatePayload, error) {
	var p TwinStatePayload
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		return TwinStatePayload{}, err
	}
	return p, nil
}
