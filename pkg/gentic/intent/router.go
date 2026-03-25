package intent

import (
	"context"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
	"github.com/daniel-dihardja/gentic/pkg/providers/openai"
)

// compile-time check that *Router satisfies gentic.IntentResolver
var _ gentic.IntentResolver = (*Router)(nil)

// Router classifies the input into one of its labels, sets s.Intent, and dispatches to the matching Flow.
type Router struct {
	labels   []string
	routes   map[string]gentic.Flow
	fallback gentic.Flow
	llm      gentic.LLM
}

// NewRouter creates a Router that will classify input into one of the given labels.
// The default LLM is openai.Provider; override with WithLLM for tests or alternate providers.
func NewRouter(labels ...string) *Router {
	return &Router{
		labels: labels,
		routes: make(map[string]gentic.Flow),
		llm:    openai.Provider{},
	}
}

// WithLLM sets the classifier LLM (non-nil replaces the default).
func (r *Router) WithLLM(llm gentic.LLM) *Router {
	if llm != nil {
		r.llm = llm
	}
	return r
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
func (r *Router) Resolve(ctx context.Context, s *gentic.State) gentic.Flow {
	intent, err := detect(ctx, r.llm, s.Input, r.labels)
	if err != nil || intent == "" {
		return r.fallback
	}

	s.Intent = intent

	if flow, ok := r.routes[intent]; ok {
		return flow
	}
	return r.fallback
}
