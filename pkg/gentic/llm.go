package gentic

import "context"

// LLM is the interface for a language model provider.
// Implementations wrap provider-specific APIs (OpenAI, Gemini, etc.).
// All methods accept context for cancellation and deadlines.
type LLM interface {
	Chat(ctx context.Context, model, systemPrompt, userContent string) (string, error)
	// ChatJSON requests a JSON object response and unmarshals it into result (typically a pointer to a struct).
	// Implementations should use the provider’s structured JSON mode when available.
	ChatJSON(ctx context.Context, model, systemPrompt, userContent string, result any) error
}
