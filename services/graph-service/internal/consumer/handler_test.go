package consumer

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/digital-twin/platform/services/graph-service/internal/graph"
)

type mockGraphStore struct {
	institutions []string
	exposures    []graph.ExposureInput
}

func (m *mockGraphStore) UpsertInstitution(_ context.Context, entityID, name string, _, _ float64) error {
	m.institutions = append(m.institutions, entityID+":"+name)
	return nil
}

func (m *mockGraphStore) UpsertExposure(_ context.Context, in graph.ExposureInput) error {
	m.exposures = append(m.exposures, in)
	return nil
}

func TestHandleTwinInstitution(t *testing.T) {
	st := &mockGraphStore{}
	h := NewHandler(st)
	env := map[string]any{
		"eventId":        "e1",
		"eventType":      "TwinStateUpdated",
		"eventVersion":   "1.0",
		"source":         "state-service",
		"correlationId":  "c1",
		"timestamp":      "2026-06-14T01:00:00Z",
		"idempotencyKey": "k1",
		"payload":        `{"personaId":"p1","personaType":"Institution","sourceEntityId":"ent-1","currentState":{"entity_id":"ent-1","legal_name":"Test Bank","liquidity":{"lcr":0.95}}}`,
	}
	body, _ := json.Marshal(env)
	if err := h.HandleTwinMessage(context.Background(), body); err != nil {
		t.Fatal(err)
	}
	if len(st.institutions) != 1 {
		t.Fatalf("institutions = %d", len(st.institutions))
	}
}

func TestHandleTwinInstrument(t *testing.T) {
	st := &mockGraphStore{}
	h := NewHandler(st)
	env := map[string]any{
		"eventId":        "e2",
		"eventType":      "TwinStateUpdated",
		"eventVersion":   "1.0",
		"source":         "state-service",
		"correlationId":  "c2",
		"timestamp":      "2026-06-14T01:00:00Z",
		"idempotencyKey": "k2",
		"payload":        `{"personaId":"p2","personaType":"Instrument","sourceEntityId":"inst-1","currentState":{"instrument_id":"inst-1","owner_entity_id":"a","counterparty_id":"b","notional_amount":1000000,"instrument_type":"Loan"}}`,
	}
	body, _ := json.Marshal(env)
	if err := h.HandleTwinMessage(context.Background(), body); err != nil {
		t.Fatal(err)
	}
	if len(st.exposures) != 1 {
		t.Fatalf("exposures = %d", len(st.exposures))
	}
	if st.exposures[0].Layer != "ShortTerm" {
		t.Fatalf("layer = %q", st.exposures[0].Layer)
	}
}
