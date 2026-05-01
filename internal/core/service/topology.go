package service

import (
	"fmt"
	"sort"

	"github.com/him-cyber/infra-inventory-stream/internal/core/domain"
)

func buildTopology(config domain.ConfigSnapshot, assets []domain.Asset, recentEvents int) domain.SupportTopology {
	sort.SliceStable(assets, func(i, j int) bool {
		return assets[i].Name < assets[j].Name
	})

	nodes := []domain.TopologyNode{
		{ID: "office-collector", Label: "Office Collector", Kind: "network", Owner: "it-ops", Risk: "low", Detail: "normalizes branch endpoint signals", Tone: "blue"},
		{ID: "go-api", Label: "Go API", Kind: "service", Owner: "core-infra", Risk: "medium", Detail: "rate limit, auth hook, DFS dependency guard", Tone: "violet"},
		{ID: "kafka-stream", Label: "Kafka Stream", Kind: "queue", Owner: "data", Risk: "low", Detail: "replayable asset.upserted events", Tone: "green"},
		{ID: "indexer", Label: "Indexer Worker", Kind: "worker", Owner: "search", Risk: "medium", Detail: "writes asset documents to OpenSearch", Tone: "violet"},
		{ID: "opensearch", Label: "OpenSearch", Kind: "database", Owner: "search", Risk: "medium", Detail: "schema-less support search index", Tone: "green"},
		{ID: "servicenow-view", Label: "ServiceNow Support View", Kind: "frontend", Owner: "product", Risk: "low", Detail: "CMDB, incident, and change evidence", Tone: "rose"},
	}

	edges := []domain.TopologyEdge{
		{From: "office-collector", To: "go-api", Label: "POST /api/assets"},
		{From: "go-api", To: "kafka-stream", Label: "asset.upserted"},
		{From: "kafka-stream", To: "indexer", Label: "consumer group"},
		{From: "indexer", To: "opensearch", Label: "document write"},
		{From: "opensearch", To: "servicenow-view", Label: "search + evidence"},
	}

	known := make(map[string]bool, len(nodes)+len(assets))
	for _, node := range nodes {
		known[node.ID] = true
	}
	highRisk := 0
	for _, asset := range assets {
		if asset.ID == "" {
			continue
		}
		if asset.Risk == "high" {
			highRisk++
		}
		nodes = append(nodes, domain.TopologyNode{
			ID:     asset.ID,
			Label:  asset.Name,
			Kind:   asset.Type,
			Owner:  asset.Owner,
			Risk:   asset.Risk,
			Detail: fmt.Sprintf("%s/%s in %s", asset.Environment, asset.Region, asset.Service),
			Tone:   toneFor(asset),
		})
		known[asset.ID] = true
		for _, dep := range asset.Dependencies {
			if dep == "" {
				continue
			}
			edges = append(edges, domain.TopologyEdge{From: asset.ID, To: dep, Label: "depends on"})
		}
	}

	return domain.SupportTopology{
		Story: domain.ProductStory{
			Mission:           "Keep local office support teams from guessing which endpoint, service, or dependency broke.",
			Vision:            "Turn branch network changes into a searchable ServiceNow-ready CMDB stream within minutes.",
			ServiceNowUseCase: "Support agents can search office assets, trace dependencies, replay recent changes, and attach evidence to incidents or change reviews.",
		},
		Metrics: domain.TopologyMetrics{
			Assets:              len(assets),
			HighRisk:            highRisk,
			RecentEvents:        recentEvents,
			ConfigVersion:       config.Version,
			QuerySLO:            "p95 search < 120ms local",
			IndexingLagSLO:      "event-to-index < 2s local",
			ReplayWindow:        config.ReplayWindow,
			DependencyGuard:     "DFS cycle check before Kafka publish",
			ConfigReloadCadence: "10s hot reload",
		},
		Nodes: compactKnownNodes(nodes, known),
		Edges: edges,
		Teams: []domain.TeamNeed{
			{Team: "Machine learning", Need: "clean incident and asset signals", ServedBy: "Kafka event contract plus replay"},
			{Team: "Search", Need: "low-latency inventory lookup", ServedBy: "OpenSearch index and configurable search fields"},
			{Team: "Product", Need: "ServiceNow support workflow", ServedBy: "CMDB-style topology and evidence APIs"},
			{Team: "Data", Need: "auditable change history", ServedBy: "asset.upserted stream and recent-event buffer"},
			{Team: "Frontend", Need: "stable read model", ServedBy: "/api/topology and /api/assets/search"},
		},
	}
}

func compactKnownNodes(nodes []domain.TopologyNode, known map[string]bool) []domain.TopologyNode {
	result := make([]domain.TopologyNode, 0, len(nodes))
	seen := make(map[string]bool, len(nodes))
	for _, node := range nodes {
		if node.ID == "" || seen[node.ID] || !known[node.ID] {
			continue
		}
		seen[node.ID] = true
		result = append(result, node)
	}
	return result
}

func toneFor(asset domain.Asset) string {
	switch asset.Risk {
	case "high":
		return "rose"
	case "medium":
		return "violet"
	case "low":
		return "green"
	}
	switch asset.Type {
	case "database", "endpoint":
		return "green"
	case "queue", "network":
		return "blue"
	case "worker":
		return "violet"
	default:
		return "blue"
	}
}
