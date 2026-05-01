package service

import (
	"testing"

	"github.com/him-cyber/infra-inventory-stream/internal/core/domain"
)

func TestBuildTopologyIncludesMissionTeamsAndAssetEdges(t *testing.T) {
	topology := buildTopology(domain.ConfigSnapshot{Version: "v1", ReplayWindow: 64}, []domain.Asset{
		{ID: "laptop-fleet", Type: "endpoint", Name: "Laptop Fleet", Owner: "it", Environment: "prod", Region: "office", Service: "support", Risk: "low", Dependencies: []string{"wifi"}},
		{ID: "wifi", Type: "network", Name: "Branch Wi-Fi", Owner: "it", Environment: "prod", Region: "office", Service: "network", Risk: "medium"},
	}, 2)

	if topology.Story.Mission == "" || topology.Story.Vision == "" || topology.Story.ServiceNowUseCase == "" {
		t.Fatal("expected mission, vision, and ServiceNow use case")
	}
	if topology.Metrics.Assets != 2 || topology.Metrics.ConfigVersion != "v1" || topology.Metrics.RecentEvents != 2 {
		t.Fatalf("unexpected metrics: %+v", topology.Metrics)
	}
	if len(topology.Teams) < 5 {
		t.Fatalf("expected cross-team needs, got %d", len(topology.Teams))
	}
	if !hasEdge(topology.Edges, "laptop-fleet", "wifi", "depends on") {
		t.Fatalf("expected asset dependency edge, got %+v", topology.Edges)
	}
}

func hasEdge(edges []domain.TopologyEdge, from, to, label string) bool {
	for _, edge := range edges {
		if edge.From == from && edge.To == to && edge.Label == label {
			return true
		}
	}
	return false
}
