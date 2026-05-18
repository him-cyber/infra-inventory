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

type assetStore struct {
	assets []domain.Asset
}

func (s assetStore) GetAsset(context.Context, string) (domain.Asset, bool, error) {
	return domain.Asset{}, false, nil
}

func (s assetStore) Search(context.Context, domain.SearchQuery, []string) (domain.SearchResult, error) {
	return domain.SearchResult{Assets: s.assets, Total: len(s.assets)}, nil
}

func TestAnalyticsCVEAnalysisAndTicketUseInventorySignals(t *testing.T) {
	cfg := config.NewManager(staticSource{})
	if err := cfg.Load(context.Background()); err != nil {
		t.Fatal(err)
	}
	recent := ring.New[domain.Event](8)
	recent.Push(domain.Event{EventType: "asset.upserted", ConfigVersion: "v1", Timestamp: time.Now(), Payload: domain.Asset{Name: "Branch Wi-Fi"}})
	app := NewService(cfg, &memoryPublisher{}, assetStore{assets: []domain.Asset{
		{Name: "Branch Wi-Fi", Type: "network", Risk: "medium"},
		{Name: "Search Indexer", Type: "worker", Risk: "high"},
	}}, recent, ratelimit.NewLimiter(), "inventory.events")

	analytics, err := app.Analytics(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(analytics.Signals) == 0 || len(analytics.LiveFeed) != 1 {
		t.Fatalf("expected signals and feed, got %+v", analytics)
	}
	report, err := app.StartCVEAnalysis(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if report.AnalysisID == "" || report.SearchBackend == "" || len(report.Findings) == 0 {
		t.Fatalf("expected CVE analysis report, got %+v", report)
	}
	if report.IncidentBrief == nil || report.IncidentBrief.Mode == "" {
		t.Fatalf("expected incident brief, got %+v", report.IncidentBrief)
	}
	ticket, err := app.CreateServiceNowTicket(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if ticket.Number == "" || ticket.Table != "incident" || len(ticket.ConfigurationItems) == 0 {
		t.Fatalf("expected ServiceNow ticket, got %+v", ticket)
	}
	if len(app.Tickets()) != 1 {
		t.Fatalf("expected submitted ticket to be stored, got %+v", app.Tickets())
	}
	intel, err := app.ConfigIntelligence(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(intel.Proposals) == 0 || intel.PreviewConfig["version"] == "" || intel.ServiceNowRouting["high_cve"] == "" {
		t.Fatalf("expected config intelligence, got %+v", intel)
	}
}
