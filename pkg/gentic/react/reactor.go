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

const defaultSystemPrompt = `You are a helpful assistant that uses tools to gather information and answer questions.

When you need information, follow this format exactly:
Thought: <brief reasoning about what you need>
Action: <tool_name>
Action Input: <valid JSON matching the tool's input_schema>

Once you have enough information to answer, output:
Thought: <final reasoning>
Final Answer: <your complete answer>

Important:
- Use only the tools listed in "Available Tools"
- Action Input must be valid JSON matching the tool's input_schema (use {} when the schema has no properties)
- Keep thoughts brief (1-2 sentences)
- Don't include the word 'Action:' or 'Action Input:' in your Thought
- Only use one Action per response
- You may use markdown bold around labels (e.g. **Action:**) as well as plain Action: — both are parsed`

// Tool represents a callable action the ReAct agent can take.
type Tool struct {
	Name        string
	Description string
	InputSchema json.RawMessage
	Run         func(ctx context.Context, state *gentic.State, input json.RawMessage) (json.RawMessage, error)
	RunCompat   func(ctx context.Context, input json.RawMessage) (json.RawMessage, error)
}

// ReactActor implements a Reasoning + Acting loop.
// It satisfies gentic.IntentResolver and is used directly as an Agent resolver.
type ReactActor struct {
	llm                    gentic.LLM
	model                  string
	maxSteps               int
	systemPrompt           string
	tools                  []Tool
	validateMetadataLeaks  bool
	logger                 *slog.Logger
	toolTimeout            time.Duration
}

// Option configures a ReactActor.
type Option func(*ReactActor)

// WithLLM sets the LLM provider. Defaults to openai.Provider{}.
func WithLLM(llm gentic.LLM) Option {
	return func(r *ReactActor) { r.llm = llm }
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

// WithValidateMetadataLeaks enables warnings when tool outputs contain sensitive metadata.
func WithValidateMetadataLeaks(enabled bool) Option {
	return func(r *ReactActor) { r.validateMetadataLeaks = enabled }
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
		llm:          openai.Provider{},
		model:        openai.DefaultModel,
		maxSteps:     defaultMaxSteps,
		systemPrompt: defaultSystemPrompt,
		tools:        []Tool{},
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Resolve implements gentic.IntentResolver.
// It returns a single-step flow containing the ReAct loop.
func (r *ReactActor) Resolve(_ context.Context, _ *gentic.State) gentic.Flow {
	return gentic.NewFlow(
		reactLoopStep{
			llm:                   r.llm,
			model:                 r.model,
			maxSteps:              r.maxSteps,
			systemPrompt:          r.systemPrompt,
			tools:                 r.tools,
			validateMetadataLeaks: r.validateMetadataLeaks,
			logger:                r.logger,
			toolTimeout:           r.toolTimeout,
		},
	)
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
func NewToolWithState(name, description string, inputSchema json.RawMessage, run func(context.Context, *gentic.State, json.RawMessage) (json.RawMessage, error)) Tool {
	return Tool{
		Name:        name,
		Description: description,
		InputSchema: inputSchema,
		Run:         run,
	}
}
