package eval

import (
	"context"
	"time"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
)

// Suite is a named group of eval cases.
type Suite struct {
	Name  string
	Cases []Case
}

// Runner runs suites against an [gentic.Agent] and returns per-case results.
type Runner struct {
	Agent gentic.Agent
}

// Run executes every case in the suite: attaches a [Recorder], runs the agent, builds a [Trace], runs scorers.
func (r Runner) Run(ctx context.Context, suite Suite) []Result {
	if ctx == nil {
		ctx = context.Background()
	}
	out := make([]Result, 0, len(suite.Cases))
	for _, c := range suite.Cases {
		out = append(out, r.runCase(ctx, c))
	}
	return out
}

func (r Runner) runCase(ctx context.Context, c Case) Result {
	rec := NewRecorder()
	ctx = WithRecorder(ctx, rec)
	start := time.Now()
	state, err := r.Agent.RunWithContext(ctx, c.Input)
	dur := time.Since(start)

	tr := Trace{
		Duration: dur,
		Err:      err,
		Steps:    rec.Steps(),
	}
	if state != nil {
		tr.Input = state.Input
		tr.Intent = state.Intent
		tr.Output = state.Output
	} else if c.Input.Query != "" {
		tr.Input = c.Input.Query
	}

	scores := make([]Score, 0, len(c.Scorers))
	allPass := err == nil
	for _, sc := range c.Scorers {
		if sc == nil {
			continue
		}
		sv := sc.Score(&tr)
		scores = append(scores, sv)
		if !sv.Pass {
			allPass = false
		}
	}

	return Result{
		Case:   c,
		Trace:  tr,
		Scores: scores,
		Pass:   allPass,
	}
}
