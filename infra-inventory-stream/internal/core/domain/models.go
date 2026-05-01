package domain

import "time"

type Asset struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"`
	Name         string    `json:"name"`
	Owner        string    `json:"owner"`
	Environment  string    `json:"environment"`
	Region       string    `json:"region"`
	Service      string    `json:"service"`
	Version      int       `json:"version"`
	Dependencies []string  `json:"dependencies"`
	Risk         string    `json:"risk"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Event struct {
	EventID       string    `json:"event_id"`
	AssetID       string    `json:"asset_id"`
	EventType     string    `json:"event_type"`
	Version       int       `json:"version"`
	Timestamp     time.Time `json:"timestamp"`
	Payload       Asset     `json:"payload"`
	ConfigVersion string    `json:"config_version"`
}

type SearchQuery struct {
	Text        string
	Type        string
	Environment string
	Owner       string
}

type SearchResult struct {
	TookMs int     `json:"took_ms"`
	Total  int     `json:"total"`
	Assets []Asset `json:"assets"`
}

type ConfigSnapshot struct {
	Version          string            `json:"version"`
	AllowedTypes     []string          `json:"allowed_asset_types"`
	SearchFields     []string          `json:"search_fields"`
	TenantRateLimits map[string]int    `json:"tenant_rate_limits"`
	IndexRouting     map[string]string `json:"index_routing"`
	ReplayWindow     int               `json:"replay_window"`
	UpdatedAt        time.Time         `json:"updated_at"`
}
