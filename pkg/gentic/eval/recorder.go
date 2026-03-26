package eval

import (
	"context"
	"sync"
	"time"
)

type recorderCtxKey struct{}

// WithRecorder attaches a Recorder to ctx so steps can record spans via [Record].
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
	copy(out, r.steps)
	return out
}

// RecordStep appends one step trace. Safe when r is nil (no-op).
func (r *Recorder) RecordStep(name string, start time.Time, err error) {
	if r == nil {
		return
	}
	st := StepTrace{Name: name, Duration: time.Since(start)}
	if err != nil {
		st.ErrMsg = err.Error()
	}
	r.mu.Lock()
	r.steps = append(r.steps, st)
	r.mu.Unlock()
}

// Record is a convenience for defer: Record(ctx, "step_name", start, err) from named error returns.
func Record(ctx context.Context, stepName string, start time.Time, err error) {
	if r := RecorderFromContext(ctx); r != nil {
		r.RecordStep(stepName, start, err)
	}
}
