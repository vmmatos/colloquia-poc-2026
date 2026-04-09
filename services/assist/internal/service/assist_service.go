package service

import (
	"context"
	"fmt"
	"log"
	"strings"

	"assist/internal/messagingclient"
	"assist/internal/provider"
)

const (
	defaultMessageLimit = 10
	maxMessageLimit     = 10

	systemPrompt = `You are a helpful chat assistant. Given the recent conversation history and the user's partial message, suggest exactly 3 natural and concise message completions in the same language as the conversation.
You may include emojis in suggestions when they feel natural and match the tone of the conversation.
Return ONLY a valid JSON array of 3 strings, like: ["completion 1", "completion 2", "completion 3"].
No explanation, no markdown, no extra text.`
)

type AssistService struct {
	msgClient messagingclient.MessageFetcher
	provider  provider.LLMProvider
}

func NewAssistService(msgClient messagingclient.MessageFetcher, p provider.LLMProvider) *AssistService {
	return &AssistService{msgClient: msgClient, provider: p}
}

func (s *AssistService) GetSuggestions(ctx context.Context, channelID, currentInput string, messageLimit int32) ([]string, error) {
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
				contextLines = append(contextLines, fmt.Sprintf("%s: %s", m.GetUserId(), m.GetContent()))
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
