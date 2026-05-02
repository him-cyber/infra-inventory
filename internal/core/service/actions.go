package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/him-cyber/infra-inventory-stream/internal/core/domain"
	"github.com/him-cyber/infra-inventory-stream/internal/core/ml"
)

func (s *Service) Analytics(ctx context.Context) (domain.AnalyticsSnapshot, error) {
	snap := s.config.Snapshot()
	result, err := s.store.Search(ctx, domain.SearchQuery{}, snap.SearchFields)
	if err != nil {
		return domain.AnalyticsSnapshot{}, err
	}
	recent := s.recent.Snapshot()
	highRisk := countRisk(result.Assets, "high")
	endpoints := countTypes(result.Assets, "endpoint", "network")

	return domain.AnalyticsSnapshot{
		GeneratedAt: time.Now().UTC(),
		Summary:     fmt.Sprintf("%d office assets indexed, %d high-risk dependency, %d recent stream events", len(result.Assets), highRisk, len(recent)),
		Signals: []domain.AnalyticsSignal{
			{Label: "Search index", Value: fmt.Sprintf("%d assets", len(result.Assets)), Status: "online"},
			{Label: "Office endpoints", Value: fmt.Sprintf("%d tracked", endpoints), Status: statusFor(endpoints > 0)},
			{Label: "High risk", Value: fmt.Sprintf("%d assets", highRisk), Status: riskStatus(highRisk)},
			{Label: "Replay buffer", Value: fmt.Sprintf("%d/%d events", len(recent), snap.ReplayWindow), Status: "ready"},
			{Label: "Config", Value: snap.Version, Status: "hot-reloaded"},
		},
		Recommendations: recommendations(highRisk, len(recent)),
		LiveFeed:        liveFeed(recent),
	}, nil
}

func (s *Service) StartCVEAnalysis(ctx context.Context) (domain.CVEAnalysisReport, error) {
	snap := s.config.Snapshot()
	result, err := s.store.Search(ctx, domain.SearchQuery{}, snap.SearchFields)
	if err != nil {
		return domain.CVEAnalysisReport{}, err
	}
	model := ml.NewCVEModel()
	findings := make([]domain.CVEFinding, 0, len(result.Assets))
	for _, asset := range result.Assets {
		finding := model.Predict(asset)
		if finding.Severity == "low" {
			continue
		}
		findings = append(findings, finding)
	}
	sort.SliceStable(findings, func(i, j int) bool {
		return findings[i].Score > findings[j].Score
	})
	now := time.Now().UTC()
	return domain.CVEAnalysisReport{
		AnalysisID:       "CVE-ANALYSIS-" + now.Format("20060102-150405"),
		GeneratedAt:      now,
		Model:            "weighted-logistic-cve-risk-v1",
		Summary:          fmt.Sprintf("%d CVE candidates scored from %d OpenSearch assets", len(findings), len(result.Assets)),
		Findings:         findings,
		KafkaTopic:       s.topic,
		SearchBackend:    "OpenSearch / Elasticsearch-compatible index",
		ServiceNowTarget: "incident + change evidence",
	}, nil
}

func (s *Service) CreateServiceNowTicket(ctx context.Context) (domain.ServiceNowTicket, error) {
	report, err := s.StartCVEAnalysis(ctx)
	if err != nil {
		return domain.ServiceNowTicket{}, err
	}
	now := time.Now().UTC()
	findings := report.Findings
	configurationItems := make([]string, 0, min(len(findings), 5))
	for _, finding := range findings {
		configurationItems = append(configurationItems, finding.AssetName)
		if len(configurationItems) == 5 {
			break
		}
	}
	return domain.ServiceNowTicket{
		Number:             "INC" + now.Format("20060102150405"),
		CreatedAt:          now,
		Table:              "incident",
		ShortDescription:   "CVE risk detected in local office support topology",
		AssignmentGroup:    "Core Infrastructure",
		Impact:             "2",
		Urgency:            "2",
		Priority:           "2",
		ConfigurationItems: configurationItems,
		WorkNotes: []string{
			report.Summary,
			"Evidence: Kafka topic " + report.KafkaTopic,
			"Evidence: " + report.SearchBackend,
			"Attach /api/security/analyze output to ServiceNow incident/change record.",
		},
	}, nil
}

func (s *Service) ConfigIntelligence(ctx context.Context) (domain.ConfigIntelligence, error) {
	snap := s.config.Snapshot()
	result, err := s.store.Search(ctx, domain.SearchQuery{}, snap.SearchFields)
	if err != nil {
		return domain.ConfigIntelligence{}, err
	}
	owners := topValues(result.Assets, func(asset domain.Asset) string { return asset.Owner }, 6)
	services := topValues(result.Assets, func(asset domain.Asset) string { return asset.Service }, 6)
	highRisk := countRisk(result.Assets, "high")
	stateful := countTypes(result.Assets, "database", "queue")
	endpoints := countTypes(result.Assets, "endpoint", "network")
	recent := len(s.recent.Snapshot())

	proposals := []domain.ConfigProposal{
		{
			Field:      "search_fields",
			Current:    strings.Join(snap.SearchFields, ", "),
			Proposed:   "name, owner, service, region, environment, risk, type",
			Reason:     "the indexed data now includes office endpoints, network devices, services, queues, and databases; support search should include ownership and service routing fields",
			Confidence: confidence(0.76, len(result.Assets), 20),
		},
		{
			Field:      "replay_window",
			Current:    fmt.Sprintf("%d events", snap.ReplayWindow),
			Proposed:   fmt.Sprintf("%d events", max(128, recent*4)),
			Reason:     "recent Kafka activity should leave enough replay history for incident evidence and change-review reconstruction",
			Confidence: confidence(0.72, recent, 30),
		},
		{
			Field:      "risk_routing",
			Current:    "manual assignment",
			Proposed:   "critical/high -> Core Infrastructure, endpoint/network -> IT Operations, frontend -> Product",
			Reason:     "CVE scoring and asset ownership can pre-route ServiceNow incidents before an agent reads the ticket",
			Confidence: confidence(0.81, highRisk+endpoints, 8),
		},
		{
			Field:      "stateful_controls",
			Current:    "generic inventory validation",
			Proposed:   "require owner, dependency, and replay evidence for database and queue assets",
			Reason:     "stateful assets create the highest incident blast radius and need stronger config evidence",
			Confidence: confidence(0.79, stateful, 5),
		},
	}

	return domain.ConfigIntelligence{
		GeneratedAt:      time.Now().UTC(),
		Model:            "config-recommendation-rules-v1",
		Summary:          fmt.Sprintf("%d assets analyzed across %d owners and %d services", len(result.Assets), len(owners), len(services)),
		ObservedAssets:   len(result.Assets),
		ObservedOwners:   owners,
		ObservedServices: services,
		Proposals:        proposals,
		ServiceNowRouting: map[string]string{
			"critical_cve": "Incident / Core Infrastructure / P1 evidence review",
			"high_cve":     "Incident / Core Infrastructure / P2 owner remediation",
			"network":      "Change / IT Operations / branch network approval",
			"endpoint":     "Incident / Helpdesk / employee device support",
			"database":     "Change / Platform / stateful dependency review",
		},
		PreviewConfig: map[string]any{
			"version":              snap.Version + ".preview",
			"search_fields":        []string{"name", "owner", "service", "region", "environment", "risk", "type"},
			"replay_window":        max(128, recent*4),
			"owner_routes":         owners,
			"stateful_asset_types": []string{"database", "queue"},
		},
	}, nil
}

func countRisk(assets []domain.Asset, risk string) int {
	count := 0
	for _, asset := range assets {
		if asset.Risk == risk {
			count++
		}
	}
	return count
}

func countTypes(assets []domain.Asset, types ...string) int {
	allowed := make(map[string]bool, len(types))
	for _, kind := range types {
		allowed[kind] = true
	}
	count := 0
	for _, asset := range assets {
		if allowed[asset.Type] {
			count++
		}
	}
	return count
}

func topValues(assets []domain.Asset, pick func(domain.Asset) string, limit int) []string {
	counts := make(map[string]int)
	for _, asset := range assets {
		value := pick(asset)
		if value != "" {
			counts[value]++
		}
	}
	values := make([]string, 0, len(counts))
	for value := range counts {
		values = append(values, value)
	}
	sort.Slice(values, func(i, j int) bool {
		if counts[values[i]] == counts[values[j]] {
			return values[i] < values[j]
		}
		return counts[values[i]] > counts[values[j]]
	})
	if len(values) > limit {
		return values[:limit]
	}
	return values
}

func confidence(base float64, observed, target int) float64 {
	if target <= 0 || observed >= target {
		return minFloat(0.95, base+0.12)
	}
	return minFloat(0.95, base+(float64(observed)/float64(target))*0.12)
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func statusFor(ok bool) string {
	if ok {
		return "ready"
	}
	return "missing"
}

func riskStatus(highRisk int) string {
	if highRisk > 0 {
		return "needs-review"
	}
	return "clear"
}

func recommendations(highRisk, recentEvents int) []string {
	items := []string{
		"Use OpenSearch-backed inventory to identify affected owners before opening a change.",
		"Use Kafka replay to recreate the last support window for debugging.",
	}
	if highRisk > 0 {
		items = append([]string{"Launch a ServiceNow support workflow for high-risk dependencies."}, items...)
	}
	if recentEvents == 0 {
		items = append(items, "Seed or ingest office endpoint events before demoing the live feed.")
	}
	return items
}

func liveFeed(events []domain.Event) []domain.LiveFeedItem {
	feed := make([]domain.LiveFeedItem, 0, min(len(events), 6))
	for i := len(events) - 1; i >= 0 && len(feed) < 6; i-- {
		event := events[i]
		feed = append(feed, domain.LiveFeedItem{
			Title:     event.Payload.Name,
			Detail:    event.EventType + " via " + event.ConfigVersion,
			Timestamp: event.Timestamp,
		})
	}
	return feed
}
