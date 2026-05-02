package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/him-cyber/infra-inventory-stream/internal/core/domain"
	"github.com/him-cyber/infra-inventory-stream/internal/core/validate"
)

var allowedProviders = map[string]bool{
	"azure": true,
	"aws":   true,
	"gcp":   true,
}

func (s *Service) ImportOrg(ctx context.Context, tenant string, req domain.OrgImportRequest) (domain.OrgImportResult, error) {
	var err error
	tenant, err = validate.Tenant(tenant)
	if err != nil {
		return domain.OrgImportResult{}, err
	}
	req, err = validate.OrgImport(req)
	if err != nil {
		return domain.OrgImportResult{}, err
	}
	provider := strings.ToLower(strings.TrimSpace(req.Provider))
	if !allowedProviders[provider] {
		return domain.OrgImportResult{}, errors.New("provider must be azure, aws, or gcp")
	}
	accountID := strings.TrimSpace(req.AccountID)
	if accountID == "" && req.SubscriptionID != "" {
		accountID = req.SubscriptionID
	}
	if accountID == "" {
		return domain.OrgImportResult{}, errors.New("account_id or subscription_id is required")
	}
	if len(req.Assets) == 0 {
		return domain.OrgImportResult{}, errors.New("at least one discovered asset is required")
	}
	published := 0
	for _, asset := range req.Assets {
		if asset.Region == "" {
			asset.Region = req.Region
		}
		if asset.Environment == "" {
			asset.Environment = "prod"
		}
		if asset.Service == "" {
			asset.Service = provider + "-discovery"
		}
		if _, err := s.UpsertAsset(ctx, tenant, asset); err != nil {
			return domain.OrgImportResult{}, err
		}
		published++
	}
	result := domain.OrgImportResult{
		ImportID:         uuid.NewString(),
		Provider:         provider,
		AccountID:        accountID,
		TenantID:         req.TenantID,
		SubscriptionID:   req.SubscriptionID,
		ImportedAssets:   len(req.Assets),
		PublishedEvents:  published,
		EncryptedAtRest:  "OpenSearch volume encryption in Azure; browser session data is AES-GCM encrypted in an HttpOnly cookie",
		EncryptedTransit: "TLS at ingress plus Kafka/Event Hubs SASL_SSL in Azure mode",
		CreatedAt:        time.Now().UTC(),
	}
	s.recordOrgImport(result)
	return result, nil
}

func (s *Service) OrgImports() []domain.OrgImportResult {
	s.orgMu.RLock()
	defer s.orgMu.RUnlock()
	results := make([]domain.OrgImportResult, len(s.orgImports))
	copy(results, s.orgImports)
	return results
}

func (s *Service) recordOrgImport(result domain.OrgImportResult) {
	s.orgMu.Lock()
	defer s.orgMu.Unlock()
	s.orgImports = append([]domain.OrgImportResult{result}, s.orgImports...)
	if len(s.orgImports) > 10 {
		s.orgImports = s.orgImports[:10]
	}
}
