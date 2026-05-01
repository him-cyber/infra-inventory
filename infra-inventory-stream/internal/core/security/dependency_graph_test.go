package security

import (
	"context"
	"errors"
	"testing"

	"github.com/him-cyber/infra-inventory-stream/internal/core/domain"
)

type assetMap map[string]domain.Asset

func (m assetMap) GetAsset(_ context.Context, id string) (domain.Asset, bool, error) {
	asset, ok := m[id]
	return asset, ok, nil
}

func TestDependencyGuardRejectsSelfCycle(t *testing.T) {
	guard := NewDependencyGuard(assetMap{})
	err := guard.ValidateUpsert(context.Background(), domain.Asset{ID: "api", Dependencies: []string{"api"}})
	if !errors.Is(err, ErrDependencyCycle) {
		t.Fatalf("expected cycle error, got %v", err)
	}
}

func TestDependencyGuardRejectsStoredCycle(t *testing.T) {
	guard := NewDependencyGuard(assetMap{
		"worker": {ID: "worker", Dependencies: []string{"queue"}},
		"queue":  {ID: "queue", Dependencies: []string{"api"}},
	})
	err := guard.ValidateUpsert(context.Background(), domain.Asset{ID: "api", Dependencies: []string{"worker"}})
	if !errors.Is(err, ErrDependencyCycle) {
		t.Fatalf("expected stored cycle error, got %v", err)
	}
}

func TestDependencyGuardAllowsAcyclicGraph(t *testing.T) {
	guard := NewDependencyGuard(assetMap{
		"queue": {ID: "queue", Dependencies: []string{"index"}},
		"index": {ID: "index"},
	})
	err := guard.ValidateUpsert(context.Background(), domain.Asset{ID: "api", Dependencies: []string{"queue"}})
	if err != nil {
		t.Fatalf("expected acyclic graph to pass, got %v", err)
	}
}
