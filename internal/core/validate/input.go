package validate

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/him-cyber/infra-inventory-stream/internal/core/domain"
)

const (
	MaxJSONBodyBytes  = 1 << 20
	MaxStringLength   = 120
	MaxSearchLength   = 160
	MaxAssetsPerBatch = 100
	MaxDependencies   = 32
)

var slugPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._:-]{0,119}$`)

var allowedRisks = map[string]bool{
	"low":    true,
	"medium": true,
	"high":   true,
}

func LimitJSONBody(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, MaxJSONBodyBytes)
}

func Tenant(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "demo", nil
	}
	if !slugPattern.MatchString(value) {
		return "", errors.New("tenant must be 1-120 characters and contain only letters, numbers, dot, underscore, colon, or dash")
	}
	return value, nil
}

func AssetID(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", errors.New("asset id is required")
	}
	if !slugPattern.MatchString(value) {
		return "", errors.New("asset id must be 1-120 characters and contain only letters, numbers, dot, underscore, colon, or dash")
	}
	return value, nil
}

func SearchQuery(q domain.SearchQuery) (domain.SearchQuery, error) {
	var err error
	q.Text, err = boundedText(q.Text, MaxSearchLength, "q")
	if err != nil {
		return domain.SearchQuery{}, err
	}
	q.Type, err = optionalSlug(q.Type, "type")
	if err != nil {
		return domain.SearchQuery{}, err
	}
	q.Environment, err = optionalSlug(q.Environment, "env")
	if err != nil {
		return domain.SearchQuery{}, err
	}
	q.Owner, err = optionalSlug(q.Owner, "owner")
	if err != nil {
		return domain.SearchQuery{}, err
	}
	return q, nil
}

func Asset(asset domain.Asset) (domain.Asset, error) {
	var err error
	asset.ID = strings.TrimSpace(asset.ID)
	if asset.ID != "" {
		asset.ID, err = AssetID(asset.ID)
		if err != nil {
			return domain.Asset{}, err
		}
	}
	asset.Type, err = requiredSlug(asset.Type, "type")
	if err != nil {
		return domain.Asset{}, err
	}
	asset.Name, err = boundedRequiredText(asset.Name, MaxStringLength, "name")
	if err != nil {
		return domain.Asset{}, err
	}
	asset.Owner, err = requiredSlug(asset.Owner, "owner")
	if err != nil {
		return domain.Asset{}, err
	}
	asset.Environment, err = requiredSlug(asset.Environment, "environment")
	if err != nil {
		return domain.Asset{}, err
	}
	asset.Region, err = requiredSlug(asset.Region, "region")
	if err != nil {
		return domain.Asset{}, err
	}
	asset.Service, err = requiredSlug(asset.Service, "service")
	if err != nil {
		return domain.Asset{}, err
	}
	asset.Risk = strings.ToLower(strings.TrimSpace(asset.Risk))
	if !allowedRisks[asset.Risk] {
		return domain.Asset{}, errors.New("risk must be low, medium, or high")
	}
	if asset.Version < 0 || asset.Version > 1_000_000_000 {
		return domain.Asset{}, errors.New("version must be between 0 and 1000000000")
	}
	if len(asset.Dependencies) > MaxDependencies {
		return domain.Asset{}, fmt.Errorf("dependencies cannot exceed %d entries", MaxDependencies)
	}
	deps := make([]string, 0, len(asset.Dependencies))
	seen := make(map[string]bool, len(asset.Dependencies))
	for _, dep := range asset.Dependencies {
		dep, err = AssetID(dep)
		if err != nil {
			return domain.Asset{}, fmt.Errorf("invalid dependency: %w", err)
		}
		if !seen[dep] {
			seen[dep] = true
			deps = append(deps, dep)
		}
	}
	asset.Dependencies = deps
	return asset, nil
}

func OrgImport(req domain.OrgImportRequest) (domain.OrgImportRequest, error) {
	var err error
	req.Provider, err = requiredSlug(strings.ToLower(req.Provider), "provider")
	if err != nil {
		return domain.OrgImportRequest{}, err
	}
	req.AccountID, err = optionalSlug(req.AccountID, "account_id")
	if err != nil {
		return domain.OrgImportRequest{}, err
	}
	req.TenantID, err = optionalSlug(req.TenantID, "tenant_id")
	if err != nil {
		return domain.OrgImportRequest{}, err
	}
	req.SubscriptionID, err = optionalSlug(req.SubscriptionID, "subscription_id")
	if err != nil {
		return domain.OrgImportRequest{}, err
	}
	req.Region, err = optionalSlug(req.Region, "region")
	if err != nil {
		return domain.OrgImportRequest{}, err
	}
	if len(req.Assets) == 0 {
		return domain.OrgImportRequest{}, errors.New("at least one asset is required")
	}
	if len(req.Assets) > MaxAssetsPerBatch {
		return domain.OrgImportRequest{}, fmt.Errorf("assets cannot exceed %d entries", MaxAssetsPerBatch)
	}
	for i, asset := range req.Assets {
		if asset.Region == "" && req.Region != "" {
			asset.Region = req.Region
		}
		if asset.Environment == "" {
			asset.Environment = "prod"
		}
		if asset.Service == "" {
			asset.Service = req.Provider + "-discovery"
		}
		req.Assets[i], err = Asset(asset)
		if err != nil {
			return domain.OrgImportRequest{}, fmt.Errorf("asset[%d]: %w", i, err)
		}
	}
	return req, nil
}

func boundedRequiredText(value string, limit int, field string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("%s is required", field)
	}
	return boundedText(value, limit, field)
}

func boundedText(value string, limit int, field string) (string, error) {
	value = strings.TrimSpace(value)
	if !utf8.ValidString(value) {
		return "", fmt.Errorf("%s must be valid UTF-8", field)
	}
	if utf8.RuneCountInString(value) > limit {
		return "", fmt.Errorf("%s cannot exceed %d characters", field, limit)
	}
	return value, nil
}

func requiredSlug(value, field string) (string, error) {
	value, err := optionalSlug(value, field)
	if err != nil {
		return "", err
	}
	if value == "" {
		return "", fmt.Errorf("%s is required", field)
	}
	return value, nil
}

func optionalSlug(value, field string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	if !slugPattern.MatchString(value) {
		return "", fmt.Errorf("%s must be 1-120 characters and contain only letters, numbers, dot, underscore, colon, or dash", field)
	}
	return value, nil
}
