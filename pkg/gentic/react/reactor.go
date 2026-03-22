package react

import (
	"encoding/json"

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
- Action Input must be valid JSON matching the tool's input_schema
- Keep thoughts brief (1-2 sentences)
- Don't include the word 'Action:' or 'Action Input:' in your Thought
- Only use one Action per response`

// Tool represents a callable action the ReAct agent can take.
type Tool struct {
	Name        string                                                                                       // name of the tool (e.g., "calculator")
	Description string                                                                                       // human-readable description of what the tool does
	InputSchema json.RawMessage                                                                              // JSON Schema describing the input parameters
	Run         func(state *gentic.State, input json.RawMessage) (json.RawMessage, error)                   // function that executes the tool with state and JSON input/output
	RunCompat   func(input json.RawMessage) (json.RawMessage, error)                                        // backward compatibility: old-style tool without state access
}

// ReactActor implements a Reasoning + Acting loop.
// It satisfies gentic.IntentResolver and is used directly as an Agent resolver.
type ReactActor struct {
	llm                    gentic.LLM
	model                  string
	maxSteps               int
	systemPrompt           string
	tools                  []Tool
	validateMetadataLeaks  bool // if true, warn when tool outputs contain sensitive metadata
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

// WithValidateMetadataLeaks enables warnings when tool outputs may contain sensitive metadata.
// This helps catch tools that accidentally leak private metadata (keys starting with '_').
func WithValidateMetadataLeaks(enabled bool) Option {
	return func(r *ReactActor) { r.validateMetadataLeaks = enabled }
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
func (r *ReactActor) Resolve(_ *gentic.State) gentic.Flow {
	return gentic.NewFlow(
		reactLoopStep{
			llm:                   r.llm,
			model:                 r.model,
			maxSteps:              r.maxSteps,
			systemPrompt:          r.systemPrompt,
			tools:                 r.tools,
			validateMetadataLeaks: r.validateMetadataLeaks,
		},
	)
}

// NewTool is a helper to create a tool with JSON schema and a function that handles JSON input/output.
// Backward compatible: use this for tools that don't need state/metadata access.
// For tools that need state/metadata, use NewToolWithState instead.
// Example:
//
//	NewTool(
//	    "calculator",
//	    "Adds two numbers",
//	    json.RawMessage(`{"type": "object", "properties": {"a": {"type": "number"}, "b": {"type": "number"}}}`),
//	    func(input json.RawMessage) (json.RawMessage, error) {
//	        var params struct{ A float64 `json:"a"`; B float64 `json:"b"` }
//	        if err := json.Unmarshal(input, &params); err != nil {
//	            return nil, err
//	        }
//	        return json.Marshal(map[string]float64{"result": params.A + params.B})
//	    },
//	)
func NewTool(name, description string, inputSchema json.RawMessage, run func(json.RawMessage) (json.RawMessage, error)) Tool {
	return Tool{
		Name:        name,
		Description: description,
		InputSchema: inputSchema,
		RunCompat:   run,
	}
}

// NewToolWithState creates a tool that has access to the State (including metadata).
// Use this when your tool needs to read metadata (user_id, tenant_id, etc.) from the ambient context.
// Example:
//
//	NewToolWithState(
//	    "fetch_analytics",
//	    "Fetches analytics for a product",
//	    json.RawMessage(`{"type": "object", "properties": {"product": {"type": "string"}}}`),
//	    func(state *gentic.State, input json.RawMessage) (json.RawMessage, error) {
//	        analyticsId := state.Metadata["analyticsId"].(string)
//	        // Use analyticsId to fetch data...
//	        return json.Marshal(map[string]interface{}{"data": "..."})
//	    },
//	)
func NewToolWithState(name, description string, inputSchema json.RawMessage, run func(*gentic.State, json.RawMessage) (json.RawMessage, error)) Tool {
	return Tool{
		Name:        name,
		Description: description,
		InputSchema: inputSchema,
		Run:         run,
	}
}
