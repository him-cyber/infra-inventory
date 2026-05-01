package service

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/him-cyber/infra-inventory-stream/internal/adapters/config"
	"github.com/him-cyber/infra-inventory-stream/internal/adapters/kafka"
	"github.com/him-cyber/infra-inventory-stream/internal/core/domain"
	"github.com/him-cyber/infra-inventory-stream/internal/core/security"
	"github.com/him-cyber/infra-inventory-stream/internal/platform/ds/ratelimit"
	"github.com/him-cyber/infra-inventory-stream/internal/platform/ds/ring"
	"github.com/him-cyber/infra-inventory-stream/internal/platform/observability"
)

type SearchStore interface {
	GetAsset(context.Context, string) (domain.Asset, bool, error)
	Search(context.Context, domain.SearchQuery, []string) (domain.SearchResult, error)
}

type Service struct {
	config     *config.Manager
	publisher  kafka.Publisher
	store      SearchStore
	recent     *ring.Buffer[domain.Event]
	limits     *ratelimit.Limiter
	topic      string
	orgMu      sync.RWMutex
	orgImports []domain.OrgImportResult
}

func NewService(cfg *config.Manager, pub kafka.Publisher, store SearchStore, recent *ring.Buffer[domain.Event], limits *ratelimit.Limiter, topic string) *Service {
	return &Service{config: cfg, publisher: pub, store: store, recent: recent, limits: limits, topic: topic}
}

func (s *Service) UpsertAsset(ctx context.Context, tenant string, asset domain.Asset) (domain.Event, error) {
	if asset.ID == "" {
		asset.ID = uuid.NewString()
	}
	if asset.UpdatedAt.IsZero() {
		asset.UpdatedAt = time.Now().UTC()
	}
	if !s.config.IsAllowedType(asset.Type) {
		return domain.Event{}, errors.New("asset type is not allowed by active config")
	}
	if err := security.NewDependencyGuard(s.store).ValidateUpsert(ctx, asset); err != nil {
		return domain.Event{}, err
	}
	snap := s.config.Snapshot()
	limit := snap.TenantRateLimits[tenant]
	if !s.limits.Allow(tenant, limit) {
		return domain.Event{}, errors.New("tenant rate limit exceeded")
	}
	event := domain.Event{
		EventID:       uuid.NewString(),
		AssetID:       asset.ID,
		EventType:     "asset.upserted",
		Version:       asset.Version,
		Timestamp:     time.Now().UTC(),
		Payload:       asset,
		ConfigVersion: snap.Version,
	}
	data, err := json.Marshal(event)
	if err != nil {
		return domain.Event{}, err
	}
	if err := s.publisher.Publish(ctx, s.topic, []byte(event.AssetID), data); err != nil {
		return domain.Event{}, err
	}
	s.recent.Push(event)
	observability.AssetsWritten.Inc()
	return event, nil
}

func (s *Service) Search(ctx context.Context, q domain.SearchQuery) (domain.SearchResult, error) {
	return s.store.Search(ctx, q, s.config.Snapshot().SearchFields)
}

func (s *Service) Topology(ctx context.Context) (domain.SupportTopology, error) {
	snap := s.config.Snapshot()
	result, err := s.store.Search(ctx, domain.SearchQuery{}, snap.SearchFields)
	if err != nil {
		return domain.SupportTopology{}, err
	}
	return buildTopology(snap, result.Assets, len(s.recent.Snapshot())), nil
}

func (s *Service) GetAsset(ctx context.Context, id string) (domain.Asset, bool, error) {
	return s.store.GetAsset(ctx, id)
}

func (s *Service) RecentEvents() []domain.Event {
	return s.recent.Snapshot()
}

func (s *Service) Replay(ctx context.Context) (int, error) {
	events := s.recent.Snapshot()
	for _, event := range events {
		data, err := json.Marshal(event)
		if err != nil {
			return 0, err
		}
		if err := s.publisher.Publish(ctx, s.topic, []byte(event.AssetID), data); err != nil {
			return 0, err
		}
		observability.EventsReplay.Inc()
	}
	return len(events), nil
}

func (s *Service) ConfigVersion() map[string]any {
	snap := s.config.Snapshot()
	return map[string]any{"version": snap.Version, "updated_at": snap.UpdatedAt, "replay_window": snap.ReplayWindow}
}
