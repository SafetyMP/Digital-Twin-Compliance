package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/digital-twin/platform/services/decision-service/internal/api"
	"github.com/digital-twin/platform/services/decision-service/internal/audit"
	"github.com/digital-twin/platform/services/decision-service/internal/config"
	"github.com/digital-twin/platform/services/decision-service/internal/engine"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	cfg := config.Load()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	policyDir := cfg.PolicyDir
	if _, err := os.Stat(policyDir); err != nil {
		policyDir = engine.PolicyDirFromRepoRoot()
	}

	eval, err := engine.NewEvaluator(policyDir)
	if err != nil {
		slog.Error("load zen policies", "error", err, "dir", policyDir)
		os.Exit(1)
	}
	defer eval.Close()

	pub := audit.NewPendingPublisher(cfg.KafkaBrokers, cfg.AuditTopic, "decision-service")
	defer pub.Close()

	srv := api.NewServer(eval, pub)
	httpServer := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		slog.Info("decision service listening", "addr", cfg.HTTPAddr, "policyDir", policyDir)
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
