package gentic

import "context"

type Flow struct {
	steps []Step
}

func NewFlow(steps ...Step) Flow {
	return Flow{steps: steps}
}

// IsEmpty reports whether the flow has no steps.
func (f Flow) IsEmpty() bool {
	return len(f.steps) == 0
}

func (f Flow) Run(ctx context.Context, s *State) error {
	for _, step := range f.steps {
		if err := step.Run(ctx, s); err != nil {
			return err
		}
	}
	return nil
}

// If runs thenStep only when predicate(state) is true. Use for "check → act if missing" flows
// without branching logic inside a step’s Run body.
func If(predicate func(*State) bool, thenStep Step) Step {
	return conditionalStep{predicate: predicate, then: thenStep}
}

type conditionalStep struct {
	predicate func(*State) bool
	then      Step
}

func (c conditionalStep) Run(ctx context.Context, s *State) error {
	if c.predicate == nil || !c.predicate(s) {
		return nil
	}
	return c.then.Run(ctx, s)
}

// Stream runs steps until a StreamingStep is found; that step owns the stream.
// If no StreamingStep exists, synchronous steps run in order and the final output
// is wrapped as a short synthetic stream (text + done).
// Errors are sent as StreamEvent{Token: StreamToken{Error: err}}.
func (f Flow) Stream(ctx context.Context, s *State, sllm StreamingLLM) <-chan StreamEvent {
	out := make(chan StreamEvent, 16)
	notifier := &Notifier{ch: out}
	ctx = WithNotifier(ctx, notifier)

	go func() {
		defer close(out)
		for _, step := range f.steps {
			if ss, ok := step.(StreamingStep); ok {
				for ev := range ss.Stream(ctx, s, sllm) {
					out <- ev
				}
				return
			}
			if err := step.Run(ctx, s); err != nil {
				out <- StreamEvent{Token: StreamToken{Error: err}}
				return
			}
		}
		if s.Output != "" {
			out <- StreamEvent{Token: StreamToken{Text: s.Output}}
		}
		out <- StreamEvent{Token: StreamToken{Done: true}}
	}()
	return out
}
