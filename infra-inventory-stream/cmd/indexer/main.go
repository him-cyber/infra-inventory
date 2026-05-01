package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/him-cyber/infra-inventory-stream/internal/adapters/kafka"
	"github.com/him-cyber/infra-inventory-stream/internal/adapters/opensearch"
	"github.com/him-cyber/infra-inventory-stream/internal/core/domain"
	"github.com/him-cyber/infra-inventory-stream/internal/platform/ds/retry"
	"github.com/him-cyber/infra-inventory-stream/internal/platform/observability"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	shutdownTrace := observability.Init(ctx, "inventory-indexer", log)
	defer shutdownTrace(context.Background())

	consumer, err := kafka.NewClient(valueOrDefault(os.Getenv("KAFKA_GROUP"), "inventory-indexer"), kafka.Topic())
	if err != nil {
		log.Error("kafka client", "error", err)
		os.Exit(1)
	}
	defer consumer.Close()

	store, err := opensearch.NewStore()
	if err != nil {
		log.Error("opensearch client", "error", err)
		os.Exit(1)
	}
	if err := store.EnsureIndex(ctx); err != nil {
		log.Warn("index bootstrap failed", "error", err)
	}

	retries := retry.NewQueue[domain.Event]()
	for ctx.Err() == nil {
		for event, attempts, ok := retries.PopReady(time.Now()); ok; event, attempts, ok = retries.PopReady(time.Now()) {
			if err := store.UpsertAsset(ctx, event.Payload); err != nil {
				retries.Schedule(event, attempts+1, kafka.Backoff(attempts+1))
				observability.IndexFailures.Inc()
				continue
			}
			observability.IndexedEvents.Inc()
		}
		err := consumer.Poll(ctx, func(_, value []byte) error {
			var event domain.Event
			if err := json.Unmarshal(value, &event); err != nil {
				return err
			}
			if err := store.UpsertAsset(ctx, event.Payload); err != nil {
				retries.Schedule(event, 1, kafka.Backoff(1))
				observability.IndexFailures.Inc()
				return nil
			}
			observability.IndexedEvents.Inc()
			return nil
		})
		if err != nil && ctx.Err() == nil {
			log.Warn("poll failed", "error", err)
			time.Sleep(time.Second)
		}
	}
}

func valueOrDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
