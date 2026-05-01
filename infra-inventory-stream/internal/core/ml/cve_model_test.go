package ml

import (
	"testing"

	"github.com/him-cyber/infra-inventory-stream/internal/core/domain"
)

func TestCVEModelScoresHighRiskNetworkAboveLowRiskWorker(t *testing.T) {
	model := NewCVEModel()
	network := model.Predict(domain.Asset{ID: "edge", Type: "network", Name: "Edge Router", Risk: "high", Version: 8, Dependencies: []string{"wifi", "api"}})
	worker := model.Predict(domain.Asset{ID: "worker", Type: "worker", Name: "Worker", Risk: "low", Version: 40})
	if network.Confidence <= worker.Confidence {
		t.Fatalf("expected network confidence > worker confidence, got %.2f <= %.2f", network.Confidence, worker.Confidence)
	}
	if network.CVEID == "" || len(network.Signals) == 0 {
		t.Fatalf("expected cve id and feature signals, got %+v", network)
	}
}
