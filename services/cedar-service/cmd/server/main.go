package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/digital-twin/platform/services/cedar-service/internal/api"
	"github.com/digital-twin/platform/services/cedar-service/internal/audit"
	"github.com/digital-twin/platform/services/cedar-service/internal/config"
	"github.com/digital-twin/platform/services/cedar-service/internal/engine"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	cfg := config.Load()

	policyDir := cfg.PolicyDir
	if _, err := os.Stat(policyDir); err != nil {
		policyDir = engine.PolicyDirFromRepoRoot()
	}

	eng, err := engine.New(policyDir)
	if err != nil {
		slog.Error("load cedar policies", "error", err, "dir", policyDir)
		os.Exit(1)
	}

	pub := audit.NewPublisher(cfg.KafkaBrokers, cfg.AuditTopic, "cedar-service")
	defer pub.Close()

	srv := api.NewServer(eng, pub, api.PrincipalDefaults{
		ID:    cfg.DefaultPrincipal,
		Roles: cfg.DefaultRoles,
	})

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	httpServer := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		slog.Info("cedar service listening", "addr", cfg.HTTPAddr, "policyDir", policyDir)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("http server", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	_ = httpServer.Shutdown(shutdownCtx)
}
