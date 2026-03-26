package eval

import (
	"context"
	"fmt"
	"time"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
)

// StepHarness runs a single [gentic.Step] in isolation (tests, debugging).
type StepHarness struct {
	Step gentic.Step
}

// StepResult is the outcome of a [StepHarness.Run].
type StepResult struct {
	State       *gentic.State
	Err         error
	EvalResults []EvalResult
	Pass        bool
	Duration    time.Duration
}

// Run executes the step with the given state, then runs evaluators when Run succeeds.
// Pass is true iff Run returns nil and every [EvalResult].Pass is true.
func (h StepHarness) Run(ctx context.Context, s *gentic.State, evals ...Evaluator) StepResult {
	if ctx == nil {
		ctx = context.Background()
	}
	start := time.Now()
	if h.Step == nil {
		return StepResult{
			State:    s,
			Err:      fmt.Errorf("eval.StepHarness: nil Step"),
			Pass:     false,
			Duration: time.Since(start),
		}
	}
	err := h.Step.Run(ctx, s)
	dur := time.Since(start)

	var results []EvalResult
	pass := err == nil
	if err == nil {
		for _, ev := range evals {
			if ev == nil {
				continue
			}
			er := ev.Evaluate(ctx, s)
			results = append(results, er)
			if !er.Pass {
				pass = false
			}
		}
	}

	return StepResult{
		State:       s,
		Err:         err,
		EvalResults: results,
		Pass:        pass,
		Duration:    dur,
	}
}
