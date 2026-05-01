package domain

import "time"

type OrgImportRequest struct {
	Provider       string  `json:"provider"`
	AccountID      string  `json:"account_id"`
	TenantID       string  `json:"tenant_id"`
	SubscriptionID string  `json:"subscription_id"`
	Region         string  `json:"region"`
	Assets         []Asset `json:"assets"`
}

type OrgImportResult struct {
	ImportID         string    `json:"import_id"`
	Provider         string    `json:"provider"`
	AccountID        string    `json:"account_id"`
	TenantID         string    `json:"tenant_id"`
	SubscriptionID   string    `json:"subscription_id"`
	ImportedAssets   int       `json:"imported_assets"`
	PublishedEvents  int       `json:"published_events"`
	EncryptedAtRest  string    `json:"encrypted_at_rest"`
	EncryptedTransit string    `json:"encrypted_in_transit"`
	CreatedAt        time.Time `json:"created_at"`
}
