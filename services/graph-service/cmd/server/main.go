package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/digital-twin/platform/services/graph-service/internal/api"
	"github.com/digital-twin/platform/services/graph-service/internal/config"
	"github.com/digital-twin/platform/services/graph-service/internal/consumer"
	"github.com/digital-twin/platform/services/graph-service/internal/graph"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	cfg := config.Load()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	store, err := graph.NewStore(cfg.Neo4jURI, cfg.Neo4jUser, cfg.Neo4jPassword, cfg.DefaultTenantID)
	if err != nil {
		slog.Error("connect neo4j", "error", err)
		os.Exit(1)
	}
	defer func() { _ = store.Close(context.Background()) }()

	handler := consumer.NewHandler(store)
	twinRunner := consumer.NewRunner(cfg.KafkaBrokers, cfg.ConsumerGroup, cfg.TwinTopic, handler)
	instrumentsRunner := consumer.NewInstrumentsRunner(cfg.KafkaBrokers, cfg.InstrumentsConsumerGroup(), cfg.InstrumentsTopic, handler)
	go func() {
		if err := twinRunner.Run(ctx); err != nil && ctx.Err() == nil {
			slog.Error("twin consumer stopped", "error", err)
		}
	}()
	go func() {
		if err := instrumentsRunner.Run(ctx); err != nil && ctx.Err() == nil {
			slog.Error("instruments consumer stopped", "error", err)
		}
	}()

	srv := api.NewServer(store)
	httpServer := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		slog.Info("graph service listening", "addr", cfg.HTTPAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("http server", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	_ = httpServer.Shutdown(shutdownCtx)
	_ = twinRunner.Close()
	_ = instrumentsRunner.Close()
}
