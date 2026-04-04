package gentic

import (
	"context"
	"encoding/json"
)

// LLM is the interface for a language model provider.
// Implementations wrap provider-specific APIs (OpenAI, Gemini, etc.).
// All methods accept context for cancellation and deadlines.
type LLM interface {
	Chat(ctx context.Context, model, systemPrompt, userContent string) (string, error)
	// ChatJSON requests a JSON object response and unmarshals it into result (typically a pointer to a struct).
	// Implementations should use the provider’s structured JSON mode when available.
	ChatJSON(ctx context.Context, model, systemPrompt, userContent string, result any) error
}

// ToolDefinition is the JSON Schema descriptor sent to the provider’s tool-calling API.
// Maps directly to OpenAI’s function object inside the "tools" array.
type ToolDefinition struct {
	Type     string           `json:"type"` // always "function"
	Function ToolFunctionSpec `json:"function"`
}

// ToolFunctionSpec describes a single callable function.
type ToolFunctionSpec struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"` // JSON Schema object
}

// ToolMessage is a wire-level message for the tool-calling conversation thread.
// Roles: "system", "user", "assistant", "tool".
type ToolMessage struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"` // set on role="tool"
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`   // set on role="assistant" with calls
	Name       string     `json:"name,omitempty"`         // set on role="tool"
}

// ToolCall is a single function call requested by the model in an assistant message.
type ToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"` // always "function"
	Function ToolCallFunction `json:"function"`
}

// ToolCallFunction holds the name and raw JSON arguments for a tool call.
type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // raw JSON string
}

// ToolCallingResponse is the structured response from a ChatWithTools call.
type ToolCallingResponse struct {
	Message      ToolMessage
	FinishReason string // "stop", "tool_calls", "length", etc.
}

// ToolCallingLLM is satisfied by providers that support native function/tool calling.
// It is intentionally separate from LLM so existing resolvers are never perturbed.
type ToolCallingLLM interface {
	ChatWithTools(
		ctx context.Context,
		model string,
		messages []ToolMessage,
		tools []ToolDefinition,
	) (*ToolCallingResponse, error)
}
