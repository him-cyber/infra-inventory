package config

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/him-cyber/infra-inventory-stream/internal/core/domain"
)

type Source interface {
	Load(context.Context) (domain.ConfigSnapshot, error)
}

type FileSource struct {
	Path string
}

func (s FileSource) Load(ctx context.Context) (domain.ConfigSnapshot, error) {
	data, err := os.ReadFile(s.Path)
	if err != nil {
		return domain.ConfigSnapshot{}, err
	}
	return parse(data)
}

type BlobSource struct {
	AccountURL string
	Container  string
	Blob       string
}

func (s BlobSource) Load(ctx context.Context) (domain.ConfigSnapshot, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return domain.ConfigSnapshot{}, err
	}
	client, err := azblob.NewClient(s.AccountURL, cred, nil)
	if err != nil {
		return domain.ConfigSnapshot{}, err
	}
	resp, err := client.DownloadStream(ctx, s.Container, s.Blob, nil)
	if err != nil {
		return domain.ConfigSnapshot{}, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return domain.ConfigSnapshot{}, err
	}
	return parse(data)
}

func SourceFromEnv() Source {
	if strings.EqualFold(os.Getenv("CONFIG_SOURCE"), "azure") {
		return BlobSource{
			AccountURL: os.Getenv("AZURE_STORAGE_ACCOUNT_URL"),
			Container:  valueOrDefault(os.Getenv("AZURE_CONFIG_CONTAINER"), "config"),
			Blob:       valueOrDefault(os.Getenv("AZURE_CONFIG_BLOB"), "inventory-config.json"),
		}
	}
	return FileSource{Path: valueOrDefault(os.Getenv("CONFIG_PATH"), "config/inventory-config.json")}
}

func parse(data []byte) (domain.ConfigSnapshot, error) {
	var snap domain.ConfigSnapshot
	if err := json.NewDecoder(bytes.NewReader(data)).Decode(&snap); err != nil {
		return snap, err
	}
	if snap.UpdatedAt.IsZero() {
		snap.UpdatedAt = time.Now().UTC()
	}
	return snap, Validate(snap)
}

func Validate(s domain.ConfigSnapshot) error {
	if s.Version == "" {
		return errors.New("config version is required")
	}
	if len(s.AllowedTypes) == 0 {
		return errors.New("at least one allowed asset type is required")
	}
	if len(s.SearchFields) == 0 {
		return errors.New("at least one search field is required")
	}
	if s.ReplayWindow <= 0 {
		return errors.New("replay_window must be positive")
	}
	for tenant, limit := range s.TenantRateLimits {
		if strings.TrimSpace(tenant) == "" {
			return errors.New("tenant rate limit key cannot be empty")
		}
		if limit <= 0 {
			return errors.New("tenant rate limits must be positive")
		}
	}
	return nil
}

func valueOrDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
