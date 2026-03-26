package eval

import (
	"context"
	"time"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
)

type evalWrappedStep struct {
	name  string
	inner gentic.Step
	evals []Evaluator
}

// WrapWithEval wraps a [gentic.Step]: times execution, runs evaluators after a successful Run,
// records a [StepTrace] on the [Recorder] in ctx (if any), and returns the inner error unchanged.
// Evaluator failures do not change the returned error from the inner step.
func WrapWithEval(name string, step gentic.Step, evals ...Evaluator) gentic.Step {
	if step == nil {
		return nil
	}
	evs := make([]Evaluator, 0, len(evals))
	for _, e := range evals {
		if e != nil {
			evs = append(evs, e)
		}
	}
	return evalWrappedStep{name: name, inner: step, evals: evs}
}

func (w evalWrappedStep) Run(ctx context.Context, s *gentic.State) error {
	start := time.Now()
	err := w.inner.Run(ctx, s)
	dur := time.Since(start)

	var results []EvalResult
	if err == nil {
		for _, ev := range w.evals {
			results = append(results, ev.Evaluate(ctx, s))
		}
	}

	tr := StepTrace{
		Name:        w.name,
		Duration:    dur,
		Err:         err,
		EvalResults: results,
	}
	if r := RecorderFromContext(ctx); r != nil {
		r.Record(tr)
	}
	return err
}
