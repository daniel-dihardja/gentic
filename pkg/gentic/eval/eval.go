package eval

import (
	"time"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
)

// StepTrace is one executed step captured during an eval run (timing, error, per-step eval results).
type StepTrace struct {
	Name        string
	Duration    time.Duration
	Err         error
	EvalResults []EvalResult
}

// Trace is the full record of one agent invocation for scoring and debugging.
type Trace struct {
	Input    string
	Intent   string
	Output   string
	Steps    []StepTrace
	Duration time.Duration
	Err      error
}

// Score is one assertion result against a trace.
type Score struct {
	Name   string
	Pass   bool
	Value  float64
	Reason string
}

// Scorer evaluates a full [Trace] (flow-level assertions).
type Scorer interface {
	Score(t *Trace) Score
}

// Case is a single eval scenario: input to the agent and scorers to run afterward.
type Case struct {
	Name    string
	Input   gentic.AgentInput
	Scorers []Scorer
}

// Result aggregates the trace and per-scorer outcomes for one case.
type Result struct {
	Case   Case
	Trace  Trace
	Scores []Score
	Pass   bool
}
