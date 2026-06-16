package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")

type Alert struct {
	AlertID        string          `json:"alertId"`
	TenantID       string          `json:"tenantId"`
	RuleCode       string          `json:"ruleCode"`
	Regime         string          `json:"regime"`
	Severity       string          `json:"severity"`
	Status         string          `json:"status"`
	PersonaID      string          `json:"personaId"`
	PersonaType    string          `json:"personaType"`
	Summary        string          `json:"summary"`
	Details        json.RawMessage `json:"details"`
	DetectedAt     time.Time       `json:"detectedAt"`
	AcknowledgedAt *time.Time      `json:"acknowledgedAt,omitempty"`
	AcknowledgedBy *string         `json:"acknowledgedBy,omitempty"`
	IdempotencyKey string          `json:"idempotencyKey"`
	EvidenceRef    *string         `json:"evidenceRef,omitempty"`
}

type UpsertInput struct {
	AlertID        string
	RuleCode       string
	Regime         string
	Severity       string
	Status         string
	PersonaID      string
	PersonaType    string
	Summary        string
	Details        json.RawMessage
	DetectedAt     time.Time
	IdempotencyKey string
}

type Store struct {
	pool     *pgxpool.Pool
	tenantID string
}

func New(pool *pgxpool.Pool, tenantID string) *Store {
	return &Store{pool: pool, tenantID: tenantID}
}

func (s *Store) UpsertAlert(ctx context.Context, input UpsertInput) (Alert, bool, error) {
	if input.Details == nil {
		input.Details = json.RawMessage(`{}`)
	}
	row := s.pool.QueryRow(ctx, `
		INSERT INTO compliance_alerts (
			alert_id, tenant_id, rule_code, regime, severity, status,
			persona_id, persona_type, summary, details, detected_at, idempotency_key, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, now())
		ON CONFLICT (idempotency_key) DO NOTHING
		RETURNING alert_id, tenant_id, rule_code, regime, severity, status,
		          persona_id, persona_type, summary, details, detected_at,
		          acknowledged_at, acknowledged_by, idempotency_key, evidence_ref
	`, input.AlertID, s.tenantID, input.RuleCode, input.Regime, input.Severity, input.Status,
		input.PersonaID, input.PersonaType, input.Summary, input.Details, input.DetectedAt, input.IdempotencyKey)

	alert, err := scanAlert(row)
	if err == nil {
		return alert, true, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return Alert{}, false, err
	}

	existing, err := s.GetByIdempotencyKey(ctx, input.IdempotencyKey)
	if err != nil {
		return Alert{}, false, err
	}
	return existing, false, nil
}

func (s *Store) GetByIdempotencyKey(ctx context.Context, key string) (Alert, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT alert_id, tenant_id, rule_code, regime, severity, status,
		       persona_id, persona_type, summary, details, detected_at,
		       acknowledged_at, acknowledged_by, idempotency_key, evidence_ref
		FROM compliance_alerts WHERE idempotency_key = $1
	`, key)
	return scanAlert(row)
}

func (s *Store) GetAlert(ctx context.Context, alertID string) (Alert, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT alert_id, tenant_id, rule_code, regime, severity, status,
		       persona_id, persona_type, summary, details, detected_at,
		       acknowledged_at, acknowledged_by, idempotency_key, evidence_ref
		FROM compliance_alerts
		WHERE tenant_id = $1 AND alert_id = $2
	`, s.tenantID, alertID)
	alert, err := scanAlert(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Alert{}, ErrNotFound
	}
	return alert, err
}

func (s *Store) ListAlerts(ctx context.Context, status, severity string, limit, offset int) ([]Alert, error) {
	query := `
		SELECT alert_id, tenant_id, rule_code, regime, severity, status,
		       persona_id, persona_type, summary, details, detected_at,
		       acknowledged_at, acknowledged_by, idempotency_key, evidence_ref
		FROM compliance_alerts
		WHERE tenant_id = $1
	`
	args := []any{s.tenantID}
	argN := 2
	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argN)
		args = append(args, status)
		argN++
	}
	if severity != "" {
		query += fmt.Sprintf(" AND severity = $%d", argN)
		args = append(args, severity)
		argN++
	}
	query += fmt.Sprintf(" ORDER BY detected_at DESC LIMIT $%d OFFSET $%d", argN, argN+1)
	args = append(args, limit, offset)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Alert
	for rows.Next() {
		alert, err := scanAlert(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, alert)
	}
	return out, rows.Err()
}

func (s *Store) Acknowledge(ctx context.Context, alertID, acknowledgedBy string) (Alert, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE compliance_alerts
		SET status = 'Acknowledged', acknowledged_at = now(), acknowledged_by = $3, updated_at = now()
		WHERE tenant_id = $1 AND alert_id = $2
		RETURNING alert_id, tenant_id, rule_code, regime, severity, status,
		          persona_id, persona_type, summary, details, detected_at,
		          acknowledged_at, acknowledged_by, idempotency_key, evidence_ref
	`, s.tenantID, alertID, acknowledgedBy)
	alert, err := scanAlert(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Alert{}, ErrNotFound
	}
	return alert, err
}

type scannable interface {
	Scan(dest ...any) error
}

func scanAlert(row scannable) (Alert, error) {
	var a Alert
	var details []byte
	err := row.Scan(
		&a.AlertID, &a.TenantID, &a.RuleCode, &a.Regime, &a.Severity, &a.Status,
		&a.PersonaID, &a.PersonaType, &a.Summary, &details, &a.DetectedAt,
		&a.AcknowledgedAt, &a.AcknowledgedBy, &a.IdempotencyKey, &a.EvidenceRef,
	)
	if err != nil {
		return Alert{}, err
	}
	a.Details = json.RawMessage(details)
	return a, nil
}

func (s *Store) SetEvidenceRef(ctx context.Context, alertID, evidenceRef string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE compliance_alerts
		SET evidence_ref = $3, updated_at = now()
		WHERE tenant_id = $1 AND alert_id = $2 AND (evidence_ref IS NULL OR evidence_ref = '')
	`, s.tenantID, alertID, evidenceRef)
	return err
}

func RunMigrations(ctx context.Context, pool *pgxpool.Pool, sql string) error {
	_, err := pool.Exec(ctx, sql)
	return err
}
