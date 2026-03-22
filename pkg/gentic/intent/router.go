package intent

import "github.com/daniel-dihardja/gentic/pkg/gentic"

// compile-time check that *Router satisfies gentic.IntentResolver
var _ gentic.IntentResolver = (*Router)(nil)

// Router classifies the input into one of its labels, sets s.Intent, and dispatches to the matching Flow.
type Router struct {
	labels   []string
	routes   map[string]gentic.Flow
	fallback gentic.Flow
}

// NewRouter creates a Router that will classify input into one of the given labels.
func NewRouter(labels ...string) *Router {
	return &Router{
		labels: labels,
		routes: make(map[string]gentic.Flow),
	}
}

// On registers a Flow for a specific label.
func (r *Router) On(label string, flow gentic.Flow) *Router {
	r.routes[label] = flow
	return r
}

// Default registers a fallback Flow used when no label matches or detection fails.
func (r *Router) Default(flow gentic.Flow) *Router {
	r.fallback = flow
	return r
}

// Resolve implements gentic.IntentResolver.
func (r *Router) Resolve(s *gentic.State) gentic.Flow {
	intent, err := detect(s.Input, r.labels)
	if err != nil || intent == "" {
		return r.fallback
	}

	s.Intent = intent

	if flow, ok := r.routes[intent]; ok {
		return flow
	}
	return r.fallback
}
