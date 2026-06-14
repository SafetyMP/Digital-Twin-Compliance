package store

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const testTenantID = "00000000-0000-0000-0000-000000000001"

func TestApplyCDCEvent_InsertsPersonaAndOutbox(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	st := newIntegrationStore(t, ctx)
	defer st.pool.Close()

	input := CDCInput{
		IdempotencyKey:  "cdc-integration-1",
		SourceTable:     "legal_entities",
		PersonaID:       "11111111-1111-1111-1111-111111111201",
		SourceEntityID:  "11111111-1111-1111-1111-111111111201",
		PersonaType:     "Institution",
		CurrentState:    json.RawMessage(`{"name":"Integration Bank"}`),
		SourceTimestamp: time.Date(2026, 6, 13, 12, 0, 0, 0, time.UTC),
	}

	persona, err := st.ApplyCDCEvent(ctx, input)
	if err != nil {
		t.Fatalf("ApplyCDCEvent: %v", err)
	}
	if persona.StateVersion != 1 {
		t.Fatalf("stateVersion = %d, want 1", persona.StateVersion)
	}

	rows, err := st.FetchUnpublishedOutbox(ctx, 10)
	if err != nil {
		t.Fatalf("FetchUnpublishedOutbox: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("outbox rows = %d, want 1", len(rows))
	}
	if rows[0].Topic != "twin.state.updated" {
		t.Fatalf("topic = %q", rows[0].Topic)
	}

	processed, err := st.IsProcessed(ctx, input.IdempotencyKey)
	if err != nil {
		t.Fatalf("IsProcessed: %v", err)
	}
	if !processed {
		t.Fatal("expected processed_events row")
	}
}

func TestApplyCDCEvent_IdempotentReplay(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	st := newIntegrationStore(t, ctx)
	defer st.pool.Close()

	input := CDCInput{
		IdempotencyKey:  "cdc-integration-2",
		SourceTable:     "legal_entities",
		PersonaID:       "11111111-1111-1111-1111-111111111202",
		SourceEntityID:  "11111111-1111-1111-1111-111111111202",
		PersonaType:     "Institution",
		CurrentState:    json.RawMessage(`{"name":"Replay Bank"}`),
		SourceTimestamp: time.Date(2026, 6, 13, 12, 0, 0, 0, time.UTC),
	}

	first, err := st.ApplyCDCEvent(ctx, input)
	if err != nil {
		t.Fatalf("first ApplyCDCEvent: %v", err)
	}

	second, err := st.ApplyCDCEvent(ctx, input)
	if err != nil {
		t.Fatalf("replay ApplyCDCEvent: %v", err)
	}
	if second.StateVersion != first.StateVersion {
		t.Fatalf("replay stateVersion = %d, want %d", second.StateVersion, first.StateVersion)
	}

	rows, err := st.FetchUnpublishedOutbox(ctx, 10)
	if err != nil {
		t.Fatalf("FetchUnpublishedOutbox: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("outbox rows after replay = %d, want 1", len(rows))
	}
}

func TestApplyCDCEvent_IncrementsVersionOnUpdate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	st := newIntegrationStore(t, ctx)
	defer st.pool.Close()

	base := CDCInput{
		IdempotencyKey:  "cdc-integration-3a",
		SourceTable:     "legal_entities",
		PersonaID:       "11111111-1111-1111-1111-111111111203",
		SourceEntityID:  "11111111-1111-1111-1111-111111111203",
		PersonaType:     "Institution",
		CurrentState:    json.RawMessage(`{"name":"Version Bank"}`),
		SourceTimestamp: time.Date(2026, 6, 13, 12, 0, 0, 0, time.UTC),
	}
	if _, err := st.ApplyCDCEvent(ctx, base); err != nil {
		t.Fatalf("first upsert: %v", err)
	}

	update := base
	update.IdempotencyKey = "cdc-integration-3b"
	update.CurrentState = json.RawMessage(`{"name":"Version Bank Updated"}`)
	update.SourceTimestamp = time.Date(2026, 6, 13, 13, 0, 0, 0, time.UTC)

	updated, err := st.ApplyCDCEvent(ctx, update)
	if err != nil {
		t.Fatalf("update ApplyCDCEvent: %v", err)
	}
	if updated.StateVersion != 2 {
		t.Fatalf("stateVersion = %d, want 2", updated.StateVersion)
	}
}

func TestMarkOutboxPublished(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	st := newIntegrationStore(t, ctx)
	defer st.pool.Close()

	input := CDCInput{
		IdempotencyKey:  "cdc-integration-4",
		SourceTable:     "legal_entities",
		PersonaID:       "11111111-1111-1111-1111-111111111204",
		SourceEntityID:  "11111111-1111-1111-1111-111111111204",
		PersonaType:     "Institution",
		CurrentState:    json.RawMessage(`{"name":"Publish Bank"}`),
		SourceTimestamp: time.Date(2026, 6, 13, 12, 0, 0, 0, time.UTC),
	}
	if _, err := st.ApplyCDCEvent(ctx, input); err != nil {
		t.Fatalf("ApplyCDCEvent: %v", err)
	}

	rows, err := st.FetchUnpublishedOutbox(ctx, 10)
	if err != nil || len(rows) != 1 {
		t.Fatalf("unpublished rows: %v len=%d", err, len(rows))
	}
	if err := st.MarkOutboxPublished(ctx, rows[0].ID); err != nil {
		t.Fatalf("MarkOutboxPublished: %v", err)
	}

	rows, err = st.FetchUnpublishedOutbox(ctx, 10)
	if err != nil {
		t.Fatalf("FetchUnpublishedOutbox after publish: %v", err)
	}
	if len(rows) != 0 {
		t.Fatalf("expected no unpublished rows, got %d", len(rows))
	}
}

type integrationStore struct {
	*Store
	pool *pgxpool.Pool
}

func newIntegrationStore(t *testing.T, ctx context.Context) *integrationStore {
	t.Helper()

	if url := os.Getenv("STATE_TEST_DB_URL"); url != "" {
		pool, err := pgxpool.New(ctx, url)
		if err != nil {
			t.Fatalf("connect STATE_TEST_DB_URL: %v", err)
		}
		if err := resetStateSchema(ctx, pool); err != nil {
			pool.Close()
			t.Fatalf("reset schema: %v", err)
		}
		return &integrationStore{Store: New(pool, testTenantID), pool: pool}
	}

	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("twin_state"),
		postgres.WithUsername("state"),
		postgres.WithPassword("state"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		if isDockerUnavailable(err) {
			t.Skipf("docker unavailable for store integration tests: %v", err)
		}
		t.Fatalf("start postgres container: %v", err)
	}
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			t.Logf("terminate postgres container: %v", err)
		}
	})

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("connection string: %v", err)
	}
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("connect test postgres: %v", err)
	}
	t.Cleanup(func() { pool.Close() })

	if err := applyStateMigrations(ctx, pool); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
	return &integrationStore{Store: New(pool, testTenantID), pool: pool}
}

func applyStateMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return os.ErrInvalid
	}
	sqlBytes, err := os.ReadFile(filepath.Join(filepath.Dir(filename), "..", "..", "migrations", "001_init.sql"))
	if err != nil {
		return err
	}
	return RunMigrations(ctx, pool, string(sqlBytes))
}

func resetStateSchema(ctx context.Context, pool *pgxpool.Pool) error {
	tables := []string{
		"outbox", "processed_events", "instruments", "accounts", "twin_personas",
	}
	for _, table := range tables {
		if _, err := pool.Exec(ctx, "DROP TABLE IF EXISTS "+table+" CASCADE"); err != nil {
			return err
		}
	}
	return applyStateMigrations(ctx, pool)
}

func isDockerUnavailable(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "Cannot connect to the Docker daemon") ||
		strings.Contains(msg, "docker daemon") ||
		strings.Contains(msg, "permission denied") ||
		strings.Contains(msg, "no such file or directory") ||
		strings.Contains(msg, "connection refused")
}
