package eval

import (
	"context"
	"sync"
)

type recorderCtxKey struct{}

// WithRecorder attaches a Recorder to ctx for [WrapWithEval] and [Runner].
func WithRecorder(ctx context.Context, r *Recorder) context.Context {
	if r == nil {
		return ctx
	}
	return context.WithValue(ctx, recorderCtxKey{}, r)
}

// RecorderFromContext returns the Recorder from ctx, or nil.
func RecorderFromContext(ctx context.Context) *Recorder {
	r, _ := ctx.Value(recorderCtxKey{}).(*Recorder)
	return r
}

// Recorder collects step-level traces (thread-safe).
type Recorder struct {
	mu    sync.Mutex
	steps []StepTrace
}

// NewRecorder creates an empty recorder.
func NewRecorder() *Recorder {
	return &Recorder{}
}

// Steps returns a copy of recorded step traces.
func (r *Recorder) Steps() []StepTrace {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]StepTrace, len(r.steps))
	for i, st := range r.steps {
		out[i] = cloneStepTrace(st)
	}
	return out
}

// Record appends one complete step trace (typically built by [WrapWithEval]). Safe when r is nil (no-op).
func (r *Recorder) Record(t StepTrace) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.steps = append(r.steps, cloneStepTrace(t))
}

func cloneStepTrace(t StepTrace) StepTrace {
	out := t
	if len(t.EvalResults) > 0 {
		out.EvalResults = make([]EvalResult, len(t.EvalResults))
		copy(out.EvalResults, t.EvalResults)
	}
	return out
}
