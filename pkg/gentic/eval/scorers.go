package eval

import (
	"fmt"
	"strings"
	"time"
)

// IntentIs passes when [Trace.Intent] equals expected (exact match).
type IntentIs struct {
	Expected string
}

func (s IntentIs) Score(t *Trace) Score {
	name := "intent"
	if t == nil {
		return Score{Name: name, Pass: false, Reason: "nil trace"}
	}
	if t.Intent != s.Expected {
		return Score{
			Name:   name,
			Pass:   false,
			Reason: fmt.Sprintf("got %q want %q", t.Intent, s.Expected),
		}
	}
	return Score{Name: name, Pass: true, Value: 1}
}

// OutputContains passes when [Trace.Output] contains substr (case-sensitive).
type OutputContains struct {
	Substr string
}

func (s OutputContains) Score(t *Trace) Score {
	name := "output_contains"
	if t == nil {
		return Score{Name: name, Pass: false, Reason: "nil trace"}
	}
	if !strings.Contains(t.Output, s.Substr) {
		return Score{
			Name:   name,
			Pass:   false,
			Reason: fmt.Sprintf("output does not contain %q", s.Substr),
		}
	}
	return Score{Name: name, Pass: true, Value: 1}
}

// NoError passes when [Trace.Err] is nil.
type NoError struct{}

func (NoError) Score(t *Trace) Score {
	name := "no_error"
	if t == nil {
		return Score{Name: name, Pass: false, Reason: "nil trace"}
	}
	if t.Err != nil {
		return Score{Name: name, Pass: false, Reason: t.Err.Error()}
	}
	return Score{Name: name, Pass: true, Value: 1}
}

// MaxDuration passes when [Trace.Duration] is at most the limit.
type MaxDuration struct {
	Limit time.Duration
}

func (s MaxDuration) Score(t *Trace) Score {
	name := "max_duration"
	if t == nil {
		return Score{Name: name, Pass: false, Reason: "nil trace"}
	}
	if t.Duration > s.Limit {
		return Score{
			Name:   name,
			Pass:   false,
			Reason: fmt.Sprintf("took %v, limit %v", t.Duration, s.Limit),
		}
	}
	return Score{Name: name, Pass: true, Value: 1}
}
