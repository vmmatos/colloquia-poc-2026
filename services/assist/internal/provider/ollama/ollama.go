package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"assist/internal/config"
	"assist/internal/provider"
)

type Provider struct {
	baseURL    string
	model      string
	timeout    time.Duration
	httpClient *http.Client
}

func New(cfg *config.Config) *Provider {
	timeout := time.Duration(cfg.OllamaTimeoutSeconds) * time.Second
	return &Provider{
		baseURL: cfg.OllamaBaseURL,
		model:   cfg.OllamaModel,
		timeout: timeout,
		httpClient: &http.Client{
			Timeout: timeout + 2*time.Second, // slightly above context timeout
		},
	}
}

// ── Ollama API types ──────────────────────────────────────────────────────────

type generateRequest struct {
	Model  string `json:"model"`
	System string `json:"system"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type generateResponse struct {
	Response string `json:"response"`
}

// ── LLMProvider implementation ────────────────────────────────────────────────

func (p *Provider) Complete(ctx context.Context, req provider.CompletionRequest) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	body, err := json.Marshal(generateRequest{
		Model:  p.model,
		System: req.SystemPrompt,
		Prompt: req.UserPrompt,
		Stream: false,
	})
	if err != nil {
		return nil, fmt.Errorf("ollama: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("ollama: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ollama: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("ollama: unexpected status %d: %s", resp.StatusCode, raw)
	}

	var genResp generateResponse
	if err := json.NewDecoder(resp.Body).Decode(&genResp); err != nil {
		return nil, fmt.Errorf("ollama: decode response: %w", err)
	}

	suggestions, err := parseSuggestions(genResp.Response)
	if err != nil {
		return nil, fmt.Errorf("ollama: parse suggestions: %w", err)
	}

	return suggestions, nil
}

// parseSuggestions extracts a JSON string array from the model response.
// The model is instructed to return only a JSON array; this is resilient to
// any surrounding whitespace or stray characters.
func parseSuggestions(raw string) ([]string, error) {
	start := strings.Index(raw, "[")
	end := strings.LastIndex(raw, "]")
	if start == -1 || end == -1 || end <= start {
		return nil, fmt.Errorf("no JSON array found in response: %q", raw)
	}

	var suggestions []string
	if err := json.Unmarshal([]byte(raw[start:end+1]), &suggestions); err != nil {
		return nil, fmt.Errorf("unmarshal suggestions: %w", err)
	}

	return suggestions, nil
}
