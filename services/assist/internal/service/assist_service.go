package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"assist/internal/messagingclient"
	"assist/internal/provider"
)

const (
	defaultMessageLimit  = 5
	maxMessageLimit      = 5
	maxContextLineLength = 200

	systemPrompt = `Suggest 3 short message completions in the same language as the conversation. You may include emojis when natural.
Output ONLY a JSON array of 3 strings: ["completion 1", "completion 2", "completion 3"]. No explanation, no markdown.`
)

type cacheEntry struct {
	suggestions []string
	expiresAt   time.Time
}

type AssistService struct {
	msgClient messagingclient.MessageFetcher
	provider  provider.LLMProvider
	cache     sync.Map // key: string → *cacheEntry
	cacheTTL  time.Duration
}

func NewAssistService(msgClient messagingclient.MessageFetcher, p provider.LLMProvider) *AssistService {
	return &AssistService{
		msgClient: msgClient,
		provider:  p,
		cacheTTL:  30 * time.Second,
	}
}

func (s *AssistService) GetSuggestions(ctx context.Context, channelID, currentInput string, messageLimit int32) ([]string, error) {
	// Cache key: normalize input so mid-word typing hits the cache more often.
	cacheKey := channelID + "|" + strings.ToLower(strings.TrimSpace(currentInput))

	// Lazy TTL check on read — no background goroutine needed.
	if v, ok := s.cache.Load(cacheKey); ok {
		entry := v.(*cacheEntry)
		if time.Now().Before(entry.expiresAt) {
			return entry.suggestions, nil
		}
		s.cache.Delete(cacheKey) // expired
	}

	// 1. Fetch recent messages for context (best-effort).
	var contextLines []string
	if s.msgClient != nil {
		limit := messageLimit
		if limit <= 0 || limit > maxMessageLimit {
			limit = defaultMessageLimit
		}
		msgs, err := s.msgClient.GetMessages(ctx, channelID, limit)
		if err != nil {
			log.Printf("assist: warning: could not fetch messages for context: %v", err)
		} else {
			for _, m := range msgs {
				line := fmt.Sprintf("%s: %s", m.GetUserId(), m.GetContent())
				if len(line) > maxContextLineLength {
					line = line[:maxContextLineLength]
				}
				contextLines = append(contextLines, line)
			}
		}
	}

	// 2. Build user prompt.
	userPrompt := buildUserPrompt(contextLines, currentInput)

	// 3. Call provider (timeout + context cancel handled inside).
	suggestions, err := s.provider.Complete(ctx, provider.CompletionRequest{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
	})
	if err != nil {
		return nil, fmt.Errorf("assist: provider: %w", err)
	}

	// Store in cache.
	s.cache.Store(cacheKey, &cacheEntry{
		suggestions: suggestions,
		expiresAt:   time.Now().Add(s.cacheTTL),
	})

	return suggestions, nil
}

func buildUserPrompt(contextLines []string, currentInput string) string {
	var sb strings.Builder

	if len(contextLines) > 0 {
		sb.WriteString("Recent conversation:\n")
		for _, line := range contextLines {
			sb.WriteString(line)
			sb.WriteByte('\n')
		}
		sb.WriteByte('\n')
	}

	sb.WriteString(`Complete this message: "`)
	sb.WriteString(currentInput)
	sb.WriteByte('"')

	return sb.String()
}
