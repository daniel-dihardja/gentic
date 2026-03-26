package eval

import (
	"time"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
)

// StepTrace is one executed step captured during an eval run (name + timing + error).
type StepTrace struct {
	Name     string
	Duration time.Duration
	ErrMsg   string
}

// Trace is the full record of one agent invocation for scoring and debugging.
type Trace struct {
	Input    string
	Intent   string
	Steps    []StepTrace
	Output   string
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

// Scorer evaluates a trace (e.g. intent match, substring in output).
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
