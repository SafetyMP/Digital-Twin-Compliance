package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/digital-twin/platform/services/alert-service/internal/events"
	"github.com/digital-twin/platform/services/alert-service/internal/hub"
	"github.com/digital-twin/platform/services/alert-service/internal/store"
	"github.com/segmentio/kafka-go"
)

type Handler struct {
	store *store.Store
	hub   *hub.Hub
}

func NewHandler(st *store.Store, h *hub.Hub) *Handler {
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

type Runner struct {
	reader  *kafka.Reader
	handler *Handler
}

func NewRunner(brokers []string, group, topic string, handler *Handler) *Runner {
	return &Runner{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			GroupID:  group,
			Topic:    topic,
			MinBytes: 1,
			MaxBytes: 10e6,
		}),
		handler: handler,
	}
}

func (r *Runner) Run(ctx context.Context) error {
	for {
		msg, err := r.reader.FetchMessage(ctx)
		if err != nil {
			return err
		}
		if err := r.handler.HandleMessage(ctx, msg.Value); err != nil {
			slog.Error("handle alert message", "error", err, "offset", msg.Offset)
		}
		if err := r.reader.CommitMessages(ctx, msg); err != nil {
			slog.Error("commit offset", "error", err)
		}
	}
}

func (r *Runner) Close() error {
	return r.reader.Close()
}
