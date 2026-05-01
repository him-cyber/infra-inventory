package security

import (
	"context"
	"errors"
	"fmt"

	"github.com/him-cyber/infra-inventory-stream/internal/core/domain"
)

const MaxDependencyDepth = 64

var (
	ErrDependencyCycle = errors.New("dependency graph contains a cycle")
	ErrDependencyDepth = errors.New("dependency graph exceeds maximum traversal depth")
)

type AssetReader interface {
	GetAsset(context.Context, string) (domain.Asset, bool, error)
}

type DependencyGuard struct {
	reader AssetReader
}

func NewDependencyGuard(reader AssetReader) *DependencyGuard {
	return &DependencyGuard{reader: reader}
}

func (g *DependencyGuard) ValidateUpsert(ctx context.Context, asset domain.Asset) error {
	if asset.ID == "" {
		return nil
	}

	graph := map[string][]string{asset.ID: asset.Dependencies}
	visiting := make(map[string]bool)
	visited := make(map[string]bool)

	var dfs func(id string, depth int) error
	dfs = func(id string, depth int) error {
		if depth > MaxDependencyDepth {
			return fmt.Errorf("%w at asset %q", ErrDependencyDepth, id)
		}
		if visiting[id] {
			return fmt.Errorf("%w at asset %q", ErrDependencyCycle, id)
		}
		if visited[id] {
			return nil
		}

		visiting[id] = true
		deps, err := g.dependencies(ctx, graph, id)
		if err != nil {
			return err
		}
		for _, dep := range deps {
			if dep == "" {
				continue
			}
			if err := dfs(dep, depth+1); err != nil {
				return err
			}
		}
		visiting[id] = false
		visited[id] = true
		return nil
	}

	return dfs(asset.ID, 0)
}

func (g *DependencyGuard) dependencies(ctx context.Context, graph map[string][]string, id string) ([]string, error) {
	if deps, ok := graph[id]; ok {
		return deps, nil
	}
	if g == nil || g.reader == nil {
		return nil, nil
	}
	asset, ok, err := g.reader.GetAsset(ctx, id)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	graph[id] = asset.Dependencies
	return asset.Dependencies, nil
}
