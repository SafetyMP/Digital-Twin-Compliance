package consumer

import (
	"context"
	"encoding/json"

	"github.com/digital-twin/platform/services/alert-service/internal/events"
	"github.com/digital-twin/platform/services/alert-service/internal/store"
)

type alertStore interface {
	UpsertAlert(ctx context.Context, input store.UpsertInput) (store.Alert, bool, error)
}

type alertHub interface {
	Broadcast(msgType string, alert store.Alert)
}

type Handler struct {
	store alertStore
	hub   alertHub
}

func NewHandler(st alertStore, h alertHub) *Handler {
	return &Handler{store: st, hub: h}
}

func (h *Handler) HandleMessage(ctx context.Context, data []byte) error {
	env, err := events.ParseEnvelope(data)
	if err != nil {
		return err
	}
	if env.EventType != "ComplianceAlertRaised" {
		return nil
	}

	alert, err := events.ParseAlertPayload(env.Payload)
	if err != nil {
		return err
	}

	detectedAt, err := events.ParseDetectedAt(alert.DetectedAt)
	if err != nil {
		return err
	}

	detailsBytes, err := json.Marshal(alert.Details)
	if err != nil {
		return err
	}

	saved, created, err := h.store.UpsertAlert(ctx, store.UpsertInput{
		AlertID:        alert.AlertID,
		RuleCode:       alert.RuleCode,
		Regime:         alert.Regime,
		Severity:       alert.Severity,
		Status:         alert.Status,
		PersonaID:      alert.PersonaID,
		PersonaType:    alert.PersonaType,
		Summary:        alert.Summary,
		Details:        detailsBytes,
		DetectedAt:     detectedAt,
		IdempotencyKey: env.IdempotencyKey,
	})
	if err != nil {
		return err
	}
	if created {
		h.hub.Broadcast("alert.raised", saved)
	}
	return nil
}
