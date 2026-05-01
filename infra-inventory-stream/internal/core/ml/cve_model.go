package ml

import (
	"math"
	"strings"

	"github.com/him-cyber/infra-inventory-stream/internal/core/domain"
)

type CVEModel struct {
	weights map[string]float64
}

func NewCVEModel() CVEModel {
	return CVEModel{weights: map[string]float64{
		"bias":           -1.2,
		"high_risk":      1.4,
		"medium_risk":    0.7,
		"internet_edge":  1.1,
		"stateful":       0.9,
		"many_deps":      0.8,
		"old_version":    0.6,
		"security_owner": -0.3,
	}}
}

func (m CVEModel) Predict(asset domain.Asset) domain.CVEFinding {
	features := m.features(asset)
	score := m.weights["bias"]
	signals := make([]string, 0, len(features))
	for feature, value := range features {
		if value == 0 {
			continue
		}
		score += m.weights[feature] * value
		signals = append(signals, feature)
	}
	confidence := sigmoid(score)
	severity := severityFor(confidence)
	return domain.CVEFinding{
		CVEID:       cveFor(asset),
		AssetID:     asset.ID,
		AssetName:   asset.Name,
		Severity:    severity,
		Score:       round(confidence * 10),
		Confidence:  round(confidence),
		Signals:     signals,
		Remediation: remediationFor(asset, severity),
	}
}

func (m CVEModel) features(asset domain.Asset) map[string]float64 {
	features := map[string]float64{}
	switch asset.Risk {
	case "high":
		features["high_risk"] = 1
	case "medium":
		features["medium_risk"] = 1
	}
	switch asset.Type {
	case "network", "frontend", "service":
		features["internet_edge"] = 1
	case "database", "cache", "queue":
		features["stateful"] = 1
	}
	if len(asset.Dependencies) >= 2 {
		features["many_deps"] = 1
	}
	if asset.Version > 0 && asset.Version < 20 {
		features["old_version"] = 1
	}
	if strings.Contains(asset.Owner, "security") {
		features["security_owner"] = 1
	}
	return features
}

func sigmoid(value float64) float64 {
	return 1 / (1 + math.Exp(-value))
}

func round(value float64) float64 {
	return math.Round(value*100) / 100
}

func severityFor(confidence float64) string {
	switch {
	case confidence >= 0.78:
		return "critical"
	case confidence >= 0.62:
		return "high"
	case confidence >= 0.45:
		return "medium"
	default:
		return "low"
	}
}

func cveFor(asset domain.Asset) string {
	switch asset.Type {
	case "network":
		return "CVE-2026-39883"
	case "endpoint":
		return "CVE-2025-68121"
	case "database":
		return "CVE-2026-32283"
	case "service", "frontend":
		return "CVE-2025-61729"
	default:
		return "CVE-2026-24051"
	}
}

func remediationFor(asset domain.Asset, severity string) string {
	if severity == "critical" || severity == "high" {
		return "open ServiceNow incident, patch owner-owned CI, and replay recent Kafka events for evidence"
	}
	return "track in ServiceNow change request and verify OpenSearch evidence after patch"
}
