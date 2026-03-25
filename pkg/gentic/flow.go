package gentic

import "context"

type Flow struct {
	steps []Step
}

func NewFlow(steps ...Step) Flow {
	return Flow{steps: steps}
}

// IsEmpty reports whether the flow has no steps (used for legacy direct-LLM streaming).
func (f Flow) IsEmpty() bool {
	return len(f.steps) == 0
}

func (f Flow) Run(s *State) error {
	for _, step := range f.steps {
		if err := step.Run(s); err != nil {
			return err
		}
	}
	return nil
}

// Stream runs steps until a StreamingStep is found; that step owns the stream.
// If no StreamingStep exists, synchronous steps run in order and the final output
// is wrapped as a short synthetic stream (text + done).
func (f Flow) Stream(ctx context.Context, s *State, sllm StreamingLLM) (<-chan StreamEvent, error) {
	for _, step := range f.steps {
		if ss, ok := step.(StreamingStep); ok {
			return ss.Stream(ctx, s, sllm)
		}
		if err := step.Run(s); err != nil {
			return nil, err
		}
	}
	return syntheticStream(s.Output), nil
}

func syntheticStream(output string) <-chan StreamEvent {
	out := make(chan StreamEvent, 2)
	go func() {
		defer close(out)
		if output != "" {
			out <- StreamEvent{Token: StreamToken{Text: output}}
		}
		out <- StreamEvent{Token: StreamToken{Done: true}}
	}()
	return out
}