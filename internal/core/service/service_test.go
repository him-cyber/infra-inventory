package service

import (
	"context"
	"testing"
	"time"

	"github.com/him-cyber/infra-inventory-stream/internal/adapters/config"
	"github.com/him-cyber/infra-inventory-stream/internal/core/domain"
	"github.com/him-cyber/infra-inventory-stream/internal/platform/ds/ratelimit"
	"github.com/him-cyber/infra-inventory-stream/internal/platform/ds/ring"
)

type memoryPublisher struct {
	count int
}

func (p *memoryPublisher) Publish(context.Context, string, []byte, []byte) error {
	p.count++
	return nil
}

type memoryStore struct{}

func (memoryStore) GetAsset(context.Context, string) (domain.Asset, bool, error) {
	return domain.Asset{}, false, nil
}

func (memoryStore) Search(context.Context, domain.SearchQuery, []string) (domain.SearchResult, error) {
	return domain.SearchResult{}, nil
}

type staticSource struct{}

func (staticSource) Load(context.Context) (domain.ConfigSnapshot, error) {
	return domain.ConfigSnapshot{
		Version:          "v1",
		AllowedTypes:     []string{"service"},
		SearchFields:     []string{"name"},
		TenantRateLimits: map[string]int{"demo": 10},
		IndexRouting:     map[string]string{"demo": "inventory"},
		ReplayWindow:     8,
		UpdatedAt:        time.Now(),
	}, nil
}

func TestUpsertPublishesAndRecordsRecentEvent(t *testing.T) {
	cfg := config.NewManager(staticSource{})
	if err := cfg.Load(context.Background()); err != nil {
		t.Fatal(err)
	}
	pub := &memoryPublisher{}
	service := NewService(cfg, pub, memoryStore{}, ring.New[domain.Event](8), ratelimit.NewLimiter(), "inventory.events")
	event, err := service.UpsertAsset(context.Background(), "demo", domain.Asset{Type: "service", Name: "Catalog", Version: 1})
	if err != nil {
		t.Fatal(err)
	}
	if event.ConfigVersion != "v1" || pub.count != 1 {
		t.Fatalf("unexpected event=%+v published=%d", event, pub.count)
	}
	if len(service.RecentEvents()) != 1 {
		t.Fatal("expected recent event")
	}
}
