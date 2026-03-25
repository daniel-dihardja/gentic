package gentic

import "context"

// Step is a single synchronous unit in a Flow.
type Step interface {
	Run(*State) error
}

// StreamingStep is a Step that can produce a token stream (e.g. LLM streaming).
// Non-streaming steps run via Run before the first StreamingStep in a Flow.
type StreamingStep interface {
	Step
	Stream(ctx context.Context, s *State, sllm StreamingLLM) (<-chan StreamEvent, error)
}