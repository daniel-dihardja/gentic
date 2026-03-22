package reflect

import (
	"github.com/daniel-dihardja/gentic/pkg/gentic"
	"github.com/daniel-dihardja/gentic/pkg/providers/openai"
)

// compile-time check that *Reflector satisfies gentic.IntentResolver
var _ gentic.IntentResolver = (*Reflector)(nil)

const defaultMaxIterations = 3

const defaultGeneratePrompt = `You are a skilled writer and analyst. Produce a high-quality, complete response to the user's request.
If you are given a previous draft and a critique, revise the draft to address every point raised.`

const defaultCritiquePrompt = `You are a rigorous editor. Evaluate the draft below against the original request.
If the draft fully and correctly addresses the request, reply with exactly: PASS
Otherwise, reply with: IMPROVE: <concise bullet-point list of specific issues to fix>
Do not add any other text before or after.`

// Reflector implements a generate→critique→refine loop.
// It satisfies gentic.IntentResolver and is used directly as an Agent resolver.
type Reflector struct {
	llm            gentic.LLM
	model          string
	maxIterations  int
	generatePrompt string
	critiquePrompt string
}

// Option configures a Reflector.
type Option func(*Reflector)

// WithLLM sets the LLM provider. Defaults to openai.Provider{}.
func WithLLM(llm gentic.LLM) Option {
	return func(r *Reflector) { r.llm = llm }
}

// WithModel overrides the model used for both generate and critique calls.
func WithModel(model string) Option {
	return func(r *Reflector) { r.model = model }
}

// WithMaxIterations sets the maximum number of generate→critique cycles.
func WithMaxIterations(n int) Option {
	return func(r *Reflector) { r.maxIterations = n }
}

// WithGeneratePrompt overrides the system prompt used for the generate LLM call.
func WithGeneratePrompt(prompt string) Option {
	return func(r *Reflector) { r.generatePrompt = prompt }
}

// WithCritiquePrompt overrides the system prompt used for the critique LLM call.
func WithCritiquePrompt(prompt string) Option {
	return func(r *Reflector) { r.critiquePrompt = prompt }
}

// NewReflector creates a Reflector ready to use as a gentic.Agent resolver.
func NewReflector(opts ...Option) *Reflector {
	r := &Reflector{
		llm:            openai.Provider{},
		model:          openai.DefaultModel,
		maxIterations:  defaultMaxIterations,
		generatePrompt: defaultGeneratePrompt,
		critiquePrompt: defaultCritiquePrompt,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Resolve implements gentic.IntentResolver.
// It returns a single-step flow containing the reflection loop.
func (r *Reflector) Resolve(_ *gentic.State) gentic.Flow {
	return gentic.NewFlow(
		reflectionLoopStep{
			llm:            r.llm,
			model:          r.model,
			maxIterations:  r.maxIterations,
			generatePrompt: r.generatePrompt,
			critiquePrompt: r.critiquePrompt,
		},
	)
}
