package provider

import "context"

// CompletionRequest holds the prompts sent to the LLM.
type CompletionRequest struct {
	SystemPrompt string
	UserPrompt   string
}

// LLMProvider is the abstraction over any LLM backend.
// Swap the implementation (Ollama → OpenAI → Anthropic) without touching the rest of the code.
type LLMProvider interface {
	Complete(ctx context.Context, req CompletionRequest) ([]string, error)
}
