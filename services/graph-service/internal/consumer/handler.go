package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/digital-twin/platform/services/graph-service/internal/events"
	"github.com/digital-twin/platform/services/graph-service/internal/graph"
)

type graphStore interface {
	UpsertInstitution(ctx context.Context, entityID, name string, lcr, cet1 float64) error
	UpsertExposure(ctx context.Context, in graph.ExposureInput) error
}

type Handler struct {
	store graphStore
}

func NewHandler(st graphStore) *Handler {
	return &Handler{store: st}
}

func (h *Handler) HandleTwinMessage(ctx context.Context, data []byte) error {
	env, err := events.ParseEnvelope(data)
	if err != nil {
		return err
	}
	if env.EventType != "TwinStateUpdated" {
		return nil
	}
	payload, err := events.ParseTwinPayload(env.Payload)
	if err != nil {
		return err
	}
	switch payload.PersonaType {
	case "Institution":
		return h.handleInstitution(ctx, payload)
	case "Instrument":
		return h.handleInstrument(ctx, payload)
	default:
		return nil
	}
}

func (h *Handler) handleInstitution(ctx context.Context, payload events.TwinStatePayload) error {
	var state events.InstitutionState
	if err := json.Unmarshal(payload.CurrentState, &state); err != nil {
		return err
	}
	entityID := payload.SourceEntityID
	if state.EntityID != "" {
		entityID = state.EntityID
	}
	if entityID == "" {
		return fmt.Errorf("institution missing entity id")
	}
	lcr := mapFloat(state.Liquidity, "lcr")
	cet1 := institutionCET1(state)
	return h.store.UpsertInstitution(ctx, entityID, state.LegalName, lcr, cet1)
}

func institutionCET1(state events.InstitutionState) float64 {
	if state.CET1Ratio != nil {
		return *state.CET1Ratio
	}
	if state.CET1 != nil {
		return *state.CET1
	}
	if v := mapFloat(state.Capital, "cet1", "cet1_ratio", "CET1"); v > 0 {
		return v
	}
	if v := mapFloat(state.Liquidity, "cet1", "cet1_ratio"); v > 0 {
		return v
	}
	// Explicit zero when twin omits capital — callers must not invent 0.12.
	return 0
}

func mapFloat(m map[string]any, keys ...string) float64 {
	if m == nil {
		return 0
	}
	for _, k := range keys {
		switch v := m[k].(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int:
			return float64(v)
		case json.Number:
			f, _ := v.Float64()
			return f
		}
	}
	return 0
}

func (h *Handler) handleInstrument(ctx context.Context, payload events.TwinStatePayload) error {
	var state events.InstrumentState
	if err := json.Unmarshal(payload.CurrentState, &state); err != nil {
		return err
	}
	if state.OwnerEntityID == "" || state.CounterpartyID == "" {
		return nil
	}
	edgeKey := state.InstrumentID
	if edgeKey == "" {
		edgeKey = payload.SourceEntityID
	}
	layer := classifyLayer(state.InstrumentType)
	expType := state.InstrumentType
	if expType == "" {
		expType = "Interbank"
	}
	return h.store.UpsertExposure(ctx, graph.ExposureInput{
		FromEntityID: state.OwnerEntityID,
		ToEntityID:   state.CounterpartyID,
		EdgeKey:      edgeKey,
		ExposureType: expType,
		NotionalEur:  state.NotionalAmount,
		Layer:        layer,
		InstrumentID: state.InstrumentID,
	})
}

func classifyLayer(instrumentType string) string {
	t := strings.ToLower(instrumentType)
	switch {
	case strings.Contains(t, "contingent"), strings.Contains(t, "guarantee"):
		return "Contingent"
	case strings.Contains(t, "bond"), strings.Contains(t, "long"):
		return "LongTerm"
	default:
		return "ShortTerm"
	}
}
