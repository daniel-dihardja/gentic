package reflect

import (
	"context"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
	"github.com/daniel-dihardja/gentic/pkg/providers/openai"
)

// compile-time check that *Reflector satisfies gentic.IntentResolver
var _ gentic.IntentResolver = (*Reflector)(nil)

const defaultMaxIterations = 3

const defaultGeneratePrompt = `You are a skilled writer and analyst. Produce a high-quality, complete response to the user's request.
If you are given a previous draft and a critique, revise the draft to address every point raised.`

const defaultCritiquePrompt = `You are a rigorous editor. Evaluate the draft below against the original request.
Respond with a single JSON object only (no markdown, no extra text).
If the draft fully and correctly addresses the request: {"verdict":"PASS"}
Otherwise: {"verdict":"IMPROVE","feedback":"<concise bullet-point issues as one string, newlines between items>"}
Use verdict values exactly PASS or IMPROVE.`

// Reflector implements a generate→critique→refine loop.
// It satisfies gentic.IntentResolver and is used directly as an Agent resolver.
type Reflector struct {
	llm                 gentic.LLM
	model               string
	maxIterations       int
	generatePrompt      string
	critiquePrompt      string
	critiqueUserBuilder func(input, draft string) string
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

// WithCritiqueUserBuilder overrides how the critique user prompt is assembled.
// fn receives (input, draft) and returns the full user content for the critique LLM call.
// When nil, the default is Original request + Draft (see reflectionLoopStep).
func WithCritiqueUserBuilder(fn func(input, draft string) string) Option {
	return func(r *Reflector) { r.critiqueUserBuilder = fn }
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

func loopStepFromReflector(r *Reflector) reflectionLoopStep {
	return reflectionLoopStep{
		llm:                 r.llm,
		model:               r.model,
		maxIterations:       r.maxIterations,
		generatePrompt:      r.generatePrompt,
		critiquePrompt:      r.critiquePrompt,
		critiqueUserBuilder: r.critiqueUserBuilder,
	}
}

// NewLoopStep returns a [gentic.Step] that runs the same generate→critique loop as [Reflector.Resolve],
// so you can embed the loop in a larger [gentic.Flow] without nesting a full [gentic.Agent].
func NewLoopStep(opts ...Option) gentic.Step {
	return loopStepFromReflector(NewReflector(opts...))
}

// Resolve implements gentic.IntentResolver.
// It returns a single-step flow containing the reflection loop.
func (r *Reflector) Resolve(_ context.Context, _ *gentic.State) gentic.Flow {
	return gentic.NewFlow(loopStepFromReflector(r))
}
