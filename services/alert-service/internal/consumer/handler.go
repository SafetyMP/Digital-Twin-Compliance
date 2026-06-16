package consumer

import (
	"context"
	"encoding/json"

	"github.com/digital-twin/platform/services/alert-service/internal/audit"
	"github.com/digital-twin/platform/services/alert-service/internal/events"
	"github.com/digital-twin/platform/services/alert-service/internal/store"
)

type alertStore interface {
	UpsertAlert(ctx context.Context, input store.UpsertInput) (store.Alert, bool, error)
}

type alertHub interface {
	Broadcast(msgType string, alert store.Alert)
}

type auditPublisher interface {
	PublishAlertRaised(ctx context.Context, in audit.AlertAuditInput) error
}

type Handler struct {
	store     alertStore
	hub       alertHub
	auditPub  auditPublisher
	source    string
	lastEnvID string
}

func NewHandler(st alertStore, h alertHub, pub auditPublisher, source string) *Handler {
	return &Handler{store: st, hub: h, auditPub: pub, source: source}
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
		if h.auditPub != nil {
			sourceEventID := env.EventID
			_ = h.auditPub.PublishAlertRaised(ctx, audit.AlertAuditInput{
				AlertID:        saved.AlertID,
				RuleCode:       saved.RuleCode,
				Regime:         saved.Regime,
				Severity:       saved.Severity,
				Summary:        saved.Summary,
				SourceEventID:  sourceEventID,
				IdempotencyKey: env.IdempotencyKey,
				CorrelationID:  env.CorrelationID,
			})
		}
	}
	return nil
}
