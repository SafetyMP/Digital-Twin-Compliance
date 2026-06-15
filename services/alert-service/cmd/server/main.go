package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/digital-twin/platform/services/alert-service/internal/api"
	"github.com/digital-twin/platform/services/alert-service/internal/config"
	"github.com/digital-twin/platform/services/alert-service/internal/consumer"
	"github.com/digital-twin/platform/services/alert-service/internal/hub"
	"github.com/digital-twin/platform/services/alert-service/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	cfg := config.Load()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.AlertDBURL)
	if err != nil {
		slog.Error("connect alert db", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := runMigrations(ctx, pool, migrationSearchPaths()); err != nil {
		slog.Error("run migrations", "error", err)
		os.Exit(1)
	}

	st := store.New(pool, cfg.DefaultTenantID)
	wsHub := hub.New()
	handler := consumer.NewHandler(st, wsHub)
	runner := consumer.NewRunner(cfg.KafkaBrokers, cfg.ConsumerGroup, cfg.AlertsTopic, cfg.AlertsDLQTopic, handler)

	go func() {
		if err := runner.Run(ctx); err != nil && ctx.Err() == nil {
			slog.Error("consumer stopped", "error", err)
		}
	}()

	srv := api.NewServer(st, wsHub, cfg.WSSPath)
	httpServer := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		slog.Info("alert service listening", "addr", cfg.HTTPAddr)
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
	return []string{"migrations/001_alerts.sql", "/app/migrations/001_alerts.sql"}
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
	sql, err := readMigrationSQL(paths)
	if err != nil {
		return err
	}
	return store.RunMigrations(ctx, pool, sql)
}
