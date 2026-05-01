package config

import (
	"testing"
	"time"

	"github.com/him-cyber/infra-inventory-stream/internal/core/domain"
)

func TestValidateRejectsBadReplayWindow(t *testing.T) {
	err := Validate(domain.ConfigSnapshot{
		Version:      "v1",
		AllowedTypes: []string{"service"},
		SearchFields: []string{"name"},
		ReplayWindow: 0,
		UpdatedAt:    time.Now(),
	})
	if err == nil {
		t.Fatal("expected replay_window validation error")
	}
}

func TestParseAddsUpdatedAt(t *testing.T) {
	snap, err := parse([]byte(`{"version":"v1","allowed_asset_types":["service"],"search_fields":["name"],"tenant_rate_limits":{"demo":10},"index_routing":{"demo":"inventory"},"replay_window":5}`))
	if err != nil {
		t.Fatal(err)
	}
	if snap.UpdatedAt.IsZero() {
		t.Fatal("expected UpdatedAt to be set")
	}
}
