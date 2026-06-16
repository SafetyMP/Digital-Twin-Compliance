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

func TestUpsertAlert_Idempotent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	st := newTestStore(t, ctx)
	defer st.pool.Close()

	detectedAt := time.Date(2026, 6, 13, 18, 45, 0, 0, time.UTC)
	input := UpsertInput{
		AlertID:        "550e8400-e29b-41d4-a716-446655440000",
		RuleCode:       "INT-M001",
		Regime:         "Internal",
		Severity:       "Warning",
		Status:         "Open",
		PersonaID:      "660e8400-e29b-41d4-a716-446655440001",
		PersonaType:    "Account",
		Summary:        "Velocity breach",
		Details:        json.RawMessage(`{"count":"51"}`),
		DetectedAt:     detectedAt,
		IdempotencyKey: "idem-upsert-1",
	}

	first, created, err := st.UpsertAlert(ctx, input)
	if err != nil {
		t.Fatalf("first upsert: %v", err)
	}
	if !created {
		t.Fatal("expected first upsert to create row")
	}
	if first.AlertID != input.AlertID {
		t.Fatalf("alertId = %q, want %q", first.AlertID, input.AlertID)
	}
	if first.IdempotencyKey != input.IdempotencyKey {
		t.Fatalf("idempotencyKey = %q", first.IdempotencyKey)
	}

	dupInput := input
	dupInput.AlertID = "770e8400-e29b-41d4-a716-446655440099"
	dupInput.Summary = "should not overwrite"

	second, created, err := st.UpsertAlert(ctx, dupInput)
	if err != nil {
		t.Fatalf("duplicate upsert: %v", err)
	}
	if created {
		t.Fatal("expected duplicate upsert to return existing row")
	}
	if second.AlertID != first.AlertID {
		t.Fatalf("duplicate alertId = %q, want %q", second.AlertID, first.AlertID)
	}
	if second.Summary != first.Summary {
		t.Fatalf("duplicate summary = %q, want %q", second.Summary, first.Summary)
	}

	got, err := st.GetByIdempotencyKey(ctx, input.IdempotencyKey)
	if err != nil {
		t.Fatalf("get by idempotency key: %v", err)
	}
	if got.AlertID != first.AlertID {
		t.Fatalf("stored alertId = %q", got.AlertID)
	}
}

func TestAcknowledge(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	st := newTestStore(t, ctx)
	defer st.pool.Close()

	input := UpsertInput{
		AlertID:        "550e8400-e29b-41d4-a716-446655440010",
		RuleCode:       "INT-M002",
		Regime:         "Internal",
		Severity:       "Critical",
		Status:         "Open",
		PersonaID:      "660e8400-e29b-41d4-a716-446655440001",
		PersonaType:    "Account",
		Summary:        "Exposure breach",
		DetectedAt:     time.Date(2026, 6, 13, 19, 0, 0, 0, time.UTC),
		IdempotencyKey: "idem-ack-1",
	}
	if _, _, err := st.UpsertAlert(ctx, input); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	ack, err := st.Acknowledge(ctx, input.AlertID, "analyst@example.com")
	if err != nil {
		t.Fatalf("acknowledge: %v", err)
	}
	if ack.Status != "Acknowledged" {
		t.Fatalf("status = %q, want Acknowledged", ack.Status)
	}
	if ack.AcknowledgedAt == nil {
		t.Fatal("expected acknowledgedAt to be set")
	}
	if ack.AcknowledgedBy == nil || *ack.AcknowledgedBy != "analyst@example.com" {
		t.Fatalf("acknowledgedBy = %v", ack.AcknowledgedBy)
	}

	got, err := st.GetAlert(ctx, input.AlertID)
	if err != nil {
		t.Fatalf("get alert: %v", err)
	}
	if got.Status != "Acknowledged" {
		t.Fatalf("stored status = %q", got.Status)
	}

	_, err = st.Acknowledge(ctx, "00000000-0000-0000-0000-000000000099", "analyst@example.com")
	if err != ErrNotFound {
		t.Fatalf("missing alert err = %v, want ErrNotFound", err)
	}
}

func newTestStore(t *testing.T, ctx context.Context) *Store {
	t.Helper()

	if url := os.Getenv("ALERT_TEST_DB_URL"); url != "" {
		pool, err := pgxpool.New(ctx, url)
		if err != nil {
			t.Fatalf("connect ALERT_TEST_DB_URL: %v", err)
		}
		if err := resetSchema(ctx, pool); err != nil {
			pool.Close()
			t.Fatalf("reset schema: %v", err)
		}
		return New(pool, testTenantID)
	}

	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("alerts"),
		postgres.WithUsername("alert"),
		postgres.WithPassword("alert"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		if isDockerUnavailable(err) {
			t.Skipf("docker unavailable for store tests: %v", err)
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

	if err := applyMigrations(ctx, pool); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
	return New(pool, testTenantID)
}

func applyMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return os.ErrInvalid
	}
	migDir := filepath.Join(filepath.Dir(filename), "..", "..", "migrations")
	var sql strings.Builder
	for _, name := range []string{"001_alerts.sql", "002_evidence_ref.sql"} {
		sqlBytes, err := os.ReadFile(filepath.Join(migDir, name))
		if err != nil {
			return err
		}
		sql.Write(sqlBytes)
		sql.WriteString("\n")
	}
	return RunMigrations(ctx, pool, sql.String())
}

func resetSchema(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `DROP TABLE IF EXISTS compliance_alerts`)
	if err != nil {
		return err
	}
	return applyMigrations(ctx, pool)
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
