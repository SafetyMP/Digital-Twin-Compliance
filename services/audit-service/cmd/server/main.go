package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/digital-twin/platform/services/audit-service/internal/api"
	"github.com/digital-twin/platform/services/audit-service/internal/config"
	"github.com/digital-twin/platform/services/audit-service/internal/consumer"
	"github.com/digital-twin/platform/services/audit-service/internal/immudb"
	"github.com/digital-twin/platform/services/audit-service/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	cfg := config.Load()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.AuditDBURL)
	if err != nil {
		slog.Error("connect audit db", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := runMigrations(ctx, pool, migrationSearchPaths()); err != nil {
		slog.Error("run migrations", "error", err)
		os.Exit(1)
	}

	ledger, err := immudb.Connect(ctx, cfg.ImmuDBHost, cfg.ImmuDBPort, cfg.ImmuDBDatabase, cfg.ImmuDBUser, cfg.ImmuDBPassword)
	if err != nil {
		slog.Error("connect immudb", "error", err)
		os.Exit(1)
	}
	defer ledger.Close()

	st := store.New(pool, cfg.DefaultTenantID)
	if err := reconcileLedgerHead(ctx, st, ledger); err != nil {
		slog.Error("reconcile ledger head", "error", err)
		os.Exit(1)
	}

	recorded := consumer.NewRecordedProducer(cfg.KafkaBrokers, cfg.RecordedTopic, "audit-service")
	defer recorded.Close()

	handler := consumer.NewHandler(ledger, st, recorded, "audit-service")
	runner := consumer.NewRunner(cfg.KafkaBrokers, cfg.ConsumerGroup, cfg.PendingTopic, cfg.PendingDLQTopic, handler, st)

	go func() {
		if err := runner.Run(ctx); err != nil && ctx.Err() == nil {
			slog.Error("consumer stopped", "error", err)
		}
	}()

	srv := api.NewServer(st, ledger)
	httpServer := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		slog.Info("audit service listening", "addr", cfg.HTTPAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("http server", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	_ = httpServer.Shutdown(shutdownCtx)
	_ = runner.Close()
}

func migrationSearchPaths() []string {
	return []string{
		"migrations/001_audit.sql",
		"migrations/002_subject_id_text.sql",
		"/app/migrations/001_audit.sql",
		"/app/migrations/002_subject_id_text.sql",
	}
}

func readMigrationSQL(paths []string) (string, error) {
	for _, path := range paths {
		sqlBytes, err := os.ReadFile(path)
		if err == nil {
			return string(sqlBytes), nil
		}
		if !os.IsNotExist(err) {
			return "", err
		}
	}
	return "", os.ErrNotExist
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool, paths []string) error {
	seen := make(map[string]struct{})
	for _, path := range paths {
		base := filepath.Base(path)
		if _, ok := seen[base]; ok {
			continue
		}
		sqlBytes, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
		seen[base] = struct{}{}
		if err := store.RunMigrations(ctx, pool, string(sqlBytes)); err != nil {
			return fmt.Errorf("%s: %w", base, err)
		}
	}
	return nil
}

func reconcileLedgerHead(ctx context.Context, st *store.Store, ledger *immudb.Client) error {
	count, err := st.IndexCount(ctx)
	if err != nil {
		return err
	}
	head, err := ledger.GetHead(ctx)
	if err != nil {
		return err
	}
	if count == 0 && head.LastSequence > 0 {
		slog.Warn("resetting immudb head; postgres audit index is empty", "immudbSequence", head.LastSequence)
		return ledger.ResetHead(ctx)
	}
	return nil
}
