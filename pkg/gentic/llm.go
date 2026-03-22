package gentic

// LLM is the interface for a language model provider.
// Implementations wrap provider-specific APIs (OpenAI, Gemini, etc.).
type LLM interface {
	Chat(model, systemPrompt, userContent string) (string, error)
}
