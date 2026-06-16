package events

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	EventTypeAuditPending  = "AuditPending"
	EventTypeAuditRecorded = "AuditRecorded"
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

type SubjectRef struct {
	SubjectID   string `json:"subjectId"`
	SubjectType string `json:"subjectType"`
}

type ActorRef struct {
	ActorID   string `json:"actorId"`
	ActorType string `json:"actorType"`
}

type AuditPending struct {
	EntryType     string          `json:"entryType"`
	CorrelationID string          `json:"correlationId"`
	Subject       SubjectRef      `json:"subject"`
	Actor         ActorRef        `json:"actor"`
	Action        string          `json:"action"`
	Payload       json.RawMessage `json:"payload"`
	Metadata      json.RawMessage `json:"metadata"`
}

type AuditEntry struct {
	EntryID        string          `json:"entryId"`
	EntryType      string          `json:"entryType"`
	SequenceNumber int64           `json:"sequenceNumber"`
	RecordedAt     string          `json:"recordedAt"`
	CorrelationID  string          `json:"correlationId"`
	Subject        SubjectRef      `json:"subject"`
	Actor          ActorRef        `json:"actor"`
	Action         string          `json:"action"`
	Payload        json.RawMessage `json:"payload"`
	PayloadHash    string          `json:"payloadHash"`
	PreviousHash   string          `json:"previousHash"`
	Metadata       json.RawMessage `json:"metadata"`
	IdempotencyKey string          `json:"idempotencyKey,omitempty"`
}

type AuditRecordedSummary struct {
	EntryID        string `json:"entryId"`
	EntryType      string `json:"entryType"`
	SequenceNumber int64  `json:"sequenceNumber"`
	RecordedAt     string `json:"recordedAt"`
	CorrelationID  string `json:"correlationId"`
	SubjectID      string `json:"subjectId"`
	SubjectType    string `json:"subjectType"`
	PayloadHash    string `json:"payloadHash"`
	PreviousHash   string `json:"previousHash"`
	EvidenceRef    string `json:"evidenceRef"`
}

func ParseEnvelope(data []byte) (Envelope, error) {
	var env Envelope
	if err := json.Unmarshal(data, &env); err != nil {
		return Envelope{}, err
	}
	return env, nil
}

func ParseAuditPending(raw json.RawMessage) (AuditPending, error) {
	var pending AuditPending
	if err := json.Unmarshal(raw, &pending); err != nil {
		return AuditPending{}, err
	}
	if pending.EntryType == "" || pending.Subject.SubjectID == "" {
		return AuditPending{}, fmt.Errorf("missing required audit pending fields")
	}
	if pending.Metadata == nil {
		pending.Metadata = json.RawMessage(`{}`)
	}
	if pending.Payload == nil {
		pending.Payload = json.RawMessage(`{}`)
	}
	return pending, nil
}

func NewRecordedEnvelope(entry AuditEntry, source string) (Envelope, error) {
	summary := AuditRecordedSummary{
		EntryID:        entry.EntryID,
		EntryType:      entry.EntryType,
		SequenceNumber: entry.SequenceNumber,
		RecordedAt:     entry.RecordedAt,
		CorrelationID:  entry.CorrelationID,
		SubjectID:      entry.Subject.SubjectID,
		SubjectType:    entry.Subject.SubjectType,
		PayloadHash:    entry.PayloadHash,
		PreviousHash:   entry.PreviousHash,
		EvidenceRef:    entry.EntryID,
	}
	payload, err := json.Marshal(summary)
	if err != nil {
		return Envelope{}, err
	}
	return Envelope{
		EventID:        entry.EntryID,
		EventType:      EventTypeAuditRecorded,
		EventVersion:   "1.0",
		Source:         source,
		CorrelationID:  entry.CorrelationID,
		Timestamp:      time.Now().UTC().Format(time.RFC3339Nano),
		IdempotencyKey: entry.IdempotencyKey,
		Payload:        payload,
	}, nil
}
