package gentic

import "context"

// StreamToken is a single text delta arriving from the LLM.
type StreamToken struct {
	Text  string // the token text; empty on finish
	Done  bool   // true on the final (sentinel) token
	Error error  // non-nil if the stream broke
}

// StreamEvent wraps a StreamToken with usage metadata populated
// only when Done == true.
type StreamEvent struct {
	Token            StreamToken
	Activity         *ActivityEvent
	PromptTokens     int
	CompletionTokens int
	FinishReason     string // "stop", "length", "error"
}

// StreamingLLM is satisfied by providers that support token streaming.
// It is intentionally separate from LLM so existing code paths are never perturbed.
type StreamingLLM interface {
	ChatStream(ctx context.Context, model, systemPrompt, userContent string) (<-chan StreamEvent, error)
}
