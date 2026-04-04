package react

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
	"github.com/daniel-dihardja/gentic/pkg/providers/openai"
)

// compile-time check that *ReactActor satisfies gentic.IntentResolver
var _ gentic.IntentResolver = (*ReactActor)(nil)

const defaultMaxSteps = 10

const defaultSystemPrompt = `You are a helpful assistant. Use the provided tools to gather information and complete tasks. When you have enough information to fully answer the user's request, respond directly without calling any tools.`

// GuardFunc runs before a tool handler (after the model chose the tool). Use it for prerequisites:
// validating ambient metadata, lazy-loading shared state, etc. If it returns a non-nil error,
// the tool handler is skipped and the error is returned to the model as the tool result.
type GuardFunc func(ctx context.Context, state *gentic.State) error

// Tool represents a callable action the ReAct agent can take.
type Tool struct {
	Name        string
	Description string
	InputSchema json.RawMessage
	Run         func(ctx context.Context, state *gentic.State, input json.RawMessage) (json.RawMessage, error)
	RunCompat   func(ctx context.Context, input json.RawMessage) (json.RawMessage, error)
	Guards      []GuardFunc
}

// ReactActor implements a Reasoning + Acting loop.
// It satisfies gentic.IntentResolver and is used directly as an Agent resolver.
type ReactActor struct {
	toolCallingLLM gentic.ToolCallingLLM
	model          string
	maxSteps       int
	systemPrompt   string
	tools          []Tool
	logger         *slog.Logger
	toolTimeout    time.Duration
}

// Option configures a ReactActor.
type Option func(*ReactActor)

// WithLLM sets the tool-calling LLM provider. Defaults to [openai.Provider].
func WithLLM(llm gentic.ToolCallingLLM) Option {
	return func(r *ReactActor) { r.toolCallingLLM = llm }
}

// WithModel overrides the model used for LLM calls.
func WithModel(model string) Option {
	return func(r *ReactActor) { r.model = model }
}

// WithMaxSteps sets the maximum number of thought→action cycles.
func WithMaxSteps(n int) Option {
	return func(r *ReactActor) { r.maxSteps = n }
}

// WithSystemPrompt overrides the system prompt.
func WithSystemPrompt(prompt string) Option {
	return func(r *ReactActor) { r.systemPrompt = prompt }
}

// WithTools sets the available tools.
func WithTools(tools ...Tool) Option {
	return func(r *ReactActor) { r.tools = tools }
}

// WithLogger sets structured logging for the ReAct loop.
// If nil, the loop still emits INFO traces via slog.Default() (component gentic.react).
func WithLogger(l *slog.Logger) Option {
	return func(r *ReactActor) { r.logger = l }
}

// WithToolTimeout caps each tool invocation with context.WithTimeout. Zero disables.
func WithToolTimeout(d time.Duration) Option {
	return func(r *ReactActor) { r.toolTimeout = d }
}

// NewReactActor creates a ReactActor ready to use as a gentic.Agent resolver.
func NewReactActor(opts ...Option) *ReactActor {
	r := &ReactActor{
		toolCallingLLM: openai.Provider{},
		model:          openai.DefaultModel,
		maxSteps:       defaultMaxSteps,
		systemPrompt:   defaultSystemPrompt,
		tools:          []Tool{},
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Flow returns the single-step flow that runs the ReAct loop (same as [ReactActor.Resolve]).
func (r *ReactActor) Flow() gentic.Flow {
	return gentic.NewFlow(r.asLoopStep())
}

func (r *ReactActor) asLoopStep() reactLoopStep {
	return reactLoopStep{
		toolCallingLLM: r.toolCallingLLM,
		model:          r.model,
		maxSteps:       r.maxSteps,
		systemPrompt:   r.systemPrompt,
		tools:          r.tools,
		logger:         r.logger,
		toolTimeout:    r.toolTimeout,
	}
}

// Resolve implements gentic.IntentResolver.
// It returns a single-step flow containing the ReAct loop.
func (r *ReactActor) Resolve(_ context.Context, _ *gentic.State) gentic.Flow {
	return r.Flow()
}

// NewTool is a helper for tools that do not need state/metadata access.
func NewTool(name, description string, inputSchema json.RawMessage, run func(context.Context, json.RawMessage) (json.RawMessage, error)) Tool {
	return Tool{
		Name:        name,
		Description: description,
		InputSchema: inputSchema,
		RunCompat:   run,
	}
}

// NewToolWithState creates a tool that has access to the State (including metadata).
// Optional guards run in order before run; see [GuardFunc].
func NewToolWithState(name, description string, inputSchema json.RawMessage, run func(context.Context, *gentic.State, json.RawMessage) (json.RawMessage, error), guards ...GuardFunc) Tool {
	return Tool{
		Name:        name,
		Description: description,
		InputSchema: inputSchema,
		Run:         run,
		Guards:      guards,
	}
}
