package ollama

import (
	"bufio"
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
			// No client-level timeout — per-request context deadline controls it.
			// A client timeout would fire on the entire response body lifetime,
			// which breaks streaming.
			Timeout: 0,
		},
	}
}

// ── Ollama API types ──────────────────────────────────────────────────────────

type generateOptions struct {
	NumPredict    int     `json:"num_predict"`
	NumCtx        int     `json:"num_ctx"`
	Temperature   float64 `json:"temperature"`
	TopK          int     `json:"top_k"`
	TopP          float64 `json:"top_p"`
	RepeatPenalty float64 `json:"repeat_penalty"`
}

type generateRequest struct {
	Model     string          `json:"model"`
	System    string          `json:"system"`
	Prompt    string          `json:"prompt"`
	Stream    bool            `json:"stream"`
	KeepAlive int             `json:"keep_alive"` // -1 = keep model loaded indefinitely
	Format    string          `json:"format"`     // "json" constrains output to valid JSON
	Options   generateOptions `json:"options"`
}

// streamChunk is one line of the NDJSON streaming response.
type streamChunk struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// ── LLMProvider implementation ────────────────────────────────────────────────

func (p *Provider) Complete(ctx context.Context, req provider.CompletionRequest) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	body, err := json.Marshal(generateRequest{
		Model:     p.model,
		System:    req.SystemPrompt,
		Prompt:    req.UserPrompt,
		Stream:    true,
		KeepAlive: -1,
		Format:    "json",
		Options: generateOptions{
			NumPredict:    256,
			NumCtx:        1024,
			Temperature:   0.6,
			TopK:          20,
			TopP:          0.9,
			RepeatPenalty: 1.0,
		},
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
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("ollama: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("ollama: unexpected status %d: %s", resp.StatusCode, raw)
	}

	return p.readStream(ctx, resp.Body)
}

// readStream reads the NDJSON streaming response from Ollama, accumulating
// token chunks until done:true, then parses the full response as suggestions.
//
// Cancellation path: when ctx is cancelled, the underlying TCP connection is
// torn down, scanner.Scan() returns false, and we return ctx.Err() — Ollama
// sees a broken pipe and stops inference immediately.
func (p *Provider) readStream(ctx context.Context, body io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(body)
	var sb strings.Builder

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var chunk streamChunk
		if err := json.Unmarshal(line, &chunk); err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			return nil, fmt.Errorf("ollama: decode stream chunk: %w", err)
		}

		sb.WriteString(chunk.Response)

		if chunk.Done {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("ollama: read stream: %w", err)
	}

	suggestions, err := parseSuggestions(sb.String())
	if err != nil {
		return nil, fmt.Errorf("ollama: parse suggestions: %w", err)
	}

	return suggestions, nil
}

// parseSuggestions extracts a JSON string array from the model response.
// Handles three formats models produce:
//  1. Plain array:  ["s1","s2","s3"]
//  2. Object with an array field: {"completions":["s1","s2","s3"],...}
//  3. Object with string values: {"1":"s1","2":"s2","3":"s3"}
func parseSuggestions(raw string) ([]string, error) {
	raw = strings.TrimSpace(raw)

	// Try top-level array first.
	if start := strings.Index(raw, "["); start != -1 {
		if end := strings.LastIndex(raw, "]"); end > start {
			var arr []string
			if err := json.Unmarshal([]byte(raw[start:end+1]), &arr); err == nil {
				if filtered := filterSuggestions(arr); len(filtered) > 0 {
					return filtered, nil
				}
			}
		}
	}

	// Fallback: model returned a JSON object.
	if start := strings.Index(raw, "{"); start != -1 {
		if end := strings.LastIndex(raw, "}"); end > start {
			var obj map[string]json.RawMessage
			if err := json.Unmarshal([]byte(raw[start:end+1]), &obj); err == nil {
				// Look for any field that is a string array.
				for _, v := range obj {
					var arr []string
					if err := json.Unmarshal(v, &arr); err == nil {
						if filtered := filterSuggestions(arr); len(filtered) > 0 {
							return filtered, nil
						}
					}
				}
				// Last resort: extract plain string values from the object.
				var strObj map[string]string
				if err := json.Unmarshal([]byte(raw[start:end+1]), &strObj); err == nil {
					var suggestions []string
					for _, v := range strObj {
						if s := strings.TrimSpace(v); s != "" {
							suggestions = append(suggestions, s)
						}
					}
					if len(suggestions) > 0 {
						return suggestions, nil
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("no JSON array found in response: %q", raw)
}

// filterSuggestions removes blank or whitespace-only strings.
func filterSuggestions(suggestions []string) []string {
	result := make([]string, 0, len(suggestions))
	for _, s := range suggestions {
		if trimmed := strings.TrimSpace(s); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
