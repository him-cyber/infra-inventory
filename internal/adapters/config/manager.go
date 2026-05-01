package config

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/him-cyber/infra-inventory-stream/internal/core/domain"
	"github.com/him-cyber/infra-inventory-stream/internal/platform/ds/radix"
)

type Manager struct {
	source Source
	mu     sync.RWMutex
	snap   domain.ConfigSnapshot
	routes *radix.Radix[string]
}

func NewManager(source Source) *Manager {
	return &Manager{source: source, routes: radix.New[string](nil)}
}

func (m *Manager) Load(ctx context.Context) error {
	if m == nil {
		return errors.New("config manager is nil")
	}
	if m.source == nil {
		return errors.New("config source is nil")
	}
	snap, err := m.source.Load(ctx)
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.snap = snap
	m.routes = radix.New[string](snap.IndexRouting)
	return nil
}

func (m *Manager) Watch(ctx context.Context, interval time.Duration, log *slog.Logger) {
	if interval <= 0 {
		interval = 10 * time.Second
	}
	if log == nil {
		log = slog.Default()
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			current := m.Snapshot().Version
			if err := m.Load(ctx); err != nil {
				log.Warn("config reload failed", "error", err)
				continue
			}
			if next := m.Snapshot().Version; next != current {
				log.Info("config hot-reloaded", "from", current, "to", next)
			}
		}
	}
}

func (m *Manager) Snapshot() domain.ConfigSnapshot {
	if m == nil {
		return domain.ConfigSnapshot{}
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.snap
}

func (m *Manager) RouteIndex(tenant, env, service string) string {
	if m == nil {
		return "inventory-assets"
	}
	key := tenant + "/" + env + "/" + service
	m.mu.RLock()
	defer m.mu.RUnlock()
	if index, ok := m.routes.LongestPrefix(key); ok {
		return index
	}
	if index, ok := m.routes.LongestPrefix(tenant); ok {
		return index
	}
	return "inventory-assets"
}

func (m *Manager) IsAllowedType(assetType string) bool {
	snap := m.Snapshot()
	for _, allowed := range snap.AllowedTypes {
		if allowed == assetType {
			return true
		}
	}
	return false
}
