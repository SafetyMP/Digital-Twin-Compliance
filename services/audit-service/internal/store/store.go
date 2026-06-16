package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/digital-twin/platform/services/audit-service/internal/events"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")

type EntryIndex struct {
	EntryID        string    `json:"entryId"`
	TenantID       string    `json:"tenantId"`
	SequenceNumber int64     `json:"sequenceNumber"`
	EntryType      string    `json:"entryType"`
	RuleCode       *string   `json:"ruleCode,omitempty"`
	SubjectID      *string   `json:"subjectId,omitempty"`
	SubjectType    *string   `json:"subjectType,omitempty"`
	CorrelationID  string    `json:"correlationId"`
	RecordedAt     time.Time `json:"recordedAt"`
	PayloadHash    string    `json:"payloadHash"`
	PreviousHash   string    `json:"previousHash"`
	IdempotencyKey string    `json:"idempotencyKey"`
}

type ListFilter struct {
	RuleCode  string
	SubjectID string
	From      *time.Time
	To        *time.Time
	Limit     int
	Offset    int
}

type Store struct {
	pool     *pgxpool.Pool
	tenantID string
}

func New(pool *pgxpool.Pool, tenantID string) *Store {
	return &Store{pool: pool, tenantID: tenantID}
}

func (s *Store) IsProcessed(ctx context.Context, idempotencyKey string) (string, bool, error) {
	if idempotencyKey == "" {
		return "", false, nil
	}
	var entryID string
	err := s.pool.QueryRow(ctx, `
		SELECT entry_id FROM audit_idempotency_keys WHERE idempotency_key = $1
	`, idempotencyKey).Scan(&entryID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return entryID, true, nil
}

func (s *Store) RecordEntry(ctx context.Context, entry events.AuditEntry, ruleCode string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if entry.IdempotencyKey != "" {
		if _, err := tx.Exec(ctx, `
			INSERT INTO audit_idempotency_keys (idempotency_key, entry_id, tenant_id)
			VALUES ($1, $2, $3)
			ON CONFLICT (idempotency_key) DO NOTHING
		`, entry.IdempotencyKey, entry.EntryID, s.tenantID); err != nil {
			return err
		}
	}

	var ruleCodeVal *string
	if ruleCode != "" {
		ruleCodeVal = &ruleCode
	}
	subjectID := entry.Subject.SubjectID
	subjectType := entry.Subject.SubjectType
	recordedAt, err := parseRecordedAt(entry.RecordedAt)
	if err != nil {
		return fmt.Errorf("parse recordedAt: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO audit_entry_index (
			entry_id, tenant_id, sequence_number, entry_type, rule_code,
			subject_id, subject_type, correlation_id, recorded_at,
			payload_hash, previous_hash, idempotency_key
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, entry.EntryID, s.tenantID, entry.SequenceNumber, entry.EntryType, ruleCodeVal,
		subjectID, subjectType, entry.CorrelationID, recordedAt,
		entry.PayloadHash, entry.PreviousHash, entry.IdempotencyKey); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *Store) ListEntries(ctx context.Context, filter ListFilter) ([]EntryIndex, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT entry_id, tenant_id, sequence_number, entry_type, rule_code,
		       subject_id, subject_type, correlation_id, recorded_at,
		       payload_hash, previous_hash, idempotency_key
		FROM audit_entry_index
		WHERE tenant_id = $1
	`
	args := []any{s.tenantID}
	argN := 2
	if filter.RuleCode != "" {
		query += fmt.Sprintf(" AND rule_code = $%d", argN)
		args = append(args, filter.RuleCode)
		argN++
	}
	if filter.SubjectID != "" {
		query += fmt.Sprintf(" AND subject_id = $%d", argN)
		args = append(args, filter.SubjectID)
		argN++
	}
	if filter.From != nil {
		query += fmt.Sprintf(" AND recorded_at >= $%d", argN)
		args = append(args, *filter.From)
		argN++
	}
	if filter.To != nil {
		query += fmt.Sprintf(" AND recorded_at <= $%d", argN)
		args = append(args, *filter.To)
		argN++
	}
	query += fmt.Sprintf(" ORDER BY sequence_number ASC LIMIT $%d OFFSET $%d", argN, argN+1)
	args = append(args, limit, offset)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []EntryIndex
	for rows.Next() {
		idx, err := scanIndex(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, idx)
	}
	return out, rows.Err()
}

func (s *Store) GetIndex(ctx context.Context, entryID string) (EntryIndex, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT entry_id, tenant_id, sequence_number, entry_type, rule_code,
		       subject_id, subject_type, correlation_id, recorded_at,
		       payload_hash, previous_hash, idempotency_key
		FROM audit_entry_index
		WHERE tenant_id = $1 AND entry_id = $2
	`, s.tenantID, entryID)
	idx, err := scanIndex(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return EntryIndex{}, ErrNotFound
	}
	return idx, err
}

func (s *Store) IndexCount(ctx context.Context) (int64, error) {
	var count int64
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM audit_entry_index WHERE tenant_id = $1
	`, s.tenantID).Scan(&count)
	return count, err
}

func (s *Store) ListForVerify(ctx context.Context, fromSeq, toSeq int64) ([]EntryIndex, error) {
	query := `
		SELECT entry_id, tenant_id, sequence_number, entry_type, rule_code,
		       subject_id, subject_type, correlation_id, recorded_at,
		       payload_hash, previous_hash, idempotency_key
		FROM audit_entry_index
		WHERE tenant_id = $1
	`
	args := []any{s.tenantID}
	argN := 2
	if fromSeq > 0 {
		query += fmt.Sprintf(" AND sequence_number >= $%d", argN)
		args = append(args, fromSeq)
		argN++
	}
	if toSeq > 0 {
		query += fmt.Sprintf(" AND sequence_number <= $%d", argN)
		args = append(args, toSeq)
		argN++
	}
	query += " ORDER BY sequence_number ASC"

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []EntryIndex
	for rows.Next() {
		idx, err := scanIndex(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, idx)
	}
	return out, rows.Err()
}

func (s *Store) InsertDLQ(ctx context.Context, idempotencyKey, topic string, partition int, offset int64, errMsg string, payload json.RawMessage) error {
	if payload == nil {
		payload = json.RawMessage(`{}`)
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO audit_outbox_dlq (
			tenant_id, idempotency_key, original_topic, partition, kafka_offset, error_message, payload
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, s.tenantID, nullIfEmpty(idempotencyKey), topic, partition, offset, errMsg, payload)
	return err
}

func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

type scannable interface {
	Scan(dest ...any) error
}

func scanIndex(row scannable) (EntryIndex, error) {
	var idx EntryIndex
	err := row.Scan(
		&idx.EntryID, &idx.TenantID, &idx.SequenceNumber, &idx.EntryType, &idx.RuleCode,
		&idx.SubjectID, &idx.SubjectType, &idx.CorrelationID, &idx.RecordedAt,
		&idx.PayloadHash, &idx.PreviousHash, &idx.IdempotencyKey,
	)
	return idx, err
}

func RunMigrations(ctx context.Context, pool *pgxpool.Pool, sql string) error {
	_, err := pool.Exec(ctx, sql)
	return err
}

func ExtractRuleCode(payload json.RawMessage) string {
	var m map[string]any
	if err := json.Unmarshal(payload, &m); err != nil {
		return ""
	}
	if v, ok := m["ruleCode"].(string); ok {
		return v
	}
	return ""
}
