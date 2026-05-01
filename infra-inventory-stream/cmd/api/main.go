package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/him-cyber/infra-inventory-stream/internal/adapters/auth"
	"github.com/him-cyber/infra-inventory-stream/internal/adapters/config"
	"github.com/him-cyber/infra-inventory-stream/internal/adapters/httpapi"
	"github.com/him-cyber/infra-inventory-stream/internal/adapters/kafka"
	"github.com/him-cyber/infra-inventory-stream/internal/adapters/opensearch"
	"github.com/him-cyber/infra-inventory-stream/internal/core/domain"
	"github.com/him-cyber/infra-inventory-stream/internal/core/service"
	"github.com/him-cyber/infra-inventory-stream/internal/platform/ds/ratelimit"
	"github.com/him-cyber/infra-inventory-stream/internal/platform/ds/ring"
	"github.com/him-cyber/infra-inventory-stream/internal/platform/observability"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	shutdownTrace := observability.Init(ctx, "inventory-api", log)
	defer shutdownTrace(context.Background())

	cfg := config.NewManager(config.SourceFromEnv())
	if err := cfg.Load(ctx); err != nil {
		log.Error("load config", "error", err)
		os.Exit(1)
	}
	go cfg.Watch(ctx, 10*time.Second, log)
	authManager, err := auth.NewFromEnv(ctx)
	if err != nil {
		log.Error("auth setup", "error", err)
		os.Exit(1)
	}

	producer, err := kafka.NewClient("")
	if err != nil {
		log.Error("kafka client", "error", err)
		os.Exit(1)
	}
	defer producer.Close()

	store, err := opensearch.NewStore()
	if err != nil {
		log.Error("opensearch client", "error", err)
		os.Exit(1)
	}
	if err := store.EnsureIndex(ctx); err != nil {
		log.Warn("index bootstrap failed", "error", err)
	}

	replayWindow := cfg.Snapshot().ReplayWindow
	inventory := service.NewService(cfg, producer, store, ring.New[domain.Event](replayWindow), ratelimit.NewLimiter(), kafka.Topic())
	server := &http.Server{
		Addr:              ":" + valueOrDefault(os.Getenv("PORT"), "8080"),
		Handler:           httpapi.New(inventory, log, authManager),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Info("api listening", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("api failed", "error", err)
			stop()
		}
	}()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
}

func valueOrDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
