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
	lcr := 0.0
	if state.Liquidity != nil {
		if v, ok := state.Liquidity["lcr"].(float64); ok {
			lcr = v
		}
	}
	return h.store.UpsertInstitution(ctx, entityID, state.LegalName, lcr, 0.12)
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
