package gentic

import "context"

// ActivityStatus is the lifecycle state of a single logical step shown in the UI.
type ActivityStatus string

const (
	ActivityRunning       ActivityStatus = "running"
	ActivityDone          ActivityStatus = "done"
	ActivityReflecting    ActivityStatus = "reflecting"
	ActivityReflectPass   ActivityStatus = "reflect_pass"
	ActivityReflectRevise ActivityStatus = "reflect_revise"
)

// ActivityEvent is the wire protocol for agent activity (mirrors the web AgentActivityFeed).
type ActivityEvent struct {
	Step      string         `json:"step"`
	Status    ActivityStatus `json:"status"`
	Label     string         `json:"label"`
	Detail    string         `json:"detail,omitempty"`
	Transient bool           `json:"transient,omitempty"`
}

// Notifier sends activity updates on the same channel as token stream events.
// A nil *Notifier is safe to use: Notify is a no-op.
type Notifier struct {
	ch chan<- StreamEvent
}

type notifyOptions struct {
	detail    string
	transient bool
}

// NotifyOption configures optional fields on an activity event.
type NotifyOption func(*notifyOptions)

// WithDetail sets a short detail string (e.g. entity name) for the UI.
func WithDetail(d string) NotifyOption {
	return func(o *notifyOptions) { o.detail = d }
}

// WithTransient marks the step as hidden after the stream ends (e.g. interim "thinking").
func WithTransient(t bool) NotifyOption {
	return func(o *notifyOptions) { o.transient = t }
}

// Notify emits one activity event. Safe when n or n.ch is nil (e.g. Invoke path without streaming).
func (n *Notifier) Notify(step string, status ActivityStatus, label string, opts ...NotifyOption) {
	if n == nil || n.ch == nil {
		return
	}
	var o notifyOptions
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}
	ev := ActivityEvent{
		Step:      step,
		Status:    status,
		Label:     label,
		Detail:    o.detail,
		Transient: o.transient,
	}
	n.ch <- StreamEvent{Activity: &ev}
}

// EmitData sends an auxiliary named payload on the stream (e.g. planning artifact JSON).
// Safe when n or n.ch is nil (e.g. Invoke path without streaming).
func (n *Notifier) EmitData(name string, payload interface{}) {
	if n == nil || n.ch == nil {
		return
	}
	n.ch <- StreamEvent{DataName: name, DataPayload: payload}
}

type notifierCtxKey struct{}

// WithNotifier attaches a Notifier to ctx for use inside steps.
func WithNotifier(ctx context.Context, n *Notifier) context.Context {
	return context.WithValue(ctx, notifierCtxKey{}, n)
}

// NotifierFromContext returns the Notifier from ctx, or nil.
func NotifierFromContext(ctx context.Context) *Notifier {
	n, _ := ctx.Value(notifierCtxKey{}).(*Notifier)
	return n
}
