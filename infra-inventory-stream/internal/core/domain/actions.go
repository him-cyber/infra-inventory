package domain

import "time"

type AnalyticsSignal struct {
	Label  string `json:"label"`
	Value  string `json:"value"`
	Status string `json:"status"`
}

type LiveFeedItem struct {
	Title     string    `json:"title"`
	Detail    string    `json:"detail"`
	Timestamp time.Time `json:"timestamp"`
}

type AnalyticsSnapshot struct {
	GeneratedAt     time.Time         `json:"generated_at"`
	Summary         string            `json:"summary"`
	Signals         []AnalyticsSignal `json:"signals"`
	Recommendations []string          `json:"recommendations"`
	LiveFeed        []LiveFeedItem    `json:"live_feed"`
}

type ConfigProposal struct {
	Field      string  `json:"field"`
	Current    string  `json:"current"`
	Proposed   string  `json:"proposed"`
	Reason     string  `json:"reason"`
	Confidence float64 `json:"confidence"`
}

type ConfigIntelligence struct {
	GeneratedAt       time.Time         `json:"generated_at"`
	Model             string            `json:"model"`
	Summary           string            `json:"summary"`
	ObservedAssets    int               `json:"observed_assets"`
	ObservedOwners    []string          `json:"observed_owners"`
	ObservedServices  []string          `json:"observed_services"`
	Proposals         []ConfigProposal  `json:"proposals"`
	ServiceNowRouting map[string]string `json:"servicenow_routing"`
	PreviewConfig     map[string]any    `json:"preview_config"`
}
