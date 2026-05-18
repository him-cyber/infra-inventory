package genai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/him-cyber/infra-inventory-stream/internal/core/domain"
)

const defaultOpenAIBaseURL = "https://api.openai.com/v1"

type OpenAIClient struct {
	apiKey string
	model  string
	url    string
	http   *http.Client
}

func NewOpenAIFromEnv() *OpenAIClient {
	key := strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	if key == "" {
		return nil
	}
	model := strings.TrimSpace(os.Getenv("OPENAI_MODEL"))
	if model == "" {
		model = "gpt-5"
	}
	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("OPENAI_BASE_URL")), "/")
	if baseURL == "" {
		baseURL = defaultOpenAIBaseURL
	}
	return &OpenAIClient{
		apiKey: key,
		model:  model,
		url:    baseURL + "/responses",
		http:   &http.Client{Timeout: 8 * time.Second},
	}
}

func (c *OpenAIClient) BriefIncident(ctx context.Context, report domain.CVEAnalysisReport, assets []domain.Asset, events []domain.Event) (domain.AIIncidentBrief, error) {
	if c == nil || c.apiKey == "" {
		return domain.AIIncidentBrief{}, errors.New("openai client is not configured")
	}
	payload := map[string]any{
		"model": c.model,
		"instructions": strings.Join([]string{
			"You are an infrastructure incident intelligence service.",
			"Return only compact JSON with keys: executive_summary, probable_cause, blast_radius, recommended_actions, servicenow_work_notes.",
			"Base every claim on the supplied inventory, CVE findings, Kafka evidence, and ServiceNow target.",
			"Do not include secrets, credentials, markdown, or unsupported product claims.",
		}, " "),
		"input":             incidentPrompt(report, assets, events),
		"max_output_tokens": 600,
		"store":             false,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return domain.AIIncidentBrief{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(data))
	if err != nil {
		return domain.AIIncidentBrief{}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return domain.AIIncidentBrief{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return domain.AIIncidentBrief{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return domain.AIIncidentBrief{}, fmt.Errorf("openai response failed: status %d", resp.StatusCode)
	}
	text, err := outputText(body)
	if err != nil {
		return domain.AIIncidentBrief{}, err
	}
	var brief domain.AIIncidentBrief
	if err := json.Unmarshal([]byte(text), &brief); err != nil {
		return domain.AIIncidentBrief{}, err
	}
	brief.Provider = "openai"
	brief.Model = c.model
	brief.Mode = "responses-api"
	if len(brief.RecommendedActions) > 5 {
		brief.RecommendedActions = brief.RecommendedActions[:5]
	}
	if len(brief.ServiceNowWorkNotes) > 5 {
		brief.ServiceNowWorkNotes = brief.ServiceNowWorkNotes[:5]
	}
	return brief, nil
}

func incidentPrompt(report domain.CVEAnalysisReport, assets []domain.Asset, events []domain.Event) string {
	topFindings := report.Findings
	if len(topFindings) > 8 {
		topFindings = topFindings[:8]
	}
	assetSample := assets
	if len(assetSample) > 16 {
		assetSample = assetSample[:16]
	}
	recentEvents := events
	if len(recentEvents) > 12 {
		recentEvents = recentEvents[len(recentEvents)-12:]
	}
	payload := map[string]any{
		"analysis_id":        report.AnalysisID,
		"summary":            report.Summary,
		"model":              report.Model,
		"kafka_topic":        report.KafkaTopic,
		"search_backend":     report.SearchBackend,
		"servicenow_target":  report.ServiceNowTarget,
		"top_findings":       topFindings,
		"asset_sample":       assetSample,
		"recent_event_count": len(events),
		"recent_events":      recentEvents,
	}
	data, _ := json.Marshal(payload)
	return string(data)
}

func outputText(body []byte) (string, error) {
	var response struct {
		OutputText string `json:"output_text"`
		Output     []struct {
			Type    string `json:"type"`
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"output"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}
	if strings.TrimSpace(response.OutputText) != "" {
		return strings.TrimSpace(response.OutputText), nil
	}
	for _, item := range response.Output {
		if item.Type != "message" {
			continue
		}
		for _, content := range item.Content {
			if strings.TrimSpace(content.Text) != "" {
				return strings.TrimSpace(content.Text), nil
			}
		}
	}
	return "", errors.New("openai response did not include output text")
}
