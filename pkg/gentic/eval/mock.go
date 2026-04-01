package eval

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
)

// MockLLM implements [gentic.LLM] with injectable functions for tests and evals.
type MockLLM struct {
	ChatFunc     func(ctx context.Context, model, systemPrompt, userContent string) (string, error)
	ChatJSONFunc func(ctx context.Context, model, systemPrompt, userContent string, result any) error
}

var _ gentic.LLM = (*MockLLM)(nil)

// Chat delegates to ChatFunc, or returns "", nil if unset.
func (m *MockLLM) Chat(ctx context.Context, model, systemPrompt, userContent string) (string, error) {
	if m != nil && m.ChatFunc != nil {
		return m.ChatFunc(ctx, model, systemPrompt, userContent)
	}
	return "", nil
}

// ChatJSON delegates to ChatJSONFunc, or returns an error if unset.
func (m *MockLLM) ChatJSON(ctx context.Context, model, systemPrompt, userContent string, result any) error {
	if m != nil && m.ChatJSONFunc != nil {
		return m.ChatJSONFunc(ctx, model, systemPrompt, userContent, result)
	}
	return fmt.Errorf("eval.MockLLM: ChatJSON not configured")
}

// ReplyChat returns a [MockLLM] whose Chat always returns reply and err.
func ReplyChat(reply string, err error) *MockLLM {
	return &MockLLM{
		ChatFunc: func(context.Context, string, string, string) (string, error) {
			return reply, err
		},
	}
}

// ReplyJSON unmarshals jsonStr into result (must be a non-nil pointer).
func ReplyJSON(jsonStr string, err error) *MockLLM {
	return &MockLLM{
		ChatJSONFunc: func(_ context.Context, _, _, _ string, result any) error {
			if err != nil {
				return err
			}
			return json.Unmarshal([]byte(jsonStr), result)
		},
	}
}

// MockToolCallingLLM implements [gentic.ToolCallingLLM] for tests.
type MockToolCallingLLM struct {
	ChatWithToolsFunc func(ctx context.Context, model string, messages []gentic.ToolMessage, tools []gentic.ToolDefinition) (*gentic.ToolCallingResponse, error)
}

var _ gentic.ToolCallingLLM = (*MockToolCallingLLM)(nil)

// ChatWithTools delegates to ChatWithToolsFunc, or returns a stop response if unset.
func (m *MockToolCallingLLM) ChatWithTools(ctx context.Context, model string, messages []gentic.ToolMessage, tools []gentic.ToolDefinition) (*gentic.ToolCallingResponse, error) {
	if m != nil && m.ChatWithToolsFunc != nil {
		return m.ChatWithToolsFunc(ctx, model, messages, tools)
	}
	return &gentic.ToolCallingResponse{
		Message:      gentic.ToolMessage{Role: "assistant", Content: ""},
		FinishReason: "stop",
	}, nil
}

// ReplyToolCalls returns a [MockToolCallingLLM] that returns the given tool calls.
func ReplyToolCalls(toolCalls []gentic.ToolCall, err error) *MockToolCallingLLM {
	return &MockToolCallingLLM{
		ChatWithToolsFunc: func(context.Context, string, []gentic.ToolMessage, []gentic.ToolDefinition) (*gentic.ToolCallingResponse, error) {
			if err != nil {
				return nil, err
			}
			return &gentic.ToolCallingResponse{
				Message: gentic.ToolMessage{
					Role:      "assistant",
					ToolCalls: toolCalls,
				},
				FinishReason: "tool_calls",
			}, nil
		},
	}
}

// ReplyText returns a [MockToolCallingLLM] that returns the given text and stops.
func ReplyText(text string, err error) *MockToolCallingLLM {
	return &MockToolCallingLLM{
		ChatWithToolsFunc: func(context.Context, string, []gentic.ToolMessage, []gentic.ToolDefinition) (*gentic.ToolCallingResponse, error) {
			if err != nil {
				return nil, err
			}
			return &gentic.ToolCallingResponse{
				Message: gentic.ToolMessage{
					Role:    "assistant",
					Content: text,
				},
				FinishReason: "stop",
			}, nil
		},
	}
}
