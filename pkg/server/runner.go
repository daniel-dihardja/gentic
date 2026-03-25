package server

import (
	"context"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
)

// Runner executes a [gentic.Agent] with request-scoped metadata from [InvokeRequest].
type Runner struct {
	agent gentic.Agent
	sllm  gentic.StreamingLLM
}

// NewRunner constructs a Runner with a shared agent and streaming provider.
func NewRunner(agent gentic.Agent, sllm gentic.StreamingLLM) *Runner {
	return &Runner{agent: agent, sllm: sllm}
}

// Invoke runs the agent for a single invoke request (batch completion).
func (r *Runner) Invoke(ctx context.Context, req InvokeRequest) (*InvokeResponse, error) {
	input := req.AgentInput()
	state, err := r.agent.RunWithContext(ctx, input)
	if err != nil {
		return nil, err
	}
	return &InvokeResponse{
		OK:     true,
		Output: state.Output,
		Intent: state.Intent,
	}, nil
}

// Stream uses the same Resolver → Flow pipeline as Invoke.
func (r *Runner) Stream(ctx context.Context, req InvokeRequest) <-chan gentic.StreamEvent {
	return r.agent.StreamWithContext(ctx, req.AgentInput(), r.sllm)
}
