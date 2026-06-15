package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")
var ErrHierarchyDepth = errors.New("legal entity hierarchy exceeds max depth of 3")

type TwinPersona struct {
	PersonaID        string          `json:"personaId"`
	SourceEntityID   string          `json:"sourceEntityId"`
	PersonaType      string          `json:"personaType"`
	StateVersion     int             `json:"stateVersion"`
	CurrentState     json.RawMessage `json:"currentState"`
	ComplianceStatus string          `json:"complianceStatus"`
	LastSyncedAt     time.Time       `json:"lastSyncedAt"`
}

type OutboxRow struct {
	ID           int64
	Topic        string
	PartitionKey string
	Payload      json.RawMessage
}

type Store struct {
	pool     *pgxpool.Pool
	tenantID string
}

func New(pool *pgxpool.Pool, tenantID string) *Store {
	return &Store{pool: pool, tenantID: tenantID}
}

func (s *Store) GetPersona(ctx context.Context, personaID string) (TwinPersona, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT persona_id, source_entity_id, persona_type, state_version,
		       current_state, compliance_status, last_synced_at
		FROM twin_personas
		WHERE tenant_id = $1 AND persona_id = $2
	`, s.tenantID, personaID)

	var p TwinPersona
	if err := row.Scan(
		&p.PersonaID, &p.SourceEntityID, &p.PersonaType, &p.StateVersion,
		&p.CurrentState, &p.ComplianceStatus, &p.LastSyncedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return TwinPersona{}, ErrNotFound
		}
		return TwinPersona{}, err
	}
	return p, nil
}

func (s *Store) ListPersonas(ctx context.Context, personaType string, limit, offset int) ([]TwinPersona, error) {
	query := `
		SELECT persona_id, source_entity_id, persona_type, state_version,
		       current_state, compliance_status, last_synced_at
		FROM twin_personas
		WHERE tenant_id = $1
	`
	args := []any{s.tenantID}
	if personaType != "" {
		query += " AND persona_type = $2"
		args = append(args, personaType)
		query += " ORDER BY updated_at DESC LIMIT $3 OFFSET $4"
		args = append(args, limit, offset)
	} else {
		query += " ORDER BY updated_at DESC LIMIT $2 OFFSET $3"
		args = append(args, limit, offset)
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []TwinPersona
	for rows.Next() {
		var p TwinPersona
		if err := rows.Scan(
			&p.PersonaID, &p.SourceEntityID, &p.PersonaType, &p.StateVersion,
			&p.CurrentState, &p.ComplianceStatus, &p.LastSyncedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) IsProcessed(ctx context.Context, idempotencyKey string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM processed_events WHERE idempotency_key = $1)`,
		idempotencyKey,
	).Scan(&exists)
	return exists, err
}

func (s *Store) ApplyCDCEvent(ctx context.Context, input CDCInput) (TwinPersona, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return TwinPersona{}, err
	}
	defer tx.Rollback(ctx)

	var already bool
	if err := tx.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM processed_events WHERE idempotency_key = $1)`,
		input.IdempotencyKey,
	).Scan(&already); err != nil {
		return TwinPersona{}, err
	}
	if already {
		return s.GetPersona(ctx, input.PersonaID)
	}

	var persona TwinPersona
	err = tx.QueryRow(ctx, `
		INSERT INTO twin_personas (
			persona_id, tenant_id, source_entity_id, persona_type,
			state_version, current_state, compliance_status, last_synced_at, updated_at
		) VALUES ($1, $2, $3, $4, 1, $5, 'Unknown', $6, now())
		ON CONFLICT (tenant_id, source_entity_id, persona_type) DO UPDATE SET
			state_version = twin_personas.state_version + 1,
			current_state = EXCLUDED.current_state,
			last_synced_at = EXCLUDED.last_synced_at,
			updated_at = now()
		RETURNING persona_id, source_entity_id, persona_type, state_version,
		          current_state, compliance_status, last_synced_at
	`, input.PersonaID, s.tenantID, input.SourceEntityID, input.PersonaType,
		input.CurrentState, input.SourceTimestamp,
	).Scan(
		&persona.PersonaID, &persona.SourceEntityID, &persona.PersonaType, &persona.StateVersion,
		&persona.CurrentState, &persona.ComplianceStatus, &persona.LastSyncedAt,
	)
	if err != nil {
		return TwinPersona{}, err
	}

	switch input.SourceTable {
	case "legal_entities":
		if err := upsertLegalEntityMirror(ctx, tx, s.tenantID, input); err != nil {
			return TwinPersona{}, err
		}
	case "accounts":
		if err := upsertAccountMirror(ctx, tx, s.tenantID, input); err != nil {
			return TwinPersona{}, err
		}
	case "instruments":
		if err := upsertInstrumentMirror(ctx, tx, s.tenantID, input); err != nil {
			return TwinPersona{}, err
		}
	}

	outboxPayload, err := json.Marshal(map[string]any{
		"personaId":        persona.PersonaID,
		"personaType":      persona.PersonaType,
		"sourceEntityId":   persona.SourceEntityID,
		"stateVersion":     persona.StateVersion,
		"complianceStatus": persona.ComplianceStatus,
		"lastSyncedAt":     persona.LastSyncedAt.UTC().Format(time.RFC3339Nano),
		"currentState":     json.RawMessage(persona.CurrentState),
	})
	if err != nil {
		return TwinPersona{}, err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO outbox (topic, partition_key, payload)
		VALUES ($1, $2, $3)
	`, "twin.state.updated", persona.PersonaID, outboxPayload); err != nil {
		return TwinPersona{}, err
	}

	if _, err := tx.Exec(ctx,
		`INSERT INTO processed_events (idempotency_key) VALUES ($1)`,
		input.IdempotencyKey,
	); err != nil {
		return TwinPersona{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return TwinPersona{}, err
	}
	return persona, nil
}

type CDCInput struct {
	IdempotencyKey  string
	SourceTable     string
	PersonaID       string
	SourceEntityID  string
	PersonaType     string
	CurrentState    json.RawMessage
	SourceTimestamp time.Time
}

func upsertLegalEntityMirror(ctx context.Context, tx pgx.Tx, tenantID string, input CDCInput) error {
	return nil
}

func upsertAccountMirror(ctx context.Context, tx pgx.Tx, tenantID string, input CDCInput) error {
	var row map[string]any
	if err := json.Unmarshal(input.CurrentState, &row); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, `
		INSERT INTO accounts (account_id, tenant_id, account_number, account_type, currency, owner_entity_id, status, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, now())
		ON CONFLICT (account_id) DO UPDATE SET
			account_number = EXCLUDED.account_number,
			account_type = EXCLUDED.account_type,
			currency = EXCLUDED.currency,
			owner_entity_id = EXCLUDED.owner_entity_id,
			status = EXCLUDED.status,
			updated_at = now()
	`, input.PersonaID, tenantID,
		stringField(row, "account_number"),
		stringField(row, "account_type"),
		stringField(row, "currency"),
		stringField(row, "owner_entity_id"),
		stringField(row, "status"),
	)
	return err
}

func upsertInstrumentMirror(ctx context.Context, tx pgx.Tx, tenantID string, input CDCInput) error {
	var row map[string]any
	if err := json.Unmarshal(input.CurrentState, &row); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, `
		INSERT INTO instruments (instrument_id, tenant_id, isin, instrument_type, notional_amount, currency, maturity_date, regulatory_class, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, now())
		ON CONFLICT (instrument_id) DO UPDATE SET
			isin = EXCLUDED.isin,
			instrument_type = EXCLUDED.instrument_type,
			notional_amount = EXCLUDED.notional_amount,
			currency = EXCLUDED.currency,
			maturity_date = EXCLUDED.maturity_date,
			regulatory_class = EXCLUDED.regulatory_class,
			updated_at = now()
	`, input.PersonaID, tenantID,
		stringField(row, "isin"),
		stringField(row, "instrument_type"),
		numericField(row, "notional_amount"),
		stringField(row, "currency"),
		dateField(row, "maturity_date"),
		stringField(row, "regulatory_class"),
	)
	return err
}

func stringField(row map[string]any, key string) string {
	v, ok := row[key]
	if !ok || v == nil {
		return ""
	}
	return fmt.Sprint(v)
}

func numericField(row map[string]any, key string) string {
	v, ok := row[key]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(t), 'f', -1, 32)
	default:
		return stringField(row, key)
	}
}

func dateField(row map[string]any, key string) *string {
	v, ok := row[key]
	if !ok || v == nil {
		return nil
	}
	switch t := v.(type) {
	case string:
		if t == "" {
			return nil
		}
		return &t
	case float64:
		s := debeziumDaysToDate(int(t))
		return &s
	case int:
		s := debeziumDaysToDate(t)
		return &s
	case int64:
		s := debeziumDaysToDate(int(t))
		return &s
	default:
		s := stringField(row, key)
		if s == "" {
			return nil
		}
		return &s
	}
}

func debeziumDaysToDate(days int) string {
	return time.Unix(int64(days)*86400, 0).UTC().Format("2006-01-02")
}

func (s *Store) FetchUnpublishedOutbox(ctx context.Context, limit int) ([]OutboxRow, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, topic, partition_key, payload
		FROM outbox
		WHERE published_at IS NULL
		ORDER BY id
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []OutboxRow
	for rows.Next() {
		var r OutboxRow
		if err := rows.Scan(&r.ID, &r.Topic, &r.PartitionKey, &r.Payload); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) MarkOutboxPublished(ctx context.Context, id int64) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE outbox SET published_at = now() WHERE id = $1`, id)
	return err
}

func RunMigrations(ctx context.Context, pool *pgxpool.Pool, sql string) error {
	_, err := pool.Exec(ctx, sql)
	return err
}
