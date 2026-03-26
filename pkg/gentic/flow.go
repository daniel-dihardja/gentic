package gentic

import (
	"context"
	"sync"
)

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

// Parallel runs each step with the same [State] concurrently. The first error returned by any step wins.
// Steps must not perform conflicting concurrent writes to [State].
func Parallel(steps ...Step) Step {
	return parallelStep{steps: steps}
}

type parallelStep struct {
	steps []Step
}

func (p parallelStep) Run(ctx context.Context, s *State) error {
	if len(p.steps) == 0 {
		return nil
	}
	var wg sync.WaitGroup
	errCh := make(chan error, len(p.steps))
	for _, st := range p.steps {
		if st == nil {
			continue
		}
		st := st
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := st.Run(ctx, s); err != nil {
				errCh <- err
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
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
