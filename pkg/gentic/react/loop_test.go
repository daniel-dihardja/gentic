package react

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
)

// MockToolCallingLLM implements gentic.ToolCallingLLM for tests
type MockToolCallingLLM struct {
	ChatWithToolsFunc func(ctx context.Context, model string, messages []gentic.ToolMessage, tools []gentic.ToolDefinition) (*gentic.ToolCallingResponse, error)
}

var _ gentic.ToolCallingLLM = (*MockToolCallingLLM)(nil)

func (m *MockToolCallingLLM) ChatWithTools(ctx context.Context, model string, messages []gentic.ToolMessage, tools []gentic.ToolDefinition) (*gentic.ToolCallingResponse, error) {
	if m.ChatWithToolsFunc != nil {
		return m.ChatWithToolsFunc(ctx, model, messages, tools)
	}
	return &gentic.ToolCallingResponse{
		Message:      gentic.ToolMessage{Role: "assistant", Content: ""},
		FinishReason: "stop",
	}, nil
}

// TestNativeLoop_StopOnFirstTurn: model answers directly without tools
func TestNativeLoop_StopOnFirstTurn(t *testing.T) {
	mockLLM := &MockToolCallingLLM{
		ChatWithToolsFunc: func(ctx context.Context, model string, messages []gentic.ToolMessage, tools []gentic.ToolDefinition) (*gentic.ToolCallingResponse, error) {
			return &gentic.ToolCallingResponse{
				Message: gentic.ToolMessage{
					Role:    "assistant",
					Content: "Here is your answer directly.",
				},
				FinishReason: "stop",
			}, nil
		},
	}

	state := &gentic.State{Input: "What is 2+2?"}
	actor := NewReactActor(
		WithToolCallingLLM(mockLLM),
		WithMaxSteps(5),
	)
	flow := actor.Resolve(context.Background(), state)
	err := flow.Run(context.Background(), state)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state.Output != "Here is your answer directly." {
		t.Errorf("expected direct answer, got: %s", state.Output)
	}
	if len(state.Observations) != 0 {
		t.Errorf("expected no observations, got %d", len(state.Observations))
	}
	if len(state.Thoughts) < 1 {
		t.Errorf("expected at least 1 thought, got %d", len(state.Thoughts))
	}
}

// TestNativeLoop_SingleToolCallThenStop: fetch tool → final answer
func TestNativeLoop_SingleToolCallThenStop(t *testing.T) {
	callCount := 0
	mockLLM := &MockToolCallingLLM{
		ChatWithToolsFunc: func(ctx context.Context, model string, messages []gentic.ToolMessage, tools []gentic.ToolDefinition) (*gentic.ToolCallingResponse, error) {
			callCount++
			if callCount == 1 {
				// First turn: call fetch_location_profile
				return &gentic.ToolCallingResponse{
					Message: gentic.ToolMessage{
						Role: "assistant",
						ToolCalls: []gentic.ToolCall{
							{
								ID:   "call_1",
								Type: "function",
								Function: gentic.ToolCallFunction{
									Name:      "fetch_location_profile",
									Arguments: "{}",
								},
							},
						},
					},
					FinishReason: "tool_calls",
				}, nil
			}
			// Second turn: answer after tool result
			return &gentic.ToolCallingResponse{
				Message: gentic.ToolMessage{
					Role:    "assistant",
					Content: "I found your profile with summary: Cafe La Mer",
				},
				FinishReason: "stop",
			}, nil
		},
	}

	// Stub tool
	stubTool := NewToolWithState(
		"fetch_location_profile",
		"Fetch profile",
		json.RawMessage(`{"type":"object"}`),
		func(ctx context.Context, state *gentic.State, input json.RawMessage) (json.RawMessage, error) {
			return json.RawMessage(`{"exists":true,"summary":"Cafe La Mer"}`), nil
		},
	)

	state := &gentic.State{Input: "Show me my profile"}
	actor := NewReactActor(
		WithToolCallingLLM(mockLLM),
		WithMaxSteps(5),
		WithTools(stubTool),
	)
	flow := actor.Resolve(context.Background(), state)
	err := flow.Run(context.Background(), state)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state.Output != "I found your profile with summary: Cafe La Mer" {
		t.Errorf("expected final answer, got: %s", state.Output)
	}
	if len(state.Observations) != 1 {
		t.Errorf("expected 1 observation, got %d", len(state.Observations))
	}
	if state.Observations[0].TaskID != "fetch_location_profile" {
		t.Errorf("expected observation from fetch_location_profile, got: %s", state.Observations[0].TaskID)
	}
}

// TestNativeLoop_ToolErrorFeedback: tool error is sent back to model
func TestNativeLoop_ToolErrorFeedback(t *testing.T) {
	callCount := 0
	mockLLM := &MockToolCallingLLM{
		ChatWithToolsFunc: func(ctx context.Context, model string, messages []gentic.ToolMessage, tools []gentic.ToolDefinition) (*gentic.ToolCallingResponse, error) {
			callCount++
			if callCount == 1 {
				return &gentic.ToolCallingResponse{
					Message: gentic.ToolMessage{
						Role: "assistant",
						ToolCalls: []gentic.ToolCall{
							{
								ID:   "call_1",
								Type: "function",
								Function: gentic.ToolCallFunction{
									Name:      "failing_tool",
									Arguments: "{}",
								},
							},
						},
					},
					FinishReason: "tool_calls",
				}, nil
			}
			// After error, model stops
			return &gentic.ToolCallingResponse{
				Message: gentic.ToolMessage{
					Role:    "assistant",
					Content: "I encountered an error",
				},
				FinishReason: "stop",
			}, nil
		},
	}

	// Tool that fails
	failingTool := NewToolWithState(
		"failing_tool",
		"This tool fails",
		json.RawMessage(`{"type":"object"}`),
		func(ctx context.Context, state *gentic.State, input json.RawMessage) (json.RawMessage, error) {
			return nil, errors.New("tool execution failed")
		},
	)

	state := &gentic.State{Input: "Try to use failing tool"}
	actor := NewReactActor(
		WithToolCallingLLM(mockLLM),
		WithMaxSteps(5),
		WithTools(failingTool),
	)
	flow := actor.Resolve(context.Background(), state)
	err := flow.Run(context.Background(), state)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Verify error was recorded in observations
	if len(state.Observations) != 1 {
		t.Errorf("expected 1 observation with error, got %d", len(state.Observations))
	}
}

// TestNativeLoop_MaxStepsReached: loop terminates after max steps
func TestNativeLoop_MaxStepsReached(t *testing.T) {
	mockLLM := &MockToolCallingLLM{
		ChatWithToolsFunc: func(ctx context.Context, model string, messages []gentic.ToolMessage, tools []gentic.ToolDefinition) (*gentic.ToolCallingResponse, error) {
			// Always return tool calls, never stop
			return &gentic.ToolCallingResponse{
				Message: gentic.ToolMessage{
					Role: "assistant",
					ToolCalls: []gentic.ToolCall{
						{
							ID:   "call_x",
							Type: "function",
							Function: gentic.ToolCallFunction{
								Name:      "fetch_location_profile",
								Arguments: "{}",
							},
						},
					},
				},
				FinishReason: "tool_calls",
			}, nil
		},
	}

	stubTool := NewToolWithState(
		"fetch_location_profile",
		"Fetch profile",
		json.RawMessage(`{"type":"object"}`),
		func(ctx context.Context, state *gentic.State, input json.RawMessage) (json.RawMessage, error) {
			return json.RawMessage(`{"exists":true}`), nil
		},
	)

	state := &gentic.State{Input: "Test max steps"}
	actor := NewReactActor(
		WithToolCallingLLM(mockLLM),
		WithMaxSteps(3),
		WithTools(stubTool),
	)
	flow := actor.Resolve(context.Background(), state)
	err := flow.Run(context.Background(), state)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state.Output == "" {
		t.Errorf("expected fallback output after max steps")
	}
	if len(state.Thoughts) < 3 {
		t.Errorf("expected at least 3 thoughts (one per step), got %d", len(state.Thoughts))
	}
}

// TestNativeLoop_MultipleToolCallsInOneTurn: parallel tool calls
func TestNativeLoop_MultipleToolCallsInOneTurn(t *testing.T) {
	mockLLM := &MockToolCallingLLM{
		ChatWithToolsFunc: func(ctx context.Context, model string, messages []gentic.ToolMessage, tools []gentic.ToolDefinition) (*gentic.ToolCallingResponse, error) {
			// Check if we've received tool results (messages should have tool role messages)
			hasToolResults := false
			for _, msg := range messages {
				if msg.Role == "tool" {
					hasToolResults = true
					break
				}
			}
			if !hasToolResults {
				// First turn: call multiple tools
				return &gentic.ToolCallingResponse{
					Message: gentic.ToolMessage{
						Role: "assistant",
						ToolCalls: []gentic.ToolCall{
							{
								ID:   "call_1",
								Type: "function",
								Function: gentic.ToolCallFunction{
									Name:      "fetch_location_profile",
									Arguments: "{}",
								},
							},
							{
								ID:   "call_2",
								Type: "function",
								Function: gentic.ToolCallFunction{
									Name:      "fetch_location_profile",
									Arguments: "{}",
								},
							},
						},
					},
					FinishReason: "tool_calls",
				}, nil
			}
			// After tool results: answer
			return &gentic.ToolCallingResponse{
				Message: gentic.ToolMessage{
					Role:    "assistant",
					Content: "Got both results",
				},
				FinishReason: "stop",
			}, nil
		},
	}

	stubTool := NewToolWithState(
		"fetch_location_profile",
		"Fetch profile",
		json.RawMessage(`{"type":"object"}`),
		func(ctx context.Context, state *gentic.State, input json.RawMessage) (json.RawMessage, error) {
			return json.RawMessage(`{"exists":true}`), nil
		},
	)

	state := &gentic.State{Input: "Test parallel calls"}
	actor := NewReactActor(
		WithToolCallingLLM(mockLLM),
		WithMaxSteps(5),
		WithTools(stubTool),
	)
	flow := actor.Resolve(context.Background(), state)
	err := flow.Run(context.Background(), state)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state.Output != "Got both results" {
		t.Errorf("expected final answer, got: %s", state.Output)
	}
	if len(state.Observations) != 2 {
		t.Errorf("expected 2 observations (parallel calls), got %d", len(state.Observations))
	}
}

// TestNativeLoop_StateObservationsCompat: state.Thoughts and state.Observations populated correctly
func TestNativeLoop_StateObservationsCompat(t *testing.T) {
	callCount := 0
	mockLLM := &MockToolCallingLLM{
		ChatWithToolsFunc: func(ctx context.Context, model string, messages []gentic.ToolMessage, tools []gentic.ToolDefinition) (*gentic.ToolCallingResponse, error) {
			callCount++
			if callCount == 1 {
				return &gentic.ToolCallingResponse{
					Message: gentic.ToolMessage{
						Role: "assistant",
						ToolCalls: []gentic.ToolCall{
							{
								ID:   "call_1",
								Type: "function",
								Function: gentic.ToolCallFunction{
									Name:      "fetch_location_profile",
									Arguments: "{}",
								},
							},
						},
					},
					FinishReason: "tool_calls",
				}, nil
			}
			if callCount == 2 {
				return &gentic.ToolCallingResponse{
					Message: gentic.ToolMessage{
						Role: "assistant",
						ToolCalls: []gentic.ToolCall{
							{
								ID:   "call_2",
								Type: "function",
								Function: gentic.ToolCallFunction{
									Name:      "update_location_profile",
									Arguments: `{"summary":"Updated"}`,
								},
							},
						},
					},
					FinishReason: "tool_calls",
				}, nil
			}
			return &gentic.ToolCallingResponse{
				Message: gentic.ToolMessage{
					Role:    "assistant",
					Content: "Done updating",
				},
				FinishReason: "stop",
			}, nil
		},
	}

	fetchTool := NewToolWithState(
		"fetch_location_profile",
		"Fetch profile",
		json.RawMessage(`{"type":"object"}`),
		func(ctx context.Context, state *gentic.State, input json.RawMessage) (json.RawMessage, error) {
			return json.RawMessage(`{"exists":true,"summary":"Original"}`), nil
		},
	)
	updateTool := NewToolWithState(
		"update_location_profile",
		"Update profile",
		json.RawMessage(`{"type":"object"}`),
		func(ctx context.Context, state *gentic.State, input json.RawMessage) (json.RawMessage, error) {
			return json.RawMessage(`{"updated":true}`), nil
		},
	)

	state := &gentic.State{Input: "Update my profile"}
	actor := NewReactActor(
		WithToolCallingLLM(mockLLM),
		WithMaxSteps(10),
		WithTools(fetchTool, updateTool),
	)
	flow := actor.Resolve(context.Background(), state)
	err := flow.Run(context.Background(), state)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(state.Thoughts) < 3 {
		t.Errorf("expected at least 3 thoughts, got %d", len(state.Thoughts))
	}
	if len(state.Observations) != 2 {
		t.Errorf("expected 2 observations, got %d", len(state.Observations))
	}
	if state.Output != "Done updating" {
		t.Errorf("expected final answer, got: %s", state.Output)
	}
}
