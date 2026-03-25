package steps

import (
	"context"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
	"github.com/daniel-dihardja/gentic/pkg/providers/openai"
)

// ChatStep runs a single LLM chat completion and writes the reply to state.
// It implements both gentic.Step and gentic.StreamingStep.
type ChatStep struct {
	LLM          gentic.LLM // optional; defaults to openai.Provider{}
	Model        string
	SystemPrompt string
}

func (c ChatStep) llm() gentic.LLM {
	if c.LLM != nil {
		return c.LLM
	}
	return openai.Provider{}
}

// Stream implements gentic.StreamingStep — streams tokens from the configured model.
func (c ChatStep) Stream(ctx context.Context, s *gentic.State, sllm gentic.StreamingLLM) (<-chan gentic.StreamEvent, error) {
	model := c.Model
	if model == "" {
		model = openai.DefaultModel
	}
	return sllm.ChatStream(ctx, model, c.SystemPrompt, s.Input)
}

// Run implements gentic.Step.
func (c ChatStep) Run(s *gentic.State) error {
	model := c.Model
	if model == "" {
		model = openai.DefaultModel
	}
	out, err := c.llm().Chat(model, c.SystemPrompt, s.Input)
	if err != nil {
		return err
	}
	s.Output = out
	return nil
}
