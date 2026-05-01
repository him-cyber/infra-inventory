package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/him-cyber/infra-inventory-stream/internal/core/domain"
	"go.yaml.in/yaml/v2"
)

type inventoryFile struct {
	Assets []domain.Asset `json:"assets" yaml:"assets"`
}

func main() {
	filePath := flag.String("file", "", "inventory YAML or JSON file")
	apiURL := flag.String("api", "http://127.0.0.1:8080", "Infra Inventory Stream API URL")
	tenant := flag.String("tenant", "demo", "tenant header value")
	dryRun := flag.Bool("dry-run", false, "parse and validate the file without posting assets")
	flag.Parse()

	if *filePath == "" {
		exitf("missing -file")
	}

	assets, err := readAssets(*filePath)
	if err != nil {
		exitf("read inventory: %v", err)
	}
	if len(assets) == 0 {
		exitf("inventory file has no assets")
	}
	if *dryRun {
		fmt.Printf("validated %d assets from %s\n", len(assets), *filePath)
		return
	}

	client := &http.Client{}
	for _, asset := range assets {
		if err := postAsset(client, strings.TrimRight(*apiURL, "/"), *tenant, asset); err != nil {
			exitf("post asset %q: %v", asset.ID, err)
		}
	}

	fmt.Printf("imported %d assets into %s\n", len(assets), *apiURL)
}

func readAssets(path string) ([]domain.Asset, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var envelope inventoryFile
	switch strings.ToLower(filepath.Ext(path)) {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &envelope); err != nil {
			return nil, err
		}
	default:
		if err := json.Unmarshal(data, &envelope); err != nil {
			var assets []domain.Asset
			if arrayErr := json.Unmarshal(data, &assets); arrayErr != nil {
				return nil, err
			}
			envelope.Assets = assets
		}
	}
	return envelope.Assets, nil
}

func postAsset(client *http.Client, apiURL string, tenant string, asset domain.Asset) error {
	body, err := json.Marshal(asset)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, apiURL+"/api/assets", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant", tenant)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		message, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("status %s: %s", resp.Status, strings.TrimSpace(string(message)))
	}
	return nil
}

func exitf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
