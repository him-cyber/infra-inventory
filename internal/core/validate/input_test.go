package validate

import (
	"strings"
	"testing"

	"github.com/him-cyber/infra-inventory-stream/internal/core/domain"
)

func TestAssetRejectsUnsafeFields(t *testing.T) {
	_, err := Asset(domain.Asset{
		ID:          "bad/id",
		Type:        "service",
		Name:        "Catalog",
		Owner:       "platform",
		Environment: "prod",
		Region:      "us-east",
		Service:     "catalog",
		Risk:        "low",
	})
	if err == nil {
		t.Fatal("expected invalid asset id to fail")
	}
}

func TestAssetNormalizesAndDeduplicatesDependencies(t *testing.T) {
	asset, err := Asset(domain.Asset{
		ID:           "api",
		Type:         "service",
		Name:         "Catalog API",
		Owner:        "platform",
		Environment:  "prod",
		Region:       "us-east",
		Service:      "catalog",
		Risk:         "HIGH",
		Dependencies: []string{"queue", "queue"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if asset.Risk != "high" || len(asset.Dependencies) != 1 {
		t.Fatalf("unexpected normalized asset: %+v", asset)
	}
}

func TestSearchQueryBoundsInput(t *testing.T) {
	_, err := SearchQuery(domain.SearchQuery{Text: strings.Repeat("x", MaxSearchLength+1)})
	if err == nil {
		t.Fatal("expected long search text to fail")
	}
	_, err = SearchQuery(domain.SearchQuery{Owner: "bad owner"})
	if err == nil {
		t.Fatal("expected unsafe owner filter to fail")
	}
}

func TestOrgImportCapsBatchSize(t *testing.T) {
	req := domain.OrgImportRequest{Provider: "azure", AccountID: "acct", Region: "us-east"}
	for i := 0; i < MaxAssetsPerBatch+1; i++ {
		req.Assets = append(req.Assets, domain.Asset{ID: "asset-a", Type: "service", Name: "Asset", Owner: "platform", Environment: "prod", Region: "us-east", Service: "svc", Risk: "low"})
	}
	_, err := OrgImport(req)
	if err == nil {
		t.Fatal("expected oversized import to fail")
	}
}
