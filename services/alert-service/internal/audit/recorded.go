package audit

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/segmentio/kafka-go"
)

type evidenceUpdater interface {
	SetEvidenceRef(ctx context.Context, alertID, evidenceRef string) error
}

type RecordedHandler struct {
	store evidenceUpdater
}

func NewRecordedHandler(st evidenceUpdater) *RecordedHandler {
	return &RecordedHandler{store: st}
}

func (h *RecordedHandler) HandleMessage(ctx context.Context, data []byte) error {
	var env struct {
		EventType string          `json:"eventType"`
		Payload   json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(data, &env); err != nil {
		return err
	}
	if env.EventType != "AuditRecorded" && env.EventType != "AuditEntryRecorded" {
		return nil
	}
	var summary struct {
		EntryID     string `json:"entryId"`
		EntryType   string `json:"entryType"`
		EvidenceRef string `json:"evidenceRef"`
		AlertID     string `json:"alertId"`
		SubjectID   string `json:"subjectId"`
		SubjectType string `json:"subjectType"`
	}
	if err := json.Unmarshal(env.Payload, &summary); err != nil {
		return err
	}
	evidenceRef := summary.EvidenceRef
	if evidenceRef == "" {
		evidenceRef = summary.EntryID
	}
	alertID := summary.AlertID
	if alertID == "" && summary.EntryType == "Alert" && summary.SubjectType == "ComplianceAlert" {
		alertID = summary.SubjectID
	}
	if summary.EntryType != "Alert" || alertID == "" {
		return nil
	}
	if err := h.store.SetEvidenceRef(ctx, alertID, evidenceRef); err != nil {
		slog.Warn("set evidence ref", "alertId", alertID, "error", err)
		return err
	}
	return nil
}

type RecordedRunner struct {
	reader  *kafka.Reader
	handler *RecordedHandler
}

func NewRecordedRunner(brokers []string, group, topic string, handler *RecordedHandler) *RecordedRunner {
	return &RecordedRunner{
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

func (r *RecordedRunner) Run(ctx context.Context) error {
	for {
		msg, err := r.reader.FetchMessage(ctx)
		if err != nil {
			return err
		}
		if err := r.handler.HandleMessage(ctx, msg.Value); err != nil {
			slog.Error("handle audit recorded", "error", err)
		}
		if err := r.reader.CommitMessages(ctx, msg); err != nil {
			slog.Error("commit audit recorded offset", "error", err)
		}
	}
}

func (r *RecordedRunner) Close() error {
	return r.reader.Close()
}
