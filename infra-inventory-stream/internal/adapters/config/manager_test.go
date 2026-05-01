package config

import (
	"context"
	"testing"
	"time"

	"github.com/him-cyber/infra-inventory-stream/internal/core/domain"
)

type managerTestSource struct{}

func (managerTestSource) Load(context.Context) (domain.ConfigSnapshot, error) {
	return domain.ConfigSnapshot{
		Version:          "v1",
		AllowedTypes:     []string{"service"},
		SearchFields:     []string{"name"},
		TenantRateLimits: map[string]int{"demo": 60},
		IndexRouting: map[string]string{
			"demo":      "inventory-assets",
			"demo/prod": "inventory-assets-prod",
		},
		ReplayWindow: 10,
		UpdatedAt:    time.Now(),
	}, nil
}

func TestManagerLoadRejectsNilSource(t *testing.T) {
	manager := NewManager(nil)
	if err := manager.Load(context.Background()); err == nil {
		t.Fatal("expected nil source error")
	}
}

func TestManagerRouteIndexFallsBackWhenRoutesAreUnset(t *testing.T) {
	manager := &Manager{}
	if got := manager.RouteIndex("demo", "prod", "inventory"); got != "inventory-assets" {
		t.Fatalf("got %q", got)
	}
}

func TestManagerRouteIndexUsesLongestPrefix(t *testing.T) {
	manager := NewManager(managerTestSource{})
	if err := manager.Load(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := manager.RouteIndex("demo", "prod", "inventory"); got != "inventory-assets-prod" {
		t.Fatalf("got %q", got)
	}
}
