package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	osclient "github.com/opensearch-project/opensearch-go/v4"

	"github.com/him-cyber/infra-inventory-stream/internal/core/domain"
)

type Store struct {
	client *osclient.Client
	index  string
}

func NewStore() (*Store, error) {
	address := valueOrDefault(os.Getenv("OPENSEARCH_URL"), "http://opensearch:9200")
	client, err := osclient.NewClient(osclient.Config{
		Addresses: []string{address},
		Username:  os.Getenv("OPENSEARCH_USERNAME"),
		Password:  os.Getenv("OPENSEARCH_PASSWORD"),
	})
	if err != nil {
		return nil, err
	}
	return &Store{client: client, index: valueOrDefault(os.Getenv("OPENSEARCH_INDEX"), "inventory-assets")}, nil
}

func (s *Store) EnsureIndex(ctx context.Context) error {
	body := strings.NewReader(`{"settings":{"index":{"number_of_shards":1,"number_of_replicas":0}},"mappings":{"properties":{"name":{"type":"text"},"service":{"type":"keyword"},"owner":{"type":"keyword"},"environment":{"type":"keyword"},"type":{"type":"keyword"},"risk":{"type":"keyword"},"updated_at":{"type":"date"}}}}`)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, "/"+s.index, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Perform(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusBadRequest {
		return nil
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("create index failed: %s", resp.Status)
	}
	return nil
}

func (s *Store) UpsertAsset(ctx context.Context, asset domain.Asset) error {
	if asset.UpdatedAt.IsZero() {
		asset.UpdatedAt = time.Now().UTC()
	}
	data, err := json.Marshal(asset)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, "/"+s.index+"/_doc/"+asset.ID+"?refresh=true", bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Perform(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("index asset failed: %s", resp.Status)
	}
	return nil
}

func (s *Store) GetAsset(ctx context.Context, id string) (domain.Asset, bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/"+s.index+"/_doc/"+id, nil)
	if err != nil {
		return domain.Asset{}, false, err
	}
	resp, err := s.client.Perform(req)
	if err != nil {
		return domain.Asset{}, false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return domain.Asset{}, false, nil
	}
	var decoded struct {
		Source domain.Asset `json:"_source"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return domain.Asset{}, false, err
	}
	return decoded.Source, true, nil
}

func (s *Store) Search(ctx context.Context, q domain.SearchQuery, fields []string) (domain.SearchResult, error) {
	filter := make([]map[string]map[string]string, 0, 3)
	if q.Type != "" {
		filter = append(filter, term("type", q.Type))
	}
	if q.Environment != "" {
		filter = append(filter, term("environment", q.Environment))
	}
	if q.Owner != "" {
		filter = append(filter, term("owner", q.Owner))
	}
	query := map[string]any{"match_all": map[string]any{}}
	if q.Text != "" {
		query = map[string]any{"multi_match": map[string]any{"query": q.Text, "fields": fields}}
	}
	body := map[string]any{"size": 50, "query": map[string]any{"bool": map[string]any{"must": []any{query}, "filter": filter}}}
	data, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/"+s.index+"/_search", bytes.NewReader(data))
	if err != nil {
		return domain.SearchResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Perform(req)
	if err != nil {
		return domain.SearchResult{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return domain.SearchResult{}, fmt.Errorf("search failed: %s", resp.Status)
	}
	var decoded struct {
		Took int `json:"took"`
		Hits struct {
			Total struct {
				Value int `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source domain.Asset `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return domain.SearchResult{}, err
	}
	result := domain.SearchResult{TookMs: decoded.Took, Total: decoded.Hits.Total.Value}
	for _, hit := range decoded.Hits.Hits {
		result.Assets = append(result.Assets, hit.Source)
	}
	return result, nil
}

func term(field, value string) map[string]map[string]string {
	return map[string]map[string]string{"term": map[string]string{field: value}}
}

func valueOrDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
